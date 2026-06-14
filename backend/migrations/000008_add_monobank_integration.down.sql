DROP INDEX IF EXISTS idx_transactions_external;
ALTER TABLE transactions DROP COLUMN IF EXISTS original_amount;
ALTER TABLE transactions DROP COLUMN IF EXISTS currency_code;
ALTER TABLE transactions DROP COLUMN IF EXISTS external_id;
ALTER TABLE transactions DROP COLUMN IF EXISTS source;
DROP TABLE IF EXISTS mono_connections;
