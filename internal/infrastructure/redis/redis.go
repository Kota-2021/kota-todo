package redis

import (
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient() *redis.Client {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")

	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: "", // 必要に応じて設定
		DB:       0,  // デフォルトDB
	})
}
