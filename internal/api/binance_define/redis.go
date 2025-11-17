package binance_define

type RedisChannelArg struct {
	TypeSubscribe string `json:"typeSubscribe"` // 发布端不用管, 订阅端加入到pending队列时写入此值
	Channel       string `json:"channel"`       // "aggTrade", "kline_1m", "miniTicker",
	TokenPair     string `json:"TokenPair"`
}
