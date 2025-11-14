package model

import "time"

// CREATE TABLE market_agg_trade (
//     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
//     event VARCHAR(20) NOT NULL,
//     event_time BIGINT,
//     symbol VARCHAR(20) NOT NULL,
//     agg_id BIGINT,
//     price NUMERIC(36,18),
//     quantity NUMERIC(36,18),
//     first_id BIGINT,
//     last_id BIGINT,
//     timestamp BIGINT,
//     buyer_maker BOOLEAN,
//     ignore BOOLEAN,
//     created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
// );

// -- 可选索引
// CREATE INDEX idx_binance_agg_trade_symbol_time ON binance_agg_trade(symbol, timestamp);
// CREATE INDEX idx_binance_agg_trade_event_time ON binance_agg_trade(event_time);

// BinanceAggTrade 对应 Binance WebSocket 聚合交易推送
type BinanceAggTrade struct {
	ID string `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	// 数据库主键，UUID，自动生成

	Event string `gorm:"type:varchar(20);not null;index" json:"event"`
	// 事件类型，例如 "aggTrade"，对应 JSON 中 "e"

	EventTime int64 `gorm:"type:bigint;index" json:"event_time"`
	// 事件触发时间，毫秒级时间戳，对应 JSON 中 "E"

	Symbol string `gorm:"type:varchar(20);not null;index" json:"symbol"`
	// 交易对，例如 BTCUSDT，对应 JSON 中 "s"

	AggID int64 `gorm:"type:bigint" json:"agg_id"`
	// 聚合交易ID，对应 JSON 中 "a"

	Price float64 `gorm:"type:numeric(36,18)" json:"price"`
	// 成交价格，对应 JSON 中 "p"

	Quantity float64 `gorm:"type:numeric(36,18)" json:"quantity"`
	// 成交数量，对应 JSON 中 "q"

	FirstID int64 `gorm:"type:bigint" json:"first_id"`
	// 本聚合交易包含的第一笔交易ID，对应 JSON 中 "f"

	LastID int64 `gorm:"type:bigint" json:"last_id"`
	// 本聚合交易包含的最后一笔交易ID，对应 JSON 中 "l"

	Timestamp int64 `gorm:"type:bigint;index" json:"timestamp"`
	// 交易发生时间，毫秒级时间戳，对应 JSON 中 "T"

	BuyerMaker bool `gorm:"type:boolean" json:"buyer_maker"`
	// 是否是买方为主动成交者，对应 JSON 中 "m"

	Ignore bool `gorm:"type:boolean" json:"ignore"`
	// 忽略字段，对应 JSON 中 "M"

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	// 记录创建时间，自动生成
}

func (BinanceAggTrade) TableName() string {
	return "market_agg_trade"
}
