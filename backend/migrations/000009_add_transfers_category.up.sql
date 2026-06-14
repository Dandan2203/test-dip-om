-- Системна категорія «Перекази» для доходів і витрат (user_id = NULL).
INSERT INTO categories (user_id, name, type) VALUES
    (NULL, 'Перекази', 'income'),
    (NULL, 'Перекази', 'expense')
ON CONFLICT (user_id, name, type) DO NOTHING;
