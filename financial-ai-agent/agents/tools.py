"""Read-only фінансові інструменти чат-агента. Працюють над контекстом запиту."""
from datetime import date, timedelta


def _filter_period(txs: list[dict], days: int) -> list[dict]:
    cutoff = (date.today() - timedelta(days=days)).isoformat()
    return [t for t in txs if t.get("transaction_date", "") >= cutoff]


def get_balance(txs: list[dict], period_days: int = 30) -> str:
    filtered = _filter_period(txs, period_days)
    income = sum(t["amount"] for t in filtered if t["type"] == "income")
    expense = sum(t["amount"] for t in filtered if t["type"] == "expense")
    return (
        f"За останні {period_days} днів: доходи {income:.2f}₴, "
        f"витрати {expense:.2f}₴, баланс {income - expense:.2f}₴."
    )


def get_top_expenses(txs: list[dict], limit: int = 5, period_days: int = 30) -> str:
    filtered = _filter_period(txs, period_days)
    totals: dict[str, float] = {}
    for t in filtered:
        if t["type"] != "expense":
            continue
        cat = t.get("category_name") or "Без категорії"
        totals[cat] = totals.get(cat, 0) + t["amount"]
    sorted_cats = sorted(totals.items(), key=lambda x: x[1], reverse=True)[:limit]
    if not sorted_cats:
        return "Витрат за цей період немає."
    lines = "\n".join(f"- {cat}: {amt:.2f}₴" for cat, amt in sorted_cats)
    return f"Топ витрат за {period_days} днів:\n{lines}"


def simulate_purchase(txs: list[dict], amount: float) -> str:
    filtered = _filter_period(txs, 30)
    income = sum(t["amount"] for t in filtered if t["type"] == "income")
    expense = sum(t["amount"] for t in filtered if t["type"] == "expense")
    current = income - expense
    new_balance = current - amount
    status = "позитивний" if new_balance >= 0 else "від'ємний"
    return (
        f"Поточний баланс місяця: {current:.2f}₴. "
        f"Після покупки {amount:.2f}₴ залишиться {new_balance:.2f}₴ ({status})."
    )


def get_goals_progress(goals: list[dict]) -> str:
    if not goals:
        return "Фінансових цілей не встановлено."
    lines = []
    for g in goals:
        pct = (g["current_amount"] / g["target_amount"] * 100) if g["target_amount"] else 0
        lines.append(
            f"- {g['title']}: {g['current_amount']:.0f}₴ / {g['target_amount']:.0f}₴ ({pct:.0f}%)"
        )
    return "Цілі:\n" + "\n".join(lines)


READ_TOOLS = [
    {
        "name": "get_balance",
        "description": "Доходи, витрати й баланс користувача за останні N днів.",
        "input_schema": {
            "type": "object",
            "properties": {
                "period_days": {"type": "integer", "description": "Період у днях (типово 30)"}
            },
        },
    },
    {
        "name": "get_top_expenses",
        "description": "Топ категорій витрат за сумою за останні N днів.",
        "input_schema": {
            "type": "object",
            "properties": {
                "limit": {"type": "integer", "description": "Скільки категорій (типово 5)"},
                "period_days": {"type": "integer", "description": "Період у днях (типово 30)"},
            },
        },
    },
    {
        "name": "simulate_purchase",
        "description": "Сценарій «що, якщо»: як покупка на суму amount вплине на баланс поточного місяця.",
        "input_schema": {
            "type": "object",
            "properties": {"amount": {"type": "number", "description": "Сума покупки"}},
            "required": ["amount"],
        },
    },
    {
        "name": "get_goals_progress",
        "description": "Прогрес досягнення фінансових цілей користувача.",
        "input_schema": {"type": "object", "properties": {}},
    },
]

READ_TOOL_NAMES = {t["name"] for t in READ_TOOLS}


def dispatch_read_tool(name: str, args: dict, txs: list[dict], goals: list[dict]) -> str:
    if name == "get_balance":
        return get_balance(txs, int(args.get("period_days") or 30))
    if name == "get_top_expenses":
        return get_top_expenses(txs, int(args.get("limit") or 5), int(args.get("period_days") or 30))
    if name == "simulate_purchase":
        return simulate_purchase(txs, float(args.get("amount") or 0))
    if name == "get_goals_progress":
        return get_goals_progress(goals)
    return "Невідомий інструмент."
