CREATE TABLE IF NOT EXISTS exchanges (
    id              UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID            NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name            VARCHAR(30)     NOT NULL,
    api_key         TEXT            NOT NULL,
    api_secret      TEXT            NOT NULL,
    created_at      TIMESTAMPTZ         DEFAULT NOW(),

    CHECK (name IN ('Mexc', 'Binance', 'Bybit', 'Gate', 'Bitget'))
);

CREATE INDEX IF NOT EXISTS idx_exchanges_user_id
ON exchanges(user_id);
