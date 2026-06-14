-- Таблиця категорій фінансових операцій.
-- Системна категорія має user_id = NULL і доступна всім користувачам.
CREATE TABLE categories (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT REFERENCES users(id) ON DELETE CASCADE,
    name       VARCHAR(100) NOT NULL,
    type       VARCHAR(20)  NOT NULL CHECK (type IN ('income', 'expense')),
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (user_id, name, type)
);

-- Системні категорії (user_id = NULL) — спільні для всіх користувачів.
INSERT INTO categories (user_id, name, type) VALUES
    (NULL, 'Зарплата', 'income'),
    (NULL, 'Підробіток', 'income'),
    (NULL, 'Подарунки', 'income'),
    (NULL, 'Інвестиції', 'income'),
    (NULL, 'Інше', 'income'),
    (NULL, 'Їжа та напої', 'expense'),
    (NULL, 'Транспорт', 'expense'),
    (NULL, 'Житло', 'expense'),
    (NULL, 'Комунальні послуги', 'expense'),
    (NULL, 'Здоров''я', 'expense'),
    (NULL, 'Розваги', 'expense'),
    (NULL, 'Одяг', 'expense'),
    (NULL, 'Освіта', 'expense'),
    (NULL, 'Підписки', 'expense'),
    (NULL, 'Інше', 'expense');
