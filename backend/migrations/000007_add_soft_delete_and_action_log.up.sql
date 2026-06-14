-- М'яке видалення: записи позначаються deleted_at, а не стираються фізично.
-- Це дає змогу відкочувати дії агента («Відмінити останню дію»).
ALTER TABLE transactions ADD COLUMN deleted_at TIMESTAMPTZ;
ALTER TABLE goals ADD COLUMN deleted_at TIMESTAMPTZ;

-- Журнал дій для функції undo. payload зберігає знімок сутності у JSONB.
CREATE TABLE action_log (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action_type VARCHAR(20) NOT NULL,   -- create | delete
    entity_type VARCHAR(20) NOT NULL,   -- transaction | goal
    entity_id   BIGINT NOT NULL,
    payload     JSONB NOT NULL DEFAULT '{}'::jsonb,
    undone      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_action_log_user ON action_log (user_id, created_at DESC);
