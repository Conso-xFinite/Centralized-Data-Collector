DROP TABLE market_agg_trade;
DROP TABLE market_kline_1m;

CREATE TABLE market_agg_trade (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event VARCHAR(20) NOT NULL,
    event_time BIGINT,
    symbol VARCHAR(20) NOT NULL,
    agg_id BIGINT,
    price NUMERIC(36,18),
    quantity NUMERIC(36,18),
    first_id BIGINT,
    last_id BIGINT,
    timestamp BIGINT,
    buyer_maker BOOLEAN,
    ignore BOOLEAN,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);



CREATE TABLE market_kline_1m (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pair VARCHAR(20) NOT NULL,
    start_time BIGINT NOT NULL,
    end_time BIGINT NOT NULL,
    interval VARCHAR(10) NOT NULL,
    first_trade_id BIGINT,
    last_trade_id BIGINT,
    open_price NUMERIC(36,18),
    close_price NUMERIC(36,18),
    high_price NUMERIC(36,18),
    low_price NUMERIC(36,18),
    volume NUMERIC(36,18),
    trade_count INT,
    is_closed BOOLEAN DEFAULT false,
    quote_volume NUMERIC(36,18),
    buy_volume NUMERIC(36,18),
    buy_quote_volume NUMERIC(36,18),
    block VARCHAR(50),
    event_time BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);