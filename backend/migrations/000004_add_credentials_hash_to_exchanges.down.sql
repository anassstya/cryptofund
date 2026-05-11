DROP INDEX IF EXISTS idx_exchanges_user_name_credentials_hash;

ALTER TABLE exchanges
DROP COLUMN IF EXISTS credentials_hash;