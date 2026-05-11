CREATE TABLE IF NOT EXISTS exchange_balance_history (
    id              UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
    exchange_id     UUID             NOT NULL REFERENCES exchanges(id) ON DELETE CASCADE,
    total_balance   DOUBLE PRECISION NOT NULL,
    created_at      TIMESTAMPTZ      NOT NULL DEFAULT NOW(),

    CHECK (total_balance >= 0)
    );

CREATE INDEX IF NOT EXISTS idx_exchange_balance_history_exchange_id_created_at
    ON exchange_balance_history(exchange_id, created_at DESC);