package r

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

type RedisService struct {
	rdb *redis.Client
}

func NewRedisService(addr string) *RedisService {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisService{rdb: rdb}
}

func (s *RedisService) Set(key string, value interface{}, expiration time.Duration) error {
	return s.rdb.Set(context.Background(), key, value, expiration).Err()
}

func (s *RedisService) Get(key string) (string, error) {
	return s.rdb.Get(context.Background(), key).Result()
}
