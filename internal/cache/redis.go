package cache

import (
	"Centralized-Data-Collector/pkg/logger"
	"context"

	"github.com/go-redis/redis/v8"
)

var redisClientInstance *redis.Client

func RedisClientInstance() *redis.Client {
	return redisClientInstance
}

func InitRedisClient(addr string, password string) (*redis.Client, error) {
	db := 6
	rdb, err := newRedisClient(addr, password, db)
	if err == nil {
		redisClientInstance = rdb
	}
	return rdb, err
}

func newRedisClient(addr string, password string, db int) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test the connection
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		logger.Error("Could not connect to Redis: %v", err)
		return nil, err
	}

	return rdb, nil
}
