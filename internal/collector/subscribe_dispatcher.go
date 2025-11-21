package collector

import (
	"Centralized-Data-Collector/internal/api/binance_define"
	"Centralized-Data-Collector/pkg/logger"
	"Centralized-Data-Collector/pkg/utils"
	"context"
	"log"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisBinanceMessage struct {
	Type        string                            `json:"type"`       // 消息类型，如 "subscribe_channel" 或 "unsubscribe_channel"
	ChannelArgs []*binance_define.RedisChannelArg `json:"channelArg"` //  的订阅时是
	Timestamp   int64                             `json:"timestamp"`  // 消息生成时间戳（Unix 毫秒）
	MessageId   string                            `json:"message_id"` // 业务端生成的MessageId，用于去重
}

// // SubscribeDispatcher 负责从 Redis Stream 拉取订阅参数并调用 OKX 订阅
type SubscribeDispatcher struct {
	RedisClient   *redis.Client
	dataCollector *DataCollector
	Stream        string // Redis Stream 名称
	Group         string // 消费者组
	Consumer      string // 消费者名
	RedisHelper   *utils.RedisHelper
}

func NewSubscribeDispatcher(redisClient *redis.Client, dataCollector *DataCollector, stream, group, consumer string) *SubscribeDispatcher {
	return &SubscribeDispatcher{
		RedisClient:   redisClient,
		dataCollector: dataCollector,
		Stream:        stream,
		Group:         group,
		Consumer:      consumer,
		RedisHelper:   utils.NewRedisHelper(redisClient),
	}
}

// Start 启动订阅参数监听与分发
func (d *SubscribeDispatcher) Start(ctx context.Context) error {
	rs := utils.NewRedisStream[RedisBinanceMessage](d.RedisClient)

	// 确保消费者组存在
	if err := rs.EnsureGroup(ctx, d.Stream, d.Group); err != nil {
		if strings.Contains(err.Error(), "BUSYGROUP") {
			logger.Warn("Consumer group already exists, continue: %v", err)
		} else {
			logger.Error("Failed to ensure consumer group: %v", err)
			return err
		}
	}
	msgCh, cancel, err := rs.SubscribeStream(ctx, d.Stream, d.Group, d.Consumer, 10*time.Second)
	if err != nil {
		return err
	}
	// go func() {
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-msgCh:
			if !ok {
				return nil
			}
			log.Printf("收到订阅参数: %+v，开始订阅", msg)
			for _, channelArg := range msg.Value.ChannelArgs {
				d.dataCollector.AddRedisChannelOption(ctx, msg.Value.Type, channelArg)
			}
			rs.AckStream(ctx, d.Stream, d.Group, msg.ID)
		}
	}
}
