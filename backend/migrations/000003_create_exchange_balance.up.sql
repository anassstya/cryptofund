CREATE TABLE IF NOT EXISTS exchange_balance (
    id              UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    exchange_id     UUID            UNIQUE NOT NULL REFERENCES exchanges(id) ON DELETE CASCADE,

    total_balance   NUMERIC(20, 8)  NOT NULL DEFAULT 0,
    change_percent  NUMERIC(10, 4)  NOT NULL DEFAULT 0,

    source          VARCHAR(20)     NOT NULL DEFAULT 'mock',
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),

    CHECK (total_balance >= 0),
    CHECK (source IN ('mock', 'live'))
    );


CREATE INDEX IF NOT EXISTS idx_exchange_balance_updated_at
    ON exchange_balance(exchange_id, updated_at DESC);