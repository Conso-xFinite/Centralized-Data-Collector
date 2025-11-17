package collector

import (
	"Centralized-Data-Collector/internal/api/binance_api"
	"Centralized-Data-Collector/internal/api/binance_define"
	"Centralized-Data-Collector/pkg/logger"
	"Centralized-Data-Collector/pkg/utils"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

// const OKX_SUBSCRIPTION_STREAM = "coinhub:subscription:okx:channel"
const BINANCE_SUBSCRIBED_CHANNELS = "data_collector:subscribed:binance:channel"

type DataCollector struct {
	apiClient   *binance_api.WSPool // 重连时, 抛弃
	redisHelper *utils.RedisHelper
	storer      *BinanceStorer
}

var dataCollectorInstance *DataCollector = nil

func DataCollectorInstance() *DataCollector {
	return dataCollectorInstance
}

func NewDataCollector(ctx context.Context, redisClient *redis.Client) (*DataCollector, error) {
	binanceConnPool := binance_api.NewWSPool(3)
	dataCollector := &DataCollector{
		apiClient:   binanceConnPool,
		redisHelper: utils.NewRedisHelper(redisClient),
		storer:      NewBinanceStorer(),
	}
	// logger.Debug("NewWSPool", dataCollector.apiClient.GetClientsLength())

	//测试写入订阅数据

	// user := User{Name: "Alice", Age: 20}
	// data, _ := json.Marshal(user)

	// r.HSet(ctx, "user:1001", "profile", data)

	// "etcusdt@aggTrade", "etcusdt@kline_1m", "etcusdt@miniTicker",

	// type RedisChannelArg struct {
	// TypeSubscribe string `json:"typeSubscribe"` // 发布端不用管, 订阅端加入到pending队列时写入此值
	// Channel       string `json:"channel"`       // "aggTrade", "kline_1m", "miniTicker",
	// TokenPair     string `json:"TokenPair"`
	// }

	// channelArgsJson, _ := json.MarshalIndent(redisOkxChannelArg, "", "  ")

	// dataCollector.redisHelper.HSet(ctx, BINANCE_SUBSCRIBED_CHANNELS, "btcusdt@aggTrade", {
	// 	TypeSubscribe: ""
	// 	Channel: ""     // "aggTrade", "kline_1m", "miniTicker",
	// 	TokenPair :" "    string `json:"TokenPair"`
	// })

	// err33 := dataCollector.redisHelper.HDel(ctx, BINANCE_SUBSCRIBED_CHANNELS,
	// 	"BTCUSDT@miniTicker",
	// 	"bnbusdt@aggTrade",
	// 	"BTCUSDT@kline_1m",
	// )
	// if err33 != nil {
	// 	logger.Debug("err33", err33)
	// }

	arg := &binance_define.RedisChannelArg{
		TypeSubscribe: "subscribe",
		Channel:       "aggTrade",
		TokenPair:     "btcusdt",
	}
	jsonBytes1, _ := json.Marshal(arg)
	dataCollector.redisHelper.HSet(ctx, BINANCE_SUBSCRIBED_CHANNELS, "btcusdt@aggTrade", jsonBytes1)

	arg2 := &binance_define.RedisChannelArg{
		TypeSubscribe: "subscribe",
		Channel:       "miniTicker",
		TokenPair:     "btcusdt",
	}
	jsonBytes2, _ := json.Marshal(arg2)
	dataCollector.redisHelper.HSet(ctx, BINANCE_SUBSCRIBED_CHANNELS, "btcusdt@miniTicker", jsonBytes2)

	arg3 := &binance_define.RedisChannelArg{
		TypeSubscribe: "subscribe",
		Channel:       "kline_1m",
		TokenPair:     "btcusdt",
	}
	jsonBytes3, _ := json.Marshal(arg3)
	dataCollector.redisHelper.HSet(ctx, BINANCE_SUBSCRIBED_CHANNELS, "btcusdt@kline_1m", jsonBytes3)

	err := dataCollector.RestoreSubscribedChannels(ctx)
	if err != nil {
		logger.Error("Failed to restore subscribed channels: %v", err)
		return nil, err
	}
	logger.Debug("Restored subscribed channels from Redis")
	dataCollectorInstance = dataCollector
	return dataCollector, nil
}

func (dc *DataCollector) Start(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Second) // Adjust the interval as needed
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("Data collection stopped.")
			return nil
		case <-ticker.C:
			dc.collectData(ctx)
		}
	}
}

