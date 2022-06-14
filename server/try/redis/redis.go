package main

import (
	"context"
	"fmt"
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

func main() {
	rs := NewRedisService("localhost:6379")
	if err := rs.Set("update", true, 5*time.Second); err != nil {
		panic(err)
	}

	val, err := rs.Get("update")
	if err == redis.Nil {
		fmt.Println("key does not exist")
	} else if err != nil {
		panic(err)
	} else {
		fmt.Println("key", val)
	}
}
