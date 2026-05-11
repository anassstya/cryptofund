ALTER TABLE exchanges
    ADD COLUMN IF NOT EXISTS credentials_hash TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_exchanges_user_name_credentials_hash
    ON exchanges(user_id, name, credentials_hash)
    WHERE credentials_hash IS NOT NULL;