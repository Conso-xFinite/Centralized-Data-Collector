package model

import "time"

// CREATE TABLE market_kline_1m (
//     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
//     pair VARCHAR(20) NOT NULL,
//     start_time BIGINT NOT NULL,
//     end_time BIGINT NOT NULL,
//     interval VARCHAR(10) NOT NULL,
//     first_trade_id BIGINT,
//     last_trade_id BIGINT,
//     open_price NUMERIC(36,18),
//     close_price NUMERIC(36,18),
//     high_price NUMERIC(36,18),
//     low_price NUMERIC(36,18),
//     volume NUMERIC(36,18),
//     trade_count INT,
//     is_closed BOOLEAN DEFAULT false,
//     quote_volume NUMERIC(36,18),
//     buy_volume NUMERIC(36,18),
//     buy_quote_volume NUMERIC(36,18),
//     block VARCHAR(50),
//     event_time BIGINT,
//     created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
// );

// -- 可选索引
// CREATE INDEX idx_kline_pair_start ON kline_1m(pair, start_time);
// CREATE INDEX idx_kline_end ON kline_1m(end_time);

// Kline1m 对应 Binance 1分钟K线数据
type Kline1m struct {
	ID string `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	// 数据库主键，UUID，自动生成

	Pair string `gorm:"type:varchar(20);not null;index" json:"pair"`
	// 交易对，例如 BTCUSDT，对应 JSON 中的 "s"

	StartTime int64 `gorm:"type:bigint;not null;index" json:"start_time"`
	// K线的起始时间，毫秒级时间戳，对应 JSON 中 "k.t"

	EndTime int64 `gorm:"type:bigint;not null" json:"end_time"`
	// K线的结束时间，毫秒级时间戳，对应 JSON 中 "k.T"

	Interval string `gorm:"type:varchar(10);not null" json:"interval"`
	// K线时间间隔，例如 "1m"，对应 JSON 中 "k.i"

	FirstTradeID int64 `gorm:"type:bigint" json:"first_trade_id"`
	// 本K线期间的第一笔交易 ID，对应 JSON 中 "k.f"

	LastTradeID int64 `gorm:"type:bigint" json:"last_trade_id"`
	// 本K线期间的最后一笔交易 ID，对应 JSON 中 "k.L"

	OpenPrice float64 `gorm:"type:numeric(36,18)" json:"open_price"`
	// 本K线的开盘价，对应 JSON 中 "k.o"

	ClosePrice float64 `gorm:"type:numeric(36,18)" json:"close_price"`
	// 本K线的收盘价，对应 JSON 中 "k.c"

	HighPrice float64 `gorm:"type:numeric(36,18)" json:"high_price"`
	// 本K线的最高价，对应 JSON 中 "k.h"

	LowPrice float64 `gorm:"type:numeric(36,18)" json:"low_price"`
	// 本K线的最低价，对应 JSON 中 "k.l"

	Volume float64 `gorm:"type:numeric(36,18)" json:"volume"`
	// 本K线的成交量，对应 JSON 中 "k.v"

	TradeCount int `gorm:"type:int" json:"trade_count"`
	// 本K线期间的成交笔数，对应 JSON 中 "k.n"

	IsClosed bool `gorm:"type:boolean;default:false" json:"is_closed"`
	// K线是否收盘，即本K线是否已结束，对应 JSON 中 "k.x"

	QuoteVolume float64 `gorm:"type:numeric(36,18)" json:"quote_volume"`
	// 本K线的成交额（以计价币种计算），对应 JSON 中 "k.q"

	BuyVolume float64 `gorm:"type:numeric(36,18)" json:"buy_volume"`
	// 主动买入的成交量，对应 JSON 中 "k.V"

	BuyQuoteVolume float64 `gorm:"type:numeric(36,18)" json:"buy_quote_volume"`
	// 主动买入的成交额，对应 JSON 中 "k.Q"

	Block string `gorm:"type:varchar(50)" json:"block"`
	// Binance WebSocket 原始字段 "k.B"，一般可忽略，保留做参考

	EventTime int64 `gorm:"type:bigint" json:"event_time"`
	// 事件触发时间，对应 JSON 中 "E"

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	// 记录创建时间，自动生成
}

func (Kline1m) TableName() string {
	return "market_kline_1m"
}
