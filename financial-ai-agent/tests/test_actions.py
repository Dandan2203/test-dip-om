"""Тести action-інструментів: коректність конвертера наміру та повнота набору.

Окремі contribute_goal/withdraw_goal — ключове рішення проти хардкод-списку слів,
тож перевіряємо, що вони існують і не злиті в один інструмент.
"""
from datetime import date

from agents.actions import (
    ACTION_TOOLS,
    ACTION_TOOL_NAMES,
    to_action_data,
)


def test_action_tool_set_complete():
    assert ACTION_TOOL_NAMES == {
        "create_transaction",
        "delete_transaction",
        "create_goal",
        "contribute_goal",
        "withdraw_goal",
        "delete_goal",
    }


def test_contribute_and_withdraw_are_separate_tools():
    # Саме розділення прибрало потребу в хардкод-списку слів «зняти/поповнити».
    assert "contribute_goal" in ACTION_TOOL_NAMES
    assert "withdraw_goal" in ACTION_TOOL_NAMES
    assert "contribute_goal" != "withdraw_goal"


def test_each_tool_has_schema():
    for t in ACTION_TOOLS:
        assert t["name"]
        assert t["description"]
        assert t["input_schema"]["type"] == "object"


def test_to_action_data_create_transaction():
    out = to_action_data("create_transaction", {
        "amount": 500,
        "transaction_type": "expense",
        "description": "кава",
        "category_name": "Їжа",
    })
    assert out["action_type"] == "create_transaction"
    assert out["amount"] == 500
    assert out["transaction_type"] == "expense"
    assert out["description"] == "кава"
    assert out["category_name"] == "Їжа"


def test_to_action_data_defaults_transaction_date_to_today():
    out = to_action_data("create_transaction", {"amount": 100, "transaction_type": "income"})
    assert out["transaction_date"] == date.today().isoformat()


def test_to_action_data_keeps_explicit_date():
    out = to_action_data("create_transaction", {
        "amount": 100, "transaction_type": "income", "transaction_date": "2026-01-01",
    })
    assert out["transaction_date"] == "2026-01-01"


def test_to_action_data_contribute_goal():
    out = to_action_data("contribute_goal", {"amount": 2000, "title": "Відпустка"})
    assert out["action_type"] == "contribute_goal"
    assert out["amount"] == 2000
    assert out["title"] == "Відпустка"


def test_to_action_data_withdraw_goal():
    out = to_action_data("withdraw_goal", {"amount": 500, "goal_id": 3})
    assert out["action_type"] == "withdraw_goal"
    assert out["amount"] == 500
    assert out["goal_id"] == 3


def test_to_action_data_delete_transaction():
    out = to_action_data("delete_transaction", {"transaction_id": 7})
    assert out["action_type"] == "delete_transaction"
    assert out["transaction_id"] == 7


def test_to_action_data_missing_fields_are_none():
    out = to_action_data("delete_goal", {"goal_id": 1})
    assert out["amount"] is None
    assert out["transaction_id"] is None
    assert out["target_amount"] is None
