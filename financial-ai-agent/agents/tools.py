"""Фінансові інструменти для Advisor-агента. Всі дані беруться з контексту запиту."""
from datetime import date, timedelta
from langchain_core.tools import tool


def build_financial_tools(transactions: list[dict], goals: list[dict]):
    """Повертає список tools із захопленим фінансовим контекстом."""

    def _filter_period(txs: list[dict], days: int) -> list[dict]:
        cutoff = (date.today() - timedelta(days=days)).isoformat()
        return [t for t in txs if t.get("transaction_date", "") >= cutoff]

    @tool
    def get_balance(period_days: int = 30) -> str:
        """Повертає баланс (доходи - витрати) за вказаний період у днях."""
        filtered = _filter_period(transactions, period_days)
        income = sum(t["amount"] for t in filtered if t["type"] == "income")
        expense = sum(t["amount"] for t in filtered if t["type"] == "expense")
        return (
            f"За останні {period_days} днів: доходи {income:.2f}₴, "
            f"витрати {expense:.2f}₴, баланс {income - expense:.2f}₴."
        )

    @tool
    def get_top_expenses(limit: int = 5, period_days: int = 30) -> str:
        """Повертає топ категорій витрат за сумою."""
        filtered = _filter_period(transactions, period_days)
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

    @tool
    def simulate_purchase(amount: float) -> str:
        """Показує, як запланована покупка вплине на бюджет поточного місяця."""
        filtered = _filter_period(transactions, 30)
        income = sum(t["amount"] for t in filtered if t["type"] == "income")
        expense = sum(t["amount"] for t in filtered if t["type"] == "expense")
        current_balance = income - expense
        new_balance = current_balance - amount
        status = "позитивний" if new_balance >= 0 else "від'ємний"
        return (
            f"Поточний баланс місяця: {current_balance:.2f}₴. "
            f"Після покупки {amount:.2f}₴ залишиться {new_balance:.2f}₴ ({status})."
        )

    @tool
    def get_goals_progress() -> str:
        """Показує прогрес досягнення фінансових цілей."""
        if not goals:
            return "Фінансових цілей не встановлено."
        lines = []
        for g in goals:
            pct = (g["current_amount"] / g["target_amount"] * 100) if g["target_amount"] else 0
            lines.append(
                f"- {g['title']}: {g['current_amount']:.0f}₴ / {g['target_amount']:.0f}₴ ({pct:.0f}%)"
            )
        return "Цілі:\n" + "\n".join(lines)

    return [get_balance, get_top_expenses, simulate_purchase, get_goals_progress]
