-- Денний лічильник звернень до чату
-- Один рядок на користувача на добу; інкремент атомарний (UPSERT … RETURNING).
CREATE TABLE chat_usage (
    user_id BIGINT  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    day     DATE    NOT NULL DEFAULT current_date,
    count   INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, day)
);
