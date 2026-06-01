from types import SimpleNamespace

from agents import chat
from agents.chat import _needs_action_repair
from models import ChatRequest


def test_repair_when_model_claims_direct_action_was_done():
    assert _needs_action_repair(
        "додай витрату 400000, кредит",
        [],
        'Додав витрату 400 000₴ в категорію "кредит".',
    )


def test_repair_after_action_clarification():
    history = [
        {"role": "user", "content": "додай 150000 просто"},
        {"role": "assistant", "content": "Це дохід чи витрата?"},
    ]

    assert _needs_action_repair("дохід", history, "Додав дохід 150 000₴.")


def test_no_repair_for_clarifying_response():
    assert not _needs_action_repair(
        "додай 150000 просто",
        [],
        "Це дохід чи витрата?",
    )


def test_run_chat_repairs_false_success_with_action(monkeypatch):
    responses = [
        SimpleNamespace(content=[SimpleNamespace(type="text", text="Додав дохід 150 000₴.")]),
        SimpleNamespace(
            content=[
                SimpleNamespace(
                    type="tool_use",
                    name="create_transaction",
                    input={"amount": 150000, "transaction_type": "income"},
                )
            ]
        ),
    ]
    calls = []

    class FakeMessages:
        def create(self, **kwargs):
            calls.append(kwargs)
            return responses.pop(0)

    monkeypatch.setattr(chat, "_client", SimpleNamespace(messages=FakeMessages()))

    result = chat.run_chat(
        ChatRequest(
            message="додай дохід 150000",
            user_id=1,
            transactions=[],
            goals=[],
            categories=[],
        )
    )

    assert result.intent == "ACTION"
    assert result.action.action_type == "create_transaction"
    assert result.action.amount == 150000
    assert calls[1]["tool_choice"]["type"] == "any"
