ALTER TABLE exchange_balance
DROP CONSTRAINT IF EXISTS exchange_balance_assets_count_non_negative;

ALTER TABLE exchange_balance
DROP COLUMN IF EXISTS assets_count;