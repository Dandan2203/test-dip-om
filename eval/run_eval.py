#!/usr/bin/env python3
"""Eval-харнес FinAgent: точність інтентів і парсингу дій.

Запуск:
    INTERNAL_API_TOKEN=<token> python eval/run_eval.py
    INTERNAL_API_TOKEN=<token> python eval/run_eval.py --url http://localhost:8000

Потребує: pip install requests
AI-сервіс має бути запущений (docker compose up ai АБО uvicorn main:app).
"""
import argparse
import json
import os
import sys
from pathlib import Path
from typing import Any

try:
    import requests
except ImportError:
    sys.exit("requests не встановлено: pip install requests")

FIXTURES = Path(__file__).parent / "fixtures"

# Фіксований контекст
SAMPLE_CONTEXT: dict[str, Any] = {
    "user_id": 1,
    "transactions": [
        {"id": 1, "category_id": 1, "category_name": "Їжа",       "type": "expense", "amount": 850.0,   "description": "Продукти",  "transaction_date": "2026-06-01"},
        {"id": 2, "category_id": 2, "category_name": "Транспорт", "type": "expense", "amount": 120.0,   "description": "Метро",     "transaction_date": "2026-06-02"},
        {"id": 3, "category_id": 3, "category_name": "Зарплата",  "type": "income",  "amount": 35000.0, "description": "Зарплата",  "transaction_date": "2026-06-01"},
        {"id": 4, "category_id": 1, "category_name": "Їжа",       "type": "expense", "amount": 420.0,   "description": "Кафе",      "transaction_date": "2026-06-03"},
        {"id": 5, "category_id": 4, "category_name": "Розваги",   "type": "expense", "amount": 650.0,   "description": "Кіно",      "transaction_date": "2026-06-04"},
        {"id": 6, "category_id": 5, "category_name": "Комунальні","type": "expense", "amount": 1800.0,  "description": "ЖКГ",       "transaction_date": "2026-06-05"},
        {"id": 7, "category_id": 1, "category_name": "Їжа",       "type": "expense", "amount": 310.0,   "description": "Доставка",  "transaction_date": "2026-06-06"},
    ],
    "goals": [
        {"id": 1, "title": "Відпустка", "target_amount": 20000.0, "current_amount": 5000.0,  "deadline": "2026-12-31"},
        {"id": 2, "title": "Машина",    "target_amount": 150000.0,"current_amount": 30000.0, "deadline": "2028-01-01"},
        {"id": 3, "title": "Ноутбук",   "target_amount": 30000.0, "current_amount": 12000.0, "deadline": "2026-09-01"},
    ],
    "categories": [
        {"id": 1, "name": "Їжа",        "type": "expense"},
        {"id": 2, "name": "Транспорт",  "type": "expense"},
        {"id": 3, "name": "Зарплата",   "type": "income"},
        {"id": 4, "name": "Розваги",    "type": "expense"},
        {"id": 5, "name": "Комунальні", "type": "expense"},
    ],
}


def call_chat(url: str, token: str, message: str) -> dict[str, Any]:
    payload = {"message": message, **SAMPLE_CONTEXT}
    r = requests.post(
        f"{url}/chat",
        json=payload,
        headers={"X-Internal-Token": token},
        timeout=60,
    )
    r.raise_for_status()
    return r.json()


def _fields_match(actual: dict, expected: dict) -> bool:
    for k, v in expected.items():
        a = actual.get(k)
        if isinstance(v, (int, float)):
            if a is None:
                return False
            try:
                if abs(float(a) - float(v)) > 0.5:
                    return False
            except (TypeError, ValueError):
                return False
        else:
            if str(a).lower().strip() != str(v).lower().strip():
                return False
    return True


def run_intent_eval(url: str, token: str) -> tuple[int, int]:
    cases: list[dict] = json.loads((FIXTURES / "intents.json").read_text(encoding="utf-8"))
    passed = 0
    print("\n── Тест інтентів ──────────────────────────────────────────────────")
    print(f"{'#':>3}  {'Очікувано':<12}  {'Отримано':<12}  OK  Повідомлення")
    print("─" * 72)
    for c in cases:
        try:
            resp = call_chat(url, token, c["message"])
            actual = resp.get("intent", "?")
            err = None
        except Exception as e:
            actual = "ERROR"
            err = str(e)
        ok = actual == c["expected_intent"]
        if ok:
            passed += 1
        mark = "✓" if ok else "✗"
        note = f"  [{err[:30]}]" if err else ""
        print(f"{c['id']:>3}  {c['expected_intent']:<12}  {actual:<12}  {mark}  {c['message'][:44]}{note}")
    total = len(cases)
    pct = 100 * passed // total
    print(f"\nІнтенти: {passed}/{total} ({pct} %)")
    return passed, total


def run_action_eval(url: str, token: str) -> tuple[int, int]:
    cases: list[dict] = json.loads((FIXTURES / "actions.json").read_text(encoding="utf-8"))
    passed = 0
    print("\n── Тест дій ───────────────────────────────────────────────────────")
    print(f"{'#':>3}  {'Очікувана дія':<22}  {'Отримана дія':<22}  OK  Поля")
    print("─" * 72)
    for c in cases:
        try:
            resp = call_chat(url, token, c["message"])
            action: dict = resp.get("action") or {}
            actual_type = action.get("action_type", "?")
            err = None
        except Exception as e:
            actual_type = "ERROR"
            action = {}
            err = str(e)
        type_ok = actual_type == c["expected_action_type"]
        fields_ok = _fields_match(action, c.get("expected_fields", {}))
        ok = type_ok and fields_ok
        if ok:
            passed += 1
        mark = "✓" if ok else "✗"
        fmark = "✓" if fields_ok else "✗"
        note = f"  [{err[:30]}]" if err else ""
        print(f"{c['id']:>3}  {c['expected_action_type']:<22}  {actual_type:<22}  {mark}  {fmark}{note}")
    total = len(cases)
    pct = 100 * passed // total
    print(f"\nДії: {passed}/{total} ({pct} %)")
    return passed, total


def main() -> None:
    parser = argparse.ArgumentParser(description="FinAgent eval harness")
    parser.add_argument("--url", default="http://localhost:8000", help="AI service base URL")
    args = parser.parse_args()

    token = os.environ.get("INTERNAL_API_TOKEN", "")
    if not token:
        sys.exit("Встановіть змінну: export INTERNAL_API_TOKEN=<token>")

    print(f"AI-сервіс: {args.url}")

    ip, it = run_intent_eval(args.url, token)
    ap, at = run_action_eval(args.url, token)

    total_p = ip + ap
    total_t = it + at
    pct = 100 * total_p // total_t
    print(f"\n══ Загалом: {total_p}/{total_t} ({pct} %) ══")
    sys.exit(0 if pct >= 80 else 1)


if __name__ == "__main__":
    main()
