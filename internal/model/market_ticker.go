package model

import (
	"time"
)

// Binance24hrMiniTickerModel 表示从 Binance WebSocket 接收到的 24hrMiniTicker 数据并用于存储到 Postgres。
type TickerModel struct {
	ID string `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	// 主键，UUID，默认使用 gen_random_uuid()

	Event string `gorm:"type:varchar(20);not null;index" json:"event"`
	// 事件类型，例如 "24hrMiniTicker"

	EventTime int64 `gorm:"type:bigint;index" json:"event_time"`
	// 事件发生时间，毫秒级时间戳 (E)

	Symbol string `gorm:"type:varchar(32);not null;index:idx_symbol_event_time" json:"symbol"`
	// 交易对，例如 "BTCUSDT"（存入时建议用大写或保持原样）

	Close string `gorm:"type:numeric(36,18)" json:"close"`
	// 最新成交价 (原始为字符串)，DB 列是 numeric(36,18)

	Open string `gorm:"type:numeric(36,18)" json:"open"`
	// 开盘价

	High string `gorm:"type:numeric(36,18)" json:"high"`
	// 最高价

	Low string `gorm:"type:numeric(36,18)" json:"low"`
	// 最低价

	Volume string `gorm:"type:numeric(36,18)" json:"volume"`
	// 24 小时成交量 (base asset volume)

	QuoteVolume string `gorm:"type:numeric(36,18)" json:"quote_volume"`
	// 24 小时成交额 (quote asset volume)

	CreatedAt time.Time `gorm:"type:timestamp with time zone;default:now()" json:"created_at"`
	// 记录插入时间（自动填充）
}

func (TickerModel) TableName() string {
	return "market_24h_ticker"
}
