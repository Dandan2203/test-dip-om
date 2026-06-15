"""Action-інструменти: модель ними лише ЗАХОПЛЮЄ намір.
Виконання робить бекенд після підтвердження користувача (human-in-the-loop)."""
from datetime import date

ACTION_TOOLS = [
    {
        "name": "create_transaction",
        "description": "Додати/записати/внести нову транзакцію (дохід або витрату).",
        "input_schema": {
            "type": "object",
            "properties": {
                "amount": {"type": "number"},
                "transaction_type": {"type": "string", "enum": ["income", "expense"]},
                "description": {"type": "string"},
                "category_name": {"type": "string"},
                "transaction_date": {"type": "string", "description": "YYYY-MM-DD"},
            },
            "required": ["amount", "transaction_type"],
        },
    },
    {
        "name": "delete_transaction",
        "description": "Видалити наявну транзакцію за її ID.",
        "input_schema": {
            "type": "object",
            "properties": {"transaction_id": {"type": "integer"}},
            "required": ["transaction_id"],
        },
    },
    {
        "name": "create_goal",
        "description": "Створити нову фінансову ціль.",
        "input_schema": {
            "type": "object",
            "properties": {
                "title": {"type": "string"},
                "target_amount": {"type": "number"},
                "deadline": {"type": "string", "description": "YYYY-MM-DD"},
            },
            "required": ["title", "target_amount"],
        },
    },
    {
        "name": "contribute_goal",
        "description": "Поповнити/відкласти/додати кошти до НАЯВНОЇ цілі.",
        "input_schema": {
            "type": "object",
            "properties": {
                "amount": {"type": "number"},
                "title": {"type": "string", "description": "Назва цілі"},
                "goal_id": {"type": "integer"},
            },
            "required": ["amount"],
        },
    },
    {
        "name": "withdraw_goal",
        "description": "Зняти/забрати/відняти кошти з НАЯВНОЇ цілі.",
        "input_schema": {
            "type": "object",
            "properties": {
                "amount": {"type": "number"},
                "title": {"type": "string", "description": "Назва цілі"},
                "goal_id": {"type": "integer"},
            },
            "required": ["amount"],
        },
    },
    {
        "name": "delete_goal",
        "description": "Видалити наявну фінансову ціль.",
        "input_schema": {
            "type": "object",
            "properties": {
                "goal_id": {"type": "integer"},
                "title": {"type": "string"},
            },
        },
    },
]

ACTION_TOOL_NAMES = {t["name"] for t in ACTION_TOOLS}


def to_action_data(name: str, args: dict) -> dict:
    """tool_use → словник під models.ActionData (виконує бекенд після підтвердження)."""
    return {
        "action_type": name,
        "amount": args.get("amount"),
        "transaction_type": args.get("transaction_type"),
        "description": args.get("description"),
        "category_name": args.get("category_name"),
        "transaction_date": args.get("transaction_date") or date.today().isoformat(),
        "transaction_id": args.get("transaction_id"),
        "title": args.get("title"),
        "target_amount": args.get("target_amount"),
        "deadline": args.get("deadline"),
        "goal_id": args.get("goal_id"),
    }
