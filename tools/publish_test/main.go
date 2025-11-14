package main

import (
	"Conso-CoinHub-DataCollector/internal/api/okx_define"
	"Conso-CoinHub-DataCollector/internal/cache"
	"Conso-CoinHub-DataCollector/internal/collector"
	"Conso-CoinHub-DataCollector/pkg/utils"
	"context"
	"fmt"
	"log"
)

func subscribeChannel(ctx context.Context, rs *utils.RedisStream[collector.RedisOKXMessage], args []*okx_define.RedisOkxChannelArg) error {
	option := "subscribe"
	msg := collector.RedisOKXMessage{
		Type:        option,
		ChannelArgs: args,
		Timestamp:   utils.GetCurrentTimestampNano(),
		MessageId:   fmt.Sprintf("unique-message-id-%s-%d", option, utils.GetCurrentTimestampNano()),
	}
	_, err := rs.PublishStream(ctx, collector.OKX_SUBSCRIPTION_STREAM, msg)
	if err != nil {
		log.Fatalf("Error publishing stream message: %v", err)
	}
	return nil
}

func unsubscribeChannel(ctx context.Context, rs *utils.RedisStream[collector.RedisOKXMessage], args []*okx_define.RedisOkxChannelArg) error {
	option := "unsubscribe"
	msg := collector.RedisOKXMessage{
		Type:        option,
		ChannelArgs: args,
		Timestamp:   utils.GetCurrentTimestampNano(),
		MessageId:   fmt.Sprintf("unique-message-id-%s-%d", option, utils.GetCurrentTimestampNano()),
	}
	_, err := rs.PublishStream(ctx, collector.OKX_SUBSCRIPTION_STREAM, msg)
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
	rs := utils.NewRedisStream[collector.RedisOKXMessage](redisClient)

	// unsubscribeChannel(ctx, rs, []*okx_define.RedisOkxChannelArg{
	// 	{
	// 		Channel:              "price",
	// 		TokenContractId:      "b98303de-a664-4c6f-a08c-c4ac65e50c10",
	// 		ChainIndex:           "1",
	// 		TokenContractAddress: "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
	// 	},
	// })
	subscribeChannel(ctx, rs, []*okx_define.RedisOkxChannelArg{
		{
			Channel:              "trades",
			TokenContractId:      "b98303de-a664-4c6f-a08c-c4ac65e50c10",
			ChainIndex:           "1",
			TokenContractAddress: "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
		},
		{
			Channel:              "price",
			TokenContractId:      "b98303de-a664-4c6f-a08c-c4ac65e50c10",
			ChainIndex:           "1",
			TokenContractAddress: "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
		},
		{
			Channel:              "dex-token-candle1s",
			TokenContractId:      "b98303de-a664-4c6f-a08c-c4ac65e50c10",
			ChainIndex:           "1",
			TokenContractAddress: "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
		},
	})
	// unsubscribeChannel(ctx, rs, []*okx_define.RedisOkxChannelArg{
	// 	{
	// 		Channel:              "trades",
	// 		TokenContractId:      "b98303de-a664-4c6f-a08c-c4ac65e50c10",
	// 		ChainIndex:           "1",
	// 		TokenContractAddress: "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
	// 	},
	// 	{
	// 		Channel:              "price",
	// 		TokenContractId:      "b98303de-a664-4c6f-a08c-c4ac65e50c10",
	// 		ChainIndex:           "1",
	// 		TokenContractAddress: "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
	// 	},
	// })
}

// type RedisChannelArg struct {
//                    string `json:"op"` // 不序列化, 本地用来区分订阅和取消订阅的
// 	Channel              string `json:"channel"`
// 	ChainIndex           string `json:"chainIndex"`
// 	TokenContractAddress string `json:"tokenContractAddress"`
// 	TokenContractId      string `json:"tokenContractId"`
// }
