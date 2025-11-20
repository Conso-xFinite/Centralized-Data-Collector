package model

import "time"

type MarketDataFill struct {
	ID             string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Event          string    `db:"event" json:"event"`
	EventStartTime int64     `db:"event_start_time" json:"event_start_time"`
	EventEndTime   int64     `db:"event_end_time" json:"event_end_time"`
	Symbol         string    `db:"symbol" json:"symbol"`
	IsClosed       bool      `db:"is_closed" json:"is_closed"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

func (MarketDataFill) TableName() string {
	return "market_data_fill"
}
