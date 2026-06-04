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
    txs = [_tx(1, 100), _tx(2, 100000)]
    res = audit(AuditRequest(user_id=1, transactions=txs))
    assert res.anomalies == []


def test_outlier_compared_without_including_itself(monkeypatch):
    monkeypatch.setattr(auditor, "_explain", lambda anomalies: {})
    txs = [_tx(1, 280), _tx(2, 320), _tx(3, 410), _tx(4, 2600)]

    res = audit(AuditRequest(user_id=1, transactions=txs))

    assert [a.transaction_id for a in res.anomalies] == [4]


def test_uncategorized_outlier_uses_categorized_history(monkeypatch):
    monkeypatch.setattr(auditor, "_explain", lambda anomalies: {})
    normal = [_tx(1, 100), _tx(2, 150), _tx(3, 200), _tx(4, 250)]
    outlier = _tx(5, 200000, category=None)
    outlier.category_id = None

    res = audit(AuditRequest(user_id=1, transactions=[*normal, outlier]))

    assert [a.transaction_id for a in res.anomalies] == [5]


def test_income_transactions_ignored(monkeypatch):
    monkeypatch.setattr(auditor, "_explain", lambda anomalies: {})
    txs = [_tx(i, 100) for i in range(1, 6)] + [_tx(50, 99999, ttype="income")]
    res = audit(AuditRequest(user_id=1, transactions=txs))
    assert res.anomalies == []


def test_reason_matched_by_transaction_id(monkeypatch):
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
    food = [_tx(i, 100, "Їжа") for i in range(1, 10)] + [_tx(99, 1000, "Їжа")]
    transport = [_tx(200 + i, 50, "Транспорт") for i in range(1, 6)]
    res = audit(AuditRequest(user_id=1, transactions=food + transport))
    ids = {a.transaction_id for a in res.anomalies}
    assert ids == {99}
