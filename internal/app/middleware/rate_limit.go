package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func RateLimiter(rdb *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. コンテキストからUserIDを取得
		val, exists := c.Get("userID")

		var identifier string
		if exists {
			// uuid.UUID 型から string に変換
			if id, ok := val.(uuid.UUID); ok {
				identifier = id.String()
			} else if idStr, ok := val.(string); ok {
				identifier = idStr
			}
		} else {
			identifier = c.ClientIP()
		}

		// 2. キーの生成
		key := "rate_limit:" + identifier + ":" + c.FullPath()
		ctx := c.Request.Context()

		// --- 以下、Redis処理 ---
		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}

		if count == 1 {
			rdb.Expire(ctx, key, window)
		}

		if int(count) > limit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "リクエスト回数が制限を超えました。しばらく時間をおいてから再試行してください。",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
