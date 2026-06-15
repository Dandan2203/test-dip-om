"""Тести детектора аномалій (z-score) у agents/auditor.audit().

_explain() ходить у LLM по пояснення — монкіпатчимо його, щоб тестувати лише
детерміновану статистику (поріг > середнє + 2σ) та зіставлення причин за id.
"""
import agents.auditor as auditor
from agents.auditor import audit, _DEFAULT_REASON
from models import AuditRequest, TransactionItem


def _tx(id_, amount, category="Їжа", ttype="expense"):
    return TransactionItem(
        id=id_,
        category_id=1,
        category_name=category,
        type=ttype,
        amount=amount,
        description=f"tx{id_}",
        transaction_date="2026-06-01",
    )


def test_flags_clear_outlier(monkeypatch):
    monkeypatch.setattr(auditor, "_explain", lambda anomalies: {})
    # Дев'ять звичних витрат по 100 + одна на 1000 (поза порогом середнє+2σ).
    txs = [_tx(i, 100) for i in range(1, 10)] + [_tx(99, 1000)]
    res = audit(AuditRequest(user_id=1, transactions=txs))

    ids = [a.transaction_id for a in res.anomalies]
    assert ids == [99]
    assert res.anomalies[0].amount == 1000


def test_no_anomaly_when_values_uniform(monkeypatch):
    monkeypatch.setattr(auditor, "_explain", lambda anomalies: {})
    txs = [_tx(i, 100) for i in range(1, 8)]
    res = audit(AuditRequest(user_id=1, transactions=txs))
    assert res.anomalies == []


def test_category_with_few_points_skipped(monkeypatch):
    monkeypatch.setattr(auditor, "_explain", lambda anomalies: {})
    # Лише 2 транзакції в категорії (< 3) — статистика не рахується, аномалій немає.
    txs = [_tx(1, 100), _tx(2, 100000)]
    res = audit(AuditRequest(user_id=1, transactions=txs))
    assert res.anomalies == []


def test_income_transactions_ignored(monkeypatch):
    monkeypatch.setattr(auditor, "_explain", lambda anomalies: {})
    txs = [_tx(i, 100) for i in range(1, 6)] + [_tx(50, 99999, ttype="income")]
    res = audit(AuditRequest(user_id=1, transactions=txs))
    assert res.anomalies == []


def test_reason_matched_by_transaction_id(monkeypatch):
    # Пояснення прив'язується до конкретного id, а не до порядку рядків.
    monkeypatch.setattr(auditor, "_explain", lambda anomalies: {99: "Підозріла велика покупка"})
    txs = [_tx(i, 100) for i in range(1, 10)] + [_tx(99, 1000)]
    res = audit(AuditRequest(user_id=1, transactions=txs))
    assert res.anomalies[0].reason == "Підозріла велика покупка"


def test_default_reason_when_explain_empty(monkeypatch):
    monkeypatch.setattr(auditor, "_explain", lambda anomalies: {})
    txs = [_tx(i, 100) for i in range(1, 10)] + [_tx(99, 1000)]
    res = audit(AuditRequest(user_id=1, transactions=txs))
    assert res.anomalies[0].reason == _DEFAULT_REASON


def test_separate_categories_independent(monkeypatch):
    monkeypatch.setattr(auditor, "_explain", lambda anomalies: {})
    # Викид у «Їжа» не впливає на «Транспорт»; кожна категорія рахується окремо.
    food = [_tx(i, 100, "Їжа") for i in range(1, 10)] + [_tx(99, 1000, "Їжа")]
    transport = [_tx(200 + i, 50, "Транспорт") for i in range(1, 6)]
    res = audit(AuditRequest(user_id=1, transactions=food + transport))
    ids = {a.transaction_id for a in res.anomalies}
    assert ids == {99}
