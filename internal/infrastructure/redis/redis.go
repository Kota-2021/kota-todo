package redis

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient() (*redis.Client, error) { // エラーを返すように変更
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	addr := fmt.Sprintf("%s:%s", host, port)

	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	// pingで接続確認
	// 5秒以内に返答がなければエラーとする（タイムアウト設定）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		// ここでエラーが出れば、ネットワーク設定や環境変数のミスがすぐわかる
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	slog.Info("Redis client connected successfully", "addr", addr)
	return rdb, nil
}
