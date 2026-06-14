-- Підключення Monobank: персональний токен зберігається ЗАШИФРОВАНИМ (AES-GCM).
CREATE TABLE mono_connections (
    user_id         BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    token_encrypted TEXT NOT NULL,
    connected_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_import_at  TIMESTAMPTZ
);

-- Поля джерела та валюти для імпортованих транзакцій.
-- amount завжди в гривні (для єдиної аналітики), original_* — оригінальна валюта.
ALTER TABLE transactions ADD COLUMN source          VARCHAR(20)  NOT NULL DEFAULT 'manual';
ALTER TABLE transactions ADD COLUMN external_id     VARCHAR(64);
ALTER TABLE transactions ADD COLUMN currency_code   SMALLINT     NOT NULL DEFAULT 980; -- ISO 4217 (980 = UAH)
ALTER TABLE transactions ADD COLUMN original_amount NUMERIC(14,2);

-- Дедуплікація імпорту: один зовнішній запис Mono на користувача.
CREATE UNIQUE INDEX idx_transactions_external
    ON transactions (user_id, external_id)
    WHERE external_id IS NOT NULL;
