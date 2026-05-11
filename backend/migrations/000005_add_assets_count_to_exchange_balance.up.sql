ALTER TABLE exchange_balance
    ADD COLUMN IF NOT EXISTS assets_count INTEGER NOT NULL DEFAULT 0;

ALTER TABLE exchange_balance
    ADD CONSTRAINT exchange_balance_assets_count_non_negative
        CHECK (assets_count >= 0);