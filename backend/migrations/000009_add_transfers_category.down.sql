DELETE FROM categories WHERE user_id IS NULL AND name = 'Перекази' AND type IN ('income', 'expense');
