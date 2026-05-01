CREATE TABLE IF NOT EXISTS exchanges (
    id              UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID            NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name            VARCHAR(30)     NOT NULL,
    api_key         VARCHAR(255)    NOT NULL,
    api_secret      VARCHAR(255)    NOT NULL,

    CHECK (name IN ('mexc', 'binance', 'bybit', 'okx', 'gate', 'kucoin'))
);

CREATE INDEX idx_exchanges_user_id
ON exchanges(user_id)

