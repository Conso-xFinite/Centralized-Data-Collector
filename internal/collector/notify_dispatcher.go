package collector

// import (
// 	"context"
// 	"time"

// 	"Centralized-Data-Collector/pkg/utils"

// 	"github.com/go-redis/redis/v8"
// )

// // 通知消息结构体，可扩展字段
// type NotifyMsg struct {
// 	Event     string    `json:"event"`     // 事件类型，如 "okx_price_stored"
// 	Timestamp time.Time `json:"timestamp"` // 通知时间
// 	Detail    string    `json:"detail"`    // 可选，附加说明
// }

// // NotifyDispatcher 负责向 Redis Stream 推送通知
// type NotifyDispatcher struct {
// 	RedisClient *redis.Client
// 	Stream      string // Redis Stream 名称
// }

// func NewNotifyDispatcher(redisClient *redis.Client, stream string) *NotifyDispatcher {
// 	return &NotifyDispatcher{
// 		RedisClient: redisClient,
// 		Stream:      stream,
// 	}
// }

// // Notify 推送通知消息到 Redis Stream
// func (d *NotifyDispatcher) Notify(ctx context.Context, event, detail string) error {
// 	rs := utils.NewRedisStream[NotifyMsg](d.RedisClient)
// 	msg := NotifyMsg{
// 		Event:     event,
// 		Timestamp: time.Now(),
// 		Detail:    detail,
// 	}
// 	_, err := rs.PublishStream(ctx, d.Stream, msg)
// 	return err
// }
