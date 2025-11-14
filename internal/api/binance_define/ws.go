package binance_define

type Errcode int

const (
	Success Errcode = 0
)

// const (
// 	ChannelPrice    = "price"
// 	ChannelTrades   = "trades"
// 	ChannelCandle1s = "dex-token-candle1s"
// )

// const Source = "okx"

// type ChainIndexType int

// const (
// 	ChainIndex_EthereumMainnet ChainIndexType = 1
// 	ChainIndex_OPMainnet       ChainIndexType = 10
// 	ChainIndex_BNBMainnet      ChainIndexType = 56
// 	ChainIndex_Sonic           ChainIndexType = 146
// 	ChainIndex_UniChain        ChainIndexType = 130
// 	ChainIndex_XLayer          ChainIndexType = 196
// 	ChainIndex_Rolygon         ChainIndexType = 137
// 	ChainIndex_ArbitrumOne     ChainIndexType = 42161
// 	ChainIndex_AvalancheCChain ChainIndexType = 43114
// 	ChainIndex_zkSyncEra       ChainIndexType = 324
// 	ChainIndex_PolygonzkEVM    ChainIndexType = 1101
// 	ChainIndex_Base            ChainIndexType = 8453
// 	ChainIndex_Linea           ChainIndexType = 59144
// 	ChainIndex_FantomOpera     ChainIndexType = 250
// 	ChainIndex_Mantle          ChainIndexType = 5000
// 	ChainIndex_ConfluxeSpace   ChainIndexType = 1030
// 	ChainIndex_MetisAndromeda  ChainIndexType = 1088
// 	ChainIndex_MerlinChain     ChainIndexType = 4200
// 	ChainIndex_Blast           ChainIndexType = 81457
// 	ChainIndex_MantaPacific    ChainIndexType = 169
// 	ChainIndex_Scroll          ChainIndexType = 534352
// 	ChainIndex_CronosMainnet   ChainIndexType = 25
// 	ChainIndex_ZetaChain       ChainIndexType = 7000
// 	ChainIndex_Plasma          ChainIndexType = 9745
// 	ChainIndex_Tron            ChainIndexType = 195
// 	ChainIndex_Solana          ChainIndexType = 501
// 	ChainIndex_SUI             ChainIndexType = 784
// 	ChainIndex_Ton             ChainIndexType = 607
// )

// const (
// 	BaseURL     = "https://www.okx.com/api/v5"
// 	WSBaseURL   = "wss://wsdex.okx.com/ws/v5/dex"
// 	WSLoginRout = "/users/self/verify"
// )

// type WSLoginArg struct {
// 	ApiKey     string `json:"apiKey"`
// 	Passphrase string `json:"passphrase"`
// 	Timestamp  string `json:"timestamp"`
// 	Sign       string `json:"sign"`
// }

type WSSubscribeArg struct {
	// Op                   string `json:"op"` // 不序列化, 本地用来区分订阅和取消订阅的
	Channel              string `json:"channel"`
	ChainIndex           string `json:"chainIndex"`
	TokenContractAddress string `json:"tokenContractAddress"`
}

// 待删除
type WSReq struct {
	Op   string        `json:"op"`
	Args []interface{} `json:"args"`
}

// type WSLoginReq struct {
// 	Op   string        `json:"op"`
// 	Args []interface{} `json:"args"`
// }

// type WSChannelReq struct {
// 	Op   string        `json:"op"`
// 	Args WSSubscribeArg `json:"args"`
// }

// type WSResp struct {
// 	Event  string      `json:"event"`
// 	Code   string      `json:"code,omitempty"`
// 	Msg    string      `json:"msg,omitempty"`
// 	Arg    interface{} `json:"arg"`
// 	ConnId string      `json:"connId"`
// }

