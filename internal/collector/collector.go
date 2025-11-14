package collector

import (
	"Centralized-Data-Collector/internal/api/binance_api"
	"Centralized-Data-Collector/pkg/utils"
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

// const OKX_SUBSCRIPTION_STREAM = "coinhub:subscription:okx:channel"
// const OKX_SUBSCRIBED_CHANNELS = "data_collector:subscribed:okx:channel"

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
	okxConnPool := binance_api.NewWSPool(1)
	dataCollector := &DataCollector{
		apiClient:   okxConnPool,
		redisHelper: utils.NewRedisHelper(redisClient),
		storer:      NewBinanceStorer(),
	}

	// err := dataCollector.RestoreSubscribedChannels(ctx)
	// if err != nil {
	// 	logger.Error("Failed to restore subscribed channels: %v", err)
	// 	return nil, err
	// }
	// logger.Debug("Restored subscribed channels from Redis")
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

// // 程序启动时, 从 Redis 读取已订阅频道列表, 恢复订阅
// func (dc *DataCollector) RestoreSubscribedChannels(ctx context.Context) error {
// 	subscribedChannelsMap, err := dc.redisHelper.HGetAll(ctx, OKX_SUBSCRIBED_CHANNELS)
// 	if err != nil {
// 		if err == redis.Nil {
// 			// 没有已订阅频道, 正常返回
// 			return nil
// 		}
// 		return err
// 	}

// 	for _, v := range subscribedChannelsMap {
// 		var channelArg okx_define.RedisOkxChannelArg
// 		if err := json.Unmarshal([]byte(v), &channelArg); err != nil {
// 			logger.Error("Failed to unmarshal subscribed channel arg: %v", err)
// 			continue
// 		}
// 		dc.apiClient.AddChannelSubscribe(&channelArg)
// 	}

// 	return nil
// }
