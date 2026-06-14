-- Логи звернень до ШІ-агента. Відповідь моделі (інтент, дії, метадані)
-- зберігається у JSONB; GIN-індекс прискорює пошук за вмістом.
CREATE TABLE ai_chat_logs (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message    TEXT NOT NULL,
    intent     VARCHAR(30),
    response   JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_ai_chat_logs_response_gin ON ai_chat_logs USING GIN (response);
CREATE INDEX idx_ai_chat_logs_user ON ai_chat_logs (user_id, created_at DESC);
