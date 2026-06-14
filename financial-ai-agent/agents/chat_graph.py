import json
from datetime import date as date_class
from typing import TypedDict
from langchain_anthropic import ChatAnthropic
from langchain_core.messages import HumanMessage, SystemMessage
from langgraph.graph import StateGraph, END
from langgraph.prebuilt import create_react_agent

from config import ANTHROPIC_API_KEY, ANTHROPIC_MODEL
from agents.tools import build_financial_tools

ROUTER_SYSTEM = """Ти класифікатор запитів фінансового асистента.
Визнач інтент повідомлення і відповідай ТІЛЬКИ одним словом:
- FINANCIAL — питання про витрати, доходи, баланс, категорії
- SCENARIO  — моделювання «що, якщо», вплив покупок на бюджет
- ADVICE    — прохання про фінансові поради
- ACTION    — команда виконати дію: додати/записати/внести транзакцію, видалити транзакцію, створити/видалити ціль
- GENERAL   — загальна розмова, не пов'язана з фінансами"""

ADVISOR_SYSTEM = """Ти фінансовий асистент. Відповідай українською.
Використовуй доступні інструменти щоб отримати реальні дані користувача.
Давай конкретні відповіді з числами. Не вигадуй дані — лише те, що є в інструментах."""

ACTION_SYSTEM = """Ти парсер дій фінансового асистента FinAgent. Відповідай ТІЛЬКИ валідним JSON без markdown.

Проаналізуй повідомлення та визнач дію:
- "create_transaction" — додати/записати/внести транзакцію
- "delete_transaction" — видалити транзакцію (потрібен ID)
- "create_goal" — створити/додати фінансову ціль
- "contribute_goal" — поповнити/відкласти/додати кошти до НАЯВНОЇ цілі (потрібні сума і назва/ID цілі)
- "withdraw_goal" — зняти/відняти/прибрати/забрати кошти з НАЯВНОЇ цілі (потрібні сума і назва/ID цілі)
- "delete_goal" — видалити ціль (потрібен ID)
- null — якщо дія незрозуміла

Приклади contribute_goal: «1500 на відпустку», «відклади 500 на авто», «додай 200 до цілі ремонт».
Приклади withdraw_goal: «зніми 500 з відпустки», «забери 300 з цілі авто», «відніми 100 від ремонту».
Для поповнення/зняття цілі: amount=<сума>, title=<назва цілі з повідомлення>, goal_id якщо знайшов у контексті.

Всі поля обов'язкові у відповіді (null якщо не застосовно):
{{
  "action_type": "create_transaction"|"delete_transaction"|"create_goal"|"contribute_goal"|"withdraw_goal"|"delete_goal"|null,
  "amount": <число або null>,
  "transaction_type": "income"|"expense"|null,
  "description": "<рядок або null>",
  "category_name": "<назва категорії або null>",
  "transaction_date": "<YYYY-MM-DD або null>",
  "transaction_id": <ціле або null>,
  "title": "<назва цілі або null>",
  "target_amount": <число або null>,
  "deadline": "<YYYY-MM-DD або null>",
  "goal_id": <ціле або null>,
  "response": "<коротке підтвердження українською>"
}}

ПРАВИЛО (дуже важливо): якщо в повідомленні є слова «зняти/зніми/забери/забрати/відняти/відніми/прибрати/витягни/мінус» щодо цілі — це ЗАВЖДИ withdraw_goal, НІКОЛИ не contribute_goal. Слова «додай/поповни/відклади/на/до цілі» — це contribute_goal.

Сьогодні: {today}. Відповідай ТІЛЬКИ JSON."""


class ChatState(TypedDict):
    message: str
    transactions: list[dict]
    goals: list[dict]
    categories: list[dict]
    intent: str
    response: str
    action: dict | None


