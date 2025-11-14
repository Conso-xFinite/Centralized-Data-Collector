package binance_define

type RedisChannelArg struct {
	IsSubscribe          bool   `json:"isSubscribe,omitempty"` // 发布端不用管, 订阅端加入到pending队列时写入此值
	Channel              string `json:"channel"`               // "price", "trades", "dex-token-candle1s"
	ChainIndex           string `json:"chainIndex"`
	TokenContractAddress string `json:"tokenContractAddress"`
	TokenContractId      string `json:"tokenContractId"`
}
