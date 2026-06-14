CREATE TABLE transactions (
    id                BIGSERIAL    PRIMARY KEY,
    user_id           BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id       BIGINT       REFERENCES categories(id) ON DELETE SET NULL,
    type              VARCHAR(20)  NOT NULL CHECK (type IN ('income', 'expense')),
    amount            NUMERIC(12, 2) NOT NULL,
    description       TEXT         NOT NULL DEFAULT '',
    transaction_date  DATE         NOT NULL,
    is_ai_categorized BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_transactions_user_date ON transactions(user_id, transaction_date);
