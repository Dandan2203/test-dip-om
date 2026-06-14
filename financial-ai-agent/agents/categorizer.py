import json
import anthropic
from models import CategorizeRequest, CategorizeResponse
from config import ANTHROPIC_API_KEY, ANTHROPIC_MODEL

_client = anthropic.Anthropic(api_key=ANTHROPIC_API_KEY)


def categorize(req: CategorizeRequest) -> CategorizeResponse:
    categories_list = "\n".join(
        f"- id={c.id}, name={c.name}" for c in req.available_categories
    )

    prompt = f"""Визнач категорію для фінансової транзакції.

Опис транзакції: "{req.description}"
Тип: {req.transaction_type}

Доступні категорії:
{categories_list}

Відповідай ТІЛЬКИ JSON без пояснень:
{{"category_id": <число або null>, "confidence": <від 0 до 1>, "category_name": "<назва або null>"}}

Якщо жодна категорія не підходить — постав null."""

    message = _client.messages.create(
        model=ANTHROPIC_MODEL,
        max_tokens=128,
        messages=[{"role": "user", "content": prompt}],
    )

    raw = message.content[0].text.strip()
    # Витягуємо JSON з можливого markdown-блоку
    if "```" in raw:
        raw = raw.split("```")[1].replace("json", "").strip()

    try:
        data = json.loads(raw)
        return CategorizeResponse(
            category_id=data.get("category_id"),
            confidence=float(data.get("confidence", 0.0)),
            category_name=data.get("category_name"),
        )
    except (json.JSONDecodeError, KeyError, ValueError):
        return CategorizeResponse(category_id=None, confidence=0.0, category_name=None)