func (dc *DataCollector) collectData(ctx context.Context) {
	// Fetch data from Binance API
	pushData := dc.apiClient.FetchData()
	if len(pushData) > 0 {
		// log.Printf("Stored %d records from Binance", len(pushData))
		dc.storer.StoreData(ctx, pushData)
	}
}

// func (dc *DataCollector) AddRedisChannelOption(ctx context.Context, op string, redisOkxChannelArg *okx_define.RedisOkxChannelArg) {
// 	channelArgsJson, _ := json.MarshalIndent(redisOkxChannelArg, "", "  ")

// 	field := fmt.Sprintf("%s:%s", redisOkxChannelArg.Channel, redisOkxChannelArg.TokenContractAddress)
// 	switch op {
// 	default:
// 		log.Printf("Unknown operation: %s", op)
// 	case "subscribe":
// 		if dc.apiClient.AddChannelSubscribe(redisOkxChannelArg) {
// 			logger.Debug("RedisChannelOption, %s: %s", op, string(channelArgsJson))
// 			for {
// 				argJson, _ := json.Marshal(redisOkxChannelArg)
// 				// { // TODO: 仅测试用, 正式版数据库有值, 不需要每次都设置
// 				// 	dc.storer.SetTokenContractIdToMap(redisOkxChannelArg.ChainIndex, redisOkxChannelArg.TokenContractAddress, redisOkxChannelArg.TokenContractId)
// 				// }
// 				err := dc.redisHelper.HSet(ctx, OKX_SUBSCRIBED_CHANNELS, field, argJson)
// 				if err == nil {
// 					break
// 				}
// 				// 写入失败，重试
// 				logger.Error("Failed to record subscription: %v, retrying...", err)
// 				time.Sleep(1 * time.Millisecond)
// 			}
// 		}
// 	case "unsubscribe":
// 		if dc.apiClient.AddChannelUnsubscribe(redisOkxChannelArg) {
// 			logger.Debug("RedisChannelOption, %s: %s", op, string(channelArgsJson))
// 			for {
// 				dc.redisHelper.HDel(ctx, OKX_SUBSCRIBED_CHANNELS, field)
// 				_, err := dc.redisHelper.HGet(ctx, OKX_SUBSCRIBED_CHANNELS, field)
// 				if err == redis.Nil {
// 					break
// 				}
// 				logger.Error("Failed to record unsubscription: %v, retrying...", err)
// 				time.Sleep(1 * time.Millisecond)
// 			}
// 		}
// 	}
// }

// 程序启动时, 从 Redis 读取已订阅频道列表, 恢复订阅 理论上不存在取消订阅数据
func (dc *DataCollector) RestoreSubscribedChannels(ctx context.Context) error {
	subscribedChannelsMap, err := dc.redisHelper.HGetAll(ctx, BINANCE_SUBSCRIBED_CHANNELS)
	logger.Debug("BINANCE_SUBSCRIBED_CHANNELS HGetAll", subscribedChannelsMap)
	if err != nil {
		if err == redis.Nil {
			// 没有已订阅频道, 正常返回
			return nil
		}
		return err
	}

	//循环定位哪些代币对应该存于哪个连接对象中，再批量插入执行
	channelMap := make(map[int32]*utils.List[*binance_define.RedisChannelArg])
	for _, v := range subscribedChannelsMap {
		var channelArg binance_define.RedisChannelArg
		if err := json.Unmarshal([]byte(v), &channelArg); err != nil {
			logger.Error("Failed to unmarshal subscribed channel arg: %v", err)
			continue
		}
		// logger.Debug("Unmarshal", channelArg)
		index := dc.apiClient.GetChanneJoinIndex(&channelArg)
		logger.Debug("index", index)
		if index >= 0 {
			if channelMap[index] == nil {
				channelMap[index] = &utils.List[*binance_define.RedisChannelArg]{}
			}
			channelMap[index].Append(&channelArg)
		}
	}

	//循环批量插入各个client
	for index, list := range channelMap {
		logger.Debug("批量插入订阅信息  s% 插入信息长度 s%", index, len(list.All()))
		dc.apiClient.BatcxhAddChannelSubscribe(index, list)
	}

	return nil
}