// const (
// 	ErrCode_Success                 string = "0"     // 0 Success
// 	ErrCode_InvalidTimestamp        string = "60004" // 60004 Invalid timestamp
// 	ErrCode_InvalidApiKey           string = "60005" // 60005 Invalid apiKey
// 	ErrCode_TimestampRequestExpired string = "60006" // 60006 Timestamp request expired
// 	ErrCode_InvalidSign             string = "60007" // 60007 Invalid sign
// 	ErrCode_PleaseLogin             string = "60011" // 60011 Please log in
// 	ErrCode_InvalidRequest          string = "60012" // 60012 Invalid request
// 	ErrCode_InvalidArgs             string = "60013" // 60013 Invalid args
// 	ErrCode_RequestsTooFrequent     string = "60014" // 60014 Requests too frequent
// 	ErrCode_WrongURLOrNotExist      string = "60018" // 60018 Wrong URL or {0} doesn't exist. Please use the correct URL, channel and parameters referring to API document
// 	ErrCode_WrongPassphase          string = "60024" // 60024 Wrong passphase
// 	ErrCode_OnlyWhitelistAllowed    string = "60029" // 60029 Only users who are in the whitelist are allowed to subscribe to this channel.
// 	ErrCode_NoMultipleLogins        string = "60031" // 60031 The WebSocket endpoint does not allow multiple or repeated logins.
// 	ErrCode_ApiKeyNotExist          string = "60032" // 60032 API key doesn't exist.
// 	ErrCode_LoginInternalError      string = "63999" // 63999 Login failed due to internal err. Please try again later.

// 	// notice codes
// 	ErrCode_ConnectionWillCloseForServiceUpgrade string = "64008" // 64008 The connection will soon be closed for service upgrade. Please reconnect.
// )

// type WSLoginResp struct {
// 	Event  string `json:"event"`
// 	Code   string `json:"code,omitempty"`
// 	Msg    string `json:"msg,omitempty"`
// 	ConnId string `json:"connId"`
// }

// func (wssr *WSLoginResp) IsSuccess() bool {
// 	return wssr.Code == "0" || wssr.Code == ""
// }

// type WSSubscribeResp struct {
// 	Event  string      `json:"event"`
// 	Arg    interface{} `json:"arg,omitempty"`
// 	Code   string      `json:"code,omitempty"`
// 	Msg    string      `json:"msg,omitempty"`
// 	ConnId string      `json:"connId"`
// }
// type WSPushedMessage struct {
// 	// price
// 	Arg  WSSubscribeArg  `json:"arg"`
// 	Data json.RawMessage `json:"data"`
// }

// 将订阅响应与数据抢着合并合并
// type WSSubscribeResp struct {
// 	Event  string          `json:"event,omitempty"`
// 	Arg    WSSubscribeArg  `json:"arg,omitempty"`
// 	Code   string          `json:"code,omitempty"`
// 	Msg    string          `json:"msg,omitempty"`
// 	ConnId string          `json:"connId,omitempty"`
// 	Data   json.RawMessage `json:"data,omitempty"`
// }

// func (wssr *WSSubscribeResp) IsSuccess() bool {
// 	return wssr.Code == "0" || wssr.Code == ""
// }

type SubscribeMsg struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
	ID     int      `json:"id"`
}

