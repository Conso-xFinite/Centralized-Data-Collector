package utils

import (
	"Centralized-Data-Collector/pkg/logger"

	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

// StreamMsg 包含消息内容和 Redis Stream 消消息ID
type StreamMsg[T any] struct {
	ID    string // Redis Stream 消息ID
	Value T      // 反序列化后的消息内容
}

// RedisStream[T] 是一个基于 Redis Streams 的泛型消息队列工具，支持可靠的生产/消费，订阅端重启不丢数据。
type RedisStream[T any] struct {
	Client *redis.Client // Redis 客户端
}

// NewRedisStream 创建一个 RedisStream 实例
func NewRedisStream[T any](client *redis.Client) *RedisStream[T] {
	return &RedisStream[T]{Client: client}
}

// PublishStream 将任意类型的消息 value 推送到指定 stream。
// value 会被序列化为 JSON，存储在 "payload" 字段。
func (r *RedisStream[T]) PublishStream(ctx context.Context, stream string, value T) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	id, err := r.Client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: map[string]interface{}{"payload": data},
	}).Result()
	return id, err
}

// EnsureGroup 确保指定 stream 的消费者组存在，不存在则创建。
// group: 消费者组名，stream: 流名
func (r *RedisStream[T]) EnsureGroup(ctx context.Context, stream, group string) error {
	t, _ := r.Client.Type(ctx, stream).Result()
	if t != "none" && t != "stream" {
		logger.Debug("Client.Type", t)
	}
	return r.Client.XGroupCreateMkStream(ctx, stream, group, "0").Err()
}

// SubscribeStream 订阅 stream，自动处理 pending（未确认）消息，保证重启期间数据不丢失。
// 返回：
//   - msgCh: 类型为 T 的只读消息通道
//   - cancel: 关闭订阅的函数
//   - error: 错误信息
//
// 订阅端收到的消息会自动反序列化为 T 类型。
func (r *RedisStream[T]) SubscribeStream(
	ctx context.Context,
	stream, group, consumer string,
	block time.Duration,
) (<-chan StreamMsg[T], func(), error) {
	out := make(chan StreamMsg[T], 1024) // 消息通道
	done := make(chan struct{})          // 关闭信号

	// 1. pending 恢复协程：拉取并处理未确认的历史消息
	func() {
		start := "0"
		for {
			select {
			case <-done:
				close(out)
				return
			default:
			}
			// XAutoClaim 拉取超过 MinIdle 未确认的消息
			res, _, err := r.Client.XAutoClaim(ctx, &redis.XAutoClaimArgs{
				Stream:   stream,
				Group:    group,
				Consumer: consumer,
				// MinIdle:  time.Minute, // 超过1分钟未ack的消息才会被拉回
				MinIdle: 0, // 所有pending未ack的消息都会被拉回
				Start:   start,
				Count:   100,
			}).Result()
			if err != nil {
				if err == redis.Nil {
					break // 没有更多pending消息
				}
				logger.Warn("Redis XAutoClaim error: %v", err)
				time.Sleep(time.Second)
				continue
			} else if len(res) == 0 {
				break // 没有更多pending消息
			}
			for _, msg := range res {
				var v T
				// 反序列化 payload 字段
				if payload, ok := msg.Values["payload"].(string); ok {
					_ = json.Unmarshal([]byte(payload), &v)
					select {
					case out <- StreamMsg[T]{ID: msg.ID, Value: v}:
					default:
					}
				}
			}
			if len(res) < 100 {
				break
			}
			start = res[len(res)-1].ID
		}
	}()

	// 2. 新消息消费协程：持续读取新消息
	go func() {
		for {
			select {
			case <-done:
				close(out)
				return
			default:
			}
			res, err := r.Client.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    group,
				Consumer: consumer,
				Streams:  []string{stream, ">"},
				Count:    100,
				Block:    block,
			}).Result()
			if err != nil {
				if err != redis.Nil {
					logger.Warn("Redis XReadGroup error: %v", err)
				}
				time.Sleep(time.Second)
				continue
			}
			for _, st := range res {
				for _, msg := range st.Messages {
					var v T
					if payload, ok := msg.Values["payload"].(string); ok {
						_ = json.Unmarshal([]byte(payload), &v)
						select {
						case out <- StreamMsg[T]{ID: msg.ID, Value: v}:
						default:
						}
					}
				}
			}
		}
	}()

	// 关闭函数，调用后会安全关闭所有协程和通道
	cancel := func() {
		select {
		case <-done:
		default:
			close(done)
		}
	}
	return out, cancel, nil
}

// AckStream 消费完成后确认消息，防止重复消费。
// ids: 消息ID列表
func (r *RedisStream[T]) AckStream(ctx context.Context, stream, group string, ids ...string) error {
	return r.Client.XAck(ctx, stream, group, ids...).Err()
}

/*
使用示例:

type PriceMsg struct {
    Price float64
    Time  int64
}

rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
rs := utils.NewRedisStream[PriceMsg](rdb)

// 生产者推送
rs.PublishStream(ctx, "okx:price", PriceMsg{Price: 4080.5, Time: time.Now().Unix()})

// 消费者订阅
_ = rs.EnsureGroup(ctx, "okx:price", "collector-group")
msgCh, cancel, _ := rs.SubscribeStream(ctx, "okx:price", "collector-group", "collector-1", 5*time.Second)
go func() {
    for msg := range msgCh {
        // msg 是 PriceMsg 类型
        // 处理 msg
        // 需要拿到消息ID才能ack，可扩展SubscribeStream返回消息结构体（带ID）
    }
}()
*/