def router_node(state: ChatState) -> dict:
    llm = ChatAnthropic(model=ANTHROPIC_MODEL, api_key=ANTHROPIC_API_KEY, temperature=0)
    result = llm.invoke([
        SystemMessage(content=ROUTER_SYSTEM),
        HumanMessage(content=state["message"]),
    ])
    intent = result.content.strip().upper()
    if intent not in ("FINANCIAL", "SCENARIO", "ADVICE", "ACTION"):
        intent = "GENERAL"

    if intent == "GENERAL":
        reply = ChatAnthropic(model=ANTHROPIC_MODEL, api_key=ANTHROPIC_API_KEY, temperature=0.5).invoke([
            SystemMessage(content="Ти дружній фінансовий асистент FinAgent. Відповідай українською."),
            HumanMessage(content=state["message"]),
        ])
        return {"intent": intent, "response": reply.content, "action": None}

    return {"intent": intent}


def advisor_node(state: ChatState) -> dict:
    tools = build_financial_tools(state["transactions"], state["goals"])
    llm = ChatAnthropic(model=ANTHROPIC_MODEL, api_key=ANTHROPIC_API_KEY, temperature=0.3)
    agent = create_react_agent(llm, tools, prompt=ADVISOR_SYSTEM)

    result = agent.invoke({"messages": [HumanMessage(content=state["message"])]})
    last = result["messages"][-1]
    return {"response": last.content, "action": None}


def action_node(state: ChatState) -> dict:
    today = date_class.today().isoformat()
    system_prompt = ACTION_SYSTEM.format(today=today)

    context_parts: list[str] = []
    if state["transactions"]:
        recent = state["transactions"][-5:]
        context_parts.append("Останні транзакції: " + "; ".join(
            f"ID={t['id']} {t['description']} {t['amount']}₴" for t in recent
        ))
    if state["goals"]:
        context_parts.append("Цілі: " + "; ".join(
            f"ID={g['id']} {g['title']}" for g in state["goals"]
        ))

    user_content = state["message"]
    if context_parts:
        user_content += "\n\nКонтекст:\n" + "\n".join(context_parts)

    llm = ChatAnthropic(model=ANTHROPIC_MODEL, api_key=ANTHROPIC_API_KEY, temperature=0)
    result = llm.invoke([
        SystemMessage(content=system_prompt),
        HumanMessage(content=user_content),
    ])

    try:
        content = result.content.strip()
        if content.startswith("```"):
            parts = content.split("```")
            content = parts[1]
            if content.startswith("json"):
                content = content[4:]

        data = json.loads(content)
        response_text = data.get("response", "Дію виконано.")

        action = None
        if data.get("action_type"):
            action = {
                "action_type": data.get("action_type"),
                "amount": data.get("amount"),
                "transaction_type": data.get("transaction_type"),
                "description": data.get("description"),
                "category_name": data.get("category_name"),
                "transaction_date": data.get("transaction_date") or today,
                "transaction_id": data.get("transaction_id"),
                "title": data.get("title"),
                "target_amount": data.get("target_amount"),
                "deadline": data.get("deadline"),
                "goal_id": data.get("goal_id"),
            }

            # Детермінований запобіжник: модель плутає зняття з поповненням.
            withdraw_words = (
                "зніми", "знім", "зняти", "знять", "забери", "забрати",
                "відніми", "відняти", "прибери", "прибрати", "витягни", "витягти", "мінус",
            )
            msg_lower = state["message"].lower()
            if action["action_type"] == "contribute_goal" and any(w in msg_lower for w in withdraw_words):
                action["action_type"] = "withdraw_goal"

        return {"intent": "ACTION", "response": response_text, "action": action}
    except (json.JSONDecodeError, KeyError, IndexError):
        return {"intent": "ACTION", "response": "Не зміг розпізнати дію. Спробуйте ще раз.", "action": None}


def _route(state: ChatState) -> str:
    if state["intent"] == "ACTION":
        return "action"
    if state["intent"] in ("FINANCIAL", "SCENARIO", "ADVICE"):
        return "advisor"
    return END


def build_chat_graph():
    graph = StateGraph(ChatState)
    graph.add_node("router", router_node)
    graph.add_node("advisor", advisor_node)
    graph.add_node("action", action_node)
    graph.set_entry_point("router")
    graph.add_conditional_edges(
        "router",
        _route,
        {"advisor": "advisor", "action": "action", END: END},
    )
    graph.add_edge("advisor", END)
    graph.add_edge("action", END)
    return graph.compile()


chat_graph = build_chat_graph()
