ALTER TABLE exchange_balance
    ADD COLUMN IF NOT EXISTS pairs JSONB NOT NULL DEFAULT '[]'::jsonb;