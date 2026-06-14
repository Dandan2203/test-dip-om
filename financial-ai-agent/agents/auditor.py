import statistics
import anthropic
from models import AuditRequest, AuditResponse, AnomalyItem
from config import ANTHROPIC_API_KEY, ANTHROPIC_MODEL

_client = anthropic.Anthropic(api_key=ANTHROPIC_API_KEY)


def audit(req: AuditRequest) -> AuditResponse:
    expense_txs = [t for t in req.transactions if t.type == "expense"]

    # Групуємо суми витрат по категорії
    by_category: dict[str, list] = {}
    for t in expense_txs:
        cat = t.category_name or "Без категорії"
        by_category.setdefault(cat, []).append(t)

    raw_anomalies = []
    for cat, txs in by_category.items():
        amounts = [t.amount for t in txs]
        if len(amounts) < 3:
            continue
        mean = statistics.mean(amounts)
        std = statistics.stdev(amounts)
        threshold = mean + 2 * std
        for t in txs:
            if t.amount > threshold:
                raw_anomalies.append({
                    "transaction_id": t.id,
                    "description": t.description,
                    "category_name": cat,
                    "amount": t.amount,
                    "mean": mean,
                    "threshold": threshold,
                })

    if not raw_anomalies:
        return AuditResponse(anomalies=[])

    # Одним запитом до Claude пояснюємо всі аномалії
    items_text = "\n".join(
        f"{i + 1}. Категорія «{a['category_name']}», сума {a['amount']:.0f}₴, "
        f"середня {a['mean']:.0f}₴, опис: {a['description'] or 'відсутній'}"
        for i, a in enumerate(raw_anomalies)
    )
    prompt = (
        f"Наступні транзакції значно перевищують середнє по категорії (> середнє + 2σ):\n\n"
        f"{items_text}\n\n"
        f"Для кожної надай коротке пояснення (1 речення) українською чому варто звернути увагу. "
        f"Формат відповіді: '1. <пояснення>' на кожному рядку."
    )

    message = _client.messages.create(
        model=ANTHROPIC_MODEL,
        max_tokens=512,
        messages=[{"role": "user", "content": prompt}],
    )
    lines = message.content[0].text.strip().splitlines()
    reasons: list[str] = []
    for line in lines:
        stripped = line.strip()
        if stripped and stripped[0].isdigit():
            # Видаляємо "1. " на початку
            reason = stripped.split(".", 1)[-1].strip() if "." in stripped else stripped
            reasons.append(reason)

    # Доповнюємо до кількості аномалій якщо рядків менше
    while len(reasons) < len(raw_anomalies):
        reasons.append("Витрата перевищує звичний рівень для цієї категорії.")

    anomalies = [
        AnomalyItem(
            transaction_id=a["transaction_id"],
            description=a["description"],
            category_name=a["category_name"],
            amount=a["amount"],
            mean=a["mean"],
            reason=reasons[i],
        )
        for i, a in enumerate(raw_anomalies)
    ]
    return AuditResponse(anomalies=anomalies)
