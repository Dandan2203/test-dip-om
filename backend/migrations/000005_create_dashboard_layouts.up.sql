-- Розкладки інтерактивного дашборда. Зберігаються на акаунт у форматі JSONB.
-- name: 'overview' (Огляд) або 'goals' (Цілі).
CREATE TABLE dashboard_layouts (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       VARCHAR(50) NOT NULL,
    layout     JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, name)
);