type WSSinglePushMsg struct {
	IsFirst   bool        `json:"isFirst,omitempty"` // 是否为本次连接后接收到的第一条推送数据
	EventType string      `json:"event,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

// type WSChangedTokenInfo struct {
// 	Amount       string `json:"amount"`
// 	TokenAddress string `json:"tokenAddress"`
// 	TokenSymbol  string `json:"tokenSymbol"`
// }

// type WSPriceData struct {
// 	ChainIndex           string `json:"chainIndex"`
// 	TokenContractAddress string `json:"tokenContractAddress"`
// 	Price                string `json:"price"`
// 	Time                 string `json:"time"` // 毫秒级时间戳
// }

type BinanceAggTrade struct {
	Event      string `json:"e"`
	EventTime  int64  `json:"E"`
	Symbol     string `json:"s"`
	AggID      int64  `json:"a"`
	Price      string `json:"p"`
	Quantity   string `json:"q"`
	FirstID    int64  `json:"f"`
	LastID     int64  `json:"l"`
	Timestamp  int64  `json:"T"`
	BuyerMaker bool   `json:"m"`
	Ignore     bool   `json:"M"`
}

// KlineMessage 表示 Binance 推送的 kline 消息
type KlineMessage struct {
	E  string `json:"e"` // 事件类型，例如 "kline"
	Ev int64  `json:"E"` // 事件时间（毫秒时间戳）
	S  string `json:"s"` // 交易对，例如 "BNBBTC"
	K  Kline  `json:"k"` // 内嵌的 kline 数据
}

// Kline 表示单根 K 线的详细数据
type Kline struct {
	T    int64  `json:"t"` // 这根K线的起始时间
	TEnd int64  `json:"T"` // 这根K线的结束时间
	S    string `json:"s"` // 交易对
	I    string `json:"i"` // K线间隔，例如 "1m"
	F    int64  `json:"f"` // 第一笔成交ID
	L    int64  `json:"L"` // 末一笔成交ID
	O    string `json:"o"` // 开盘价
	C    string `json:"c"` // 收盘价
	H    string `json:"h"` // 最高价
	Lw   string `json:"l"` // 最低价
	V    string `json:"v"` // 成交量
	N    int64  `json:"n"` // 成交笔数
	X    bool   `json:"x"` // K线是否完结
	Q    string `json:"q"` // 成交额
	VB   string `json:"V"` // 主动买入成交量
	QB   string `json:"Q"` // 主动买入成交额
	B    string `json:"B"` // 忽略字段
}

type Binance24hrMiniTicker struct {
	Event       string `json:"e"` // 事件类型
	EventTime   int64  `json:"E"` // 事件时间
	Symbol      string `json:"s"` // 交易对
	Close       string `json:"c"` // 最新成交价
	Open        string `json:"o"` // 开盘价
	High        string `json:"h"` // 最高价
	Low         string `json:"l"` // 最低价
	Volume      string `json:"v"` // 24h 成交量
	QuoteVolume string `json:"q"` // 24h 成交额
}

// type WSTradeData struct {
// 	ChainIndex           string               `json:"chainIndex"`
// 	ChangedTokenInfos    []WSChangedTokenInfo `json:"changedTokenInfo"`
// 	TokenContractAddress string               `json:"tokenContractAddress"`
// 	DexName              string               `json:"dexName"`
// 	ID                   string               `json:"id"`
// 	IsFiltered           string               `json:"isFiltered"`
// 	PoolLogoUrl          string               `json:"poolLogoUrl"`
// 	Price                string               `json:"price"`
// 	Time                 string               `json:"time"` // 毫秒级时间戳
// 	TxHashUrl            string               `json:"txHashUrl"`
// 	Type                 string               `json:"type"` // buy or sell
// 	UserAddress          string               `json:"userAddress"`
// 	Volume               string               `json:"volume"` // 交易的美元价值
// }

// type WSCandleDataFieldType int

// const (
// 	WSCandleDataFieldTimestamp  WSCandleDataFieldType = iota // K线产生的毫秒级时间戳
// 	WSCandleDataFieldOpen                                    // 开盘价
// 	WSCandleDataFieldHigh                                    // 最高价
// 	WSCandleDataFieldLow                                     // 最低价
// 	WSCandleDataFieldClose                                   // 收盘价
// 	WSCandleDataFieldVolume                                  // 交易量
// 	WSCandleDataFieldVolumeUsd                               // 交易量的美元计价
// 	WSCandleDataFieldConfirm                                 // K线的状态, 0:代表完成, 1:代表未完成
// 	WSCandleDataFieldTotalCount                              // 总的Field数量
// )
