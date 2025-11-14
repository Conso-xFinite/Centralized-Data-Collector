package model

// import "github.com/shopspring/decimal"

// \d price_history
//                                     Partitioned table "public.price_history"
//       Column      |           Type           | Collation | Nullable |                  Default
// ------------------+--------------------------+-----------+----------+-------------------------------------------
//  id               | bigint                   |           | not null | nextval('price_history_id_seq'::regclass)
//  token_address    | character varying(64)    |           | not null |
//  chain_id         | integer                  |           | not null |
//  price            | numeric(36,18)           |           | not null |
//  volume_24h       | numeric(36,18)           |           |          |
//  market_cap       | numeric(36,18)           |           |          |
//  price_change_24h | numeric(10,4)            |           |          |
//  timestamp        | timestamp with time zone |           | not null |
//  metadata         | jsonb                    |           |          | '{}'::jsonb
//  source           | character varying(50)    |           | not null | 'okx'::character varying
//  created_at       | timestamp with time zone |           |          | now()
// Partition key: RANGE ("timestamp")
// Indexes:
//     "price_history_pkey" PRIMARY KEY, btree (id, "timestamp")
// Number of partitions: 0

type BinancePrice struct {
	// ID              int64           `gorm:"type:bigint;default:nextval('price_history_id_seq'::regclass);primaryKey" json:"id"`
	// ChainIndex      int             `gorm:"type:int;not null;index" json:"chain_index"`
	// TokenContractID string          `gorm:"type:uuid;not null;index" json:"token_contract_id"`
	// Price           decimal.Decimal `gorm:"type:decimal(36,18);not null" json:"price"`
	// Timestamp       int64           `gorm:"type:timestamptz;not null;index" json:"timestamp"`
	// Source          string          `gorm:"type:varchar(50);not null;default:'okx'" json:"source"`
	// IsFirst         bool            `gorm:"type:boolean;default:false;index" json:"is_first"` // 是否为本次连接的第一条消息, rest http 补拉数据使用的
}

func (BinancePrice) TableName() string {
	return "price_history"
}
