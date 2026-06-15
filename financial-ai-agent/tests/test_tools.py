"""Тести read-інструментів чат-агента — чистих функцій над контекстом запиту.

Дати рахуються відносно date.today(), щоб транзакції завжди потрапляли у вікно періоду.
"""
from datetime import date, timedelta

from agents.tools import (
    get_balance,
    get_top_expenses,
    simulate_purchase,
    get_goals_progress,
    dispatch_read_tool,
    READ_TOOL_NAMES,
)


def _d(days_ago: int) -> str:
    return (date.today() - timedelta(days=days_ago)).isoformat()


def _tx(amount, ttype, category=None, days_ago=1):
    return {
        "amount": amount,
        "type": ttype,
        "category_name": category,
        "transaction_date": _d(days_ago),
    }


# --- get_balance ---

def test_get_balance_income_minus_expense():
    txs = [
        _tx(1000, "income"),
        _tx(300, "expense"),
        _tx(200, "expense"),
    ]
    out = get_balance(txs, 30)
    assert "1000.00₴" in out
    assert "500.00₴" in out  # витрати
    assert "500.00₴" in out  # баланс = 1000 - 500


def test_get_balance_excludes_out_of_period():
    txs = [
        _tx(1000, "income", days_ago=1),
        _tx(5000, "income", days_ago=90),  # поза вікном 30 днів
    ]
    out = get_balance(txs, 30)
    assert "1000.00₴" in out
    assert "5000" not in out


def test_get_balance_empty():
    out = get_balance([], 30)
    assert "0.00₴" in out


# --- get_top_expenses ---

def test_get_top_expenses_sorted_desc():
    txs = [
        _tx(100, "expense", "Їжа"),
        _tx(400, "expense", "Їжа"),
        _tx(700, "expense", "Транспорт"),
        _tx(50, "income", "Зарплата"),  # дохід не враховується
    ]
    out = get_top_expenses(txs, 5, 30)
    # Транспорт (700) має йти перед Їжа (500)
    assert out.index("Транспорт") < out.index("Їжа")
    assert "Зарплата" not in out


def test_get_top_expenses_respects_limit():
    txs = [_tx(10 * i, "expense", f"Кат{i}") for i in range(1, 6)]
    out = get_top_expenses(txs, 2, 30)
    assert out.count("- ") == 2


def test_get_top_expenses_uncategorized_label():
    txs = [_tx(100, "expense", None)]
    out = get_top_expenses(txs, 5, 30)
    assert "Без категорії" in out


def test_get_top_expenses_empty():
    assert "немає" in get_top_expenses([], 5, 30)


# --- simulate_purchase ---

def test_simulate_purchase_positive_remainder():
    txs = [_tx(1000, "income"), _tx(200, "expense")]
    out = simulate_purchase(txs, 300)
    assert "позитивний" in out
    assert "500.00₴" in out  # 800 поточний - 300 = 500


def test_simulate_purchase_negative_remainder():
    txs = [_tx(500, "income")]
    out = simulate_purchase(txs, 800)
    assert "від'ємний" in out
    assert "-300.00₴" in out


# --- get_goals_progress ---

def test_get_goals_progress_percentage():
    goals = [{"title": "Авто", "current_amount": 25000, "target_amount": 100000}]
    out = get_goals_progress(goals)
    assert "Авто" in out
    assert "25%" in out


def test_get_goals_progress_zero_target_no_crash():
    goals = [{"title": "X", "current_amount": 0, "target_amount": 0}]
    out = get_goals_progress(goals)
    assert "0%" in out  # ділення на нуль захищене


def test_get_goals_progress_empty():
    assert "не встановлено" in get_goals_progress([])


# --- dispatch_read_tool ---

def test_dispatch_routes_each_tool():
    txs = [_tx(1000, "income"), _tx(300, "expense", "Їжа")]
    goals = [{"title": "Ціль", "current_amount": 10, "target_amount": 100}]

    assert "баланс" in dispatch_read_tool("get_balance", {}, txs, goals)
    assert "Топ" in dispatch_read_tool("get_top_expenses", {}, txs, goals)
    assert "залишиться" in dispatch_read_tool("simulate_purchase", {"amount": 100}, txs, goals)
    assert "Ціль" in dispatch_read_tool("get_goals_progress", {}, txs, goals)


def test_dispatch_defaults_when_args_missing():
    txs = [_tx(1000, "income")]
    # period_days/limit відсутні — мають підставитися дефолти без помилок
    assert "30 днів" in dispatch_read_tool("get_balance", {}, txs, [])


def test_dispatch_unknown_tool():
    assert "Невідомий" in dispatch_read_tool("no_such_tool", {}, [], [])


def test_read_tool_names_match_schemas():
    assert READ_TOOL_NAMES == {
        "get_balance",
        "get_top_expenses",
        "simulate_purchase",
        "get_goals_progress",
    }
