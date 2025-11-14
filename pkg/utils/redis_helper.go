package utils

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisHelper 提供通用的 Redis 操作方法
type RedisHelper struct {
	Client *redis.Client
}

// NewRedisHelper 创建 RedisHelper 实例
func NewRedisHelper(client *redis.Client) *RedisHelper {
	return &RedisHelper{Client: client}
}

// -------------------- KV 操作 --------------------

// SetString 设置字符串
func (r *RedisHelper) SetString(ctx context.Context, key string, value string, ttl time.Duration) error {
	return r.Client.Set(ctx, key, value, ttl).Err()
}

// GetString 获取字符串
func (r *RedisHelper) GetString(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

// SetInt 设置整数
func (r *RedisHelper) SetInt(ctx context.Context, key string, value int64, ttl time.Duration) error {
	return r.Client.Set(ctx, key, value, ttl).Err()
}

// GetInt 获取整数
func (r *RedisHelper) GetInt(ctx context.Context, key string) (int64, error) {
	return r.Client.Get(ctx, key).Int64()
}

// SetJSON 设置任意结构体为 JSON
func (r *RedisHelper) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.Client.Set(ctx, key, data, ttl).Err()
}

// GetJSON 获取 JSON 并反序列化为指定类型
func (r *RedisHelper) GetJSON(ctx context.Context, key string, out interface{}) error {
	data, err := r.Client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

// -------------------- List 队列 --------------------

// LPush 队列左入
func (r *RedisHelper) LPush(ctx context.Context, key string, values ...interface{}) error {
	return r.Client.LPush(ctx, key, values...).Err()
}

// RPop 队列右出
func (r *RedisHelper) RPop(ctx context.Context, key string) (string, error) {
	return r.Client.RPop(ctx, key).Result()
}

// BLPop 阻塞左出
func (r *RedisHelper) BLPop(ctx context.Context, key string, timeout time.Duration) (string, error) {
	res, err := r.Client.BLPop(ctx, timeout, key).Result()
	if err != nil || len(res) < 2 {
		return "", err
	}
	return res[1], nil
}

// -------------------- Hash --------------------

// HSet 设置 Hash 字段
func (r *RedisHelper) HSet(ctx context.Context, key string, field string, value interface{}) error {
	return r.Client.HSet(ctx, key, field, value).Err()
}

// HGet 获取 Hash 字段
func (r *RedisHelper) HGet(ctx context.Context, key string, field string) (string, error) {
	return r.Client.HGet(ctx, key, field).Result()
}

// HGetAll 获取整个 Hash
func (r *RedisHelper) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.Client.HGetAll(ctx, key).Result()
}

// HDel 删除 Hash 字段
func (r *RedisHelper) HDel(ctx context.Context, key string, fields ...string) error {
	return r.Client.HDel(ctx, key, fields...).Err()
}

// -------------------- Set --------------------

// SAdd 添加 Set 元素
func (r *RedisHelper) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return r.Client.SAdd(ctx, key, members...).Err()
}

// SMembers 获取 Set 所有元素
func (r *RedisHelper) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.Client.SMembers(ctx, key).Result()
}

// -------------------- Pub/Sub --------------------

// Publish 发布消息
func (r *RedisHelper) Publish(ctx context.Context, channel string, message interface{}) error {
	var msg string
	switch v := message.(type) {
	case string:
		msg = v
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		msg = string(b)
	}
	return r.Client.Publish(ctx, channel, msg).Err()
}

// Subscribe 订阅消息（返回只读通道和取消函数）
func (r *RedisHelper) Subscribe(ctx context.Context, channel string) (<-chan *redis.Message, func(), error) {
	sub := r.Client.Subscribe(ctx, channel)
	_, err := sub.Receive(ctx)
	if err != nil {
		return nil, nil, err
	}
	ch := sub.Channel()
	done := make(chan struct{})
	out := make(chan *redis.Message, 1024)
	go func() {
		for {
			select {
			case <-done:
				sub.Close()
				close(out)
				return
			case msg, ok := <-ch:
				if !ok {
					sub.Close()
					close(out)
					return
				}
				out <- msg
			}
		}
	}()
	cancel := func() {
		close(done)
	}
	return out, cancel, nil
}

// -------------------- Stream 泛型队列 --------------------
// 推荐用你已有的 RedisStream[T] 实现
