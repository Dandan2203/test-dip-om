-- user_id NULL = системна категорія.
INSERT INTO categories (user_id, name, type) VALUES
    (NULL, 'Перекази', 'income'),
    (NULL, 'Перекази', 'expense')
ON CONFLICT (user_id, name, type) DO NOTHING;
