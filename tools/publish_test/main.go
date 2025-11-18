package main

import (
	"Centralized-Data-Collector/internal/api/binance_define"
	"Centralized-Data-Collector/internal/cache"
	"Centralized-Data-Collector/internal/collector"
	"Centralized-Data-Collector/pkg/utils"
	"context"
	"fmt"
	"log"
)

func subscribeChannel(ctx context.Context, rs *utils.RedisStream[collector.RedisBinanceMessage], args []*binance_define.RedisChannelArg) error {
	option := "subscribe"
	msg := collector.RedisBinanceMessage{
		Type:        option,
		ChannelArgs: args,
		Timestamp:   utils.GetCurrentTimestampNano(),
		MessageId:   fmt.Sprintf("unique-message-id-%s-%d", option, utils.GetCurrentTimestampNano()),
	}
	_, err := rs.PublishStream(ctx, collector.BINANCE_SUBSCRIPTION_STREAM, msg)
	if err != nil {
		log.Fatalf("Error publishing stream message: %v", err)
	}
	return nil
}

func unsubscribeChannel(ctx context.Context, rs *utils.RedisStream[collector.RedisBinanceMessage], args []*binance_define.RedisChannelArg) error {
	option := "unsubscribe"
	msg := collector.RedisBinanceMessage{
		Type:        option,
		ChannelArgs: args,
		Timestamp:   utils.GetCurrentTimestampNano(),
		MessageId:   fmt.Sprintf("unique-message-id-%s-%d", option, utils.GetCurrentTimestampNano()),
	}
	_, err := rs.PublishStream(ctx, collector.BINANCE_SUBSCRIPTION_STREAM, msg)
	if err != nil {
		log.Fatalf("Error publishing stream message: %v", err)
	}
	return nil
}

func main() {
	// This is a placeholder main function.
	ctx := context.Background()

	// Initialize Redis connection
	redisAddr := "localhost:6379"
	redisPassword := ""
	redisClient, err := cache.InitRedisClient(redisAddr, redisPassword)
	if err != nil {
		log.Fatalf("Error connecting to Redis: %v", err)
	}
	defer redisClient.Close()

	// 创建消息发布者对象
	rs := utils.NewRedisStream[collector.RedisBinanceMessage](redisClient)

	// subscribeChannel(ctx, rs, []*binance_define.RedisChannelArg{
	// 	{
	// 		TypeSubscribe: "subscribe",
	// 		Channel:       "aggTrade",
	// 		TokenPair:     "solusdt",
	// 	},
	// 	{
	// 		TypeSubscribe: "subscribe",
	// 		Channel:       "kline_1m",
	// 		TokenPair:     "solusdt",
	// 	},
	// 	{
	// 		TypeSubscribe: "subscribe",
	// 		Channel:       "miniTicker",
	// 		TokenPair:     "solusdt",
	// 	},
	// })
	unsubscribeChannel(ctx, rs, []*binance_define.RedisChannelArg{
		{
			TypeSubscribe: "unsubscribe",
			Channel:       "kline_1m",
			TokenPair:     "btcusdt",
		},
		{
			TypeSubscribe: "unsubscribe",
			Channel:       "miniTicker",
			TokenPair:     "btcusdt",
		},
		{
			TypeSubscribe: "unsubscribe",
			Channel:       "aggTrade",
			TokenPair:     "btcusdt",
		},
	})
}
