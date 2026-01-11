package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func RateLimiter(rdb *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 識別子（UserID または IPアドレス）の決定
		identifier := c.ClientIP() // デフォルトはIP
		if val, exists := c.Get("userID"); exists {
			if id, ok := val.(uuid.UUID); ok {
				identifier = id.String()
			}
		}

		// 2. Redisキーの生成（パスごとに制限をかける）
		key := fmt.Sprintf("rate_limit:%s:%s", identifier, c.FullPath())
		ctx := c.Request.Context()

		// 3. インクリメント処理
		// NOTE: 原子性を高めるため、本来はLuaスクリプトが推奨されますが、
		// 今回は既存のロジックをベースに堅牢化します。
		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			// Redisがダウンしている場合にサービス全体を止めないよう、エラーログを出してパスさせる（Fail-open）
			slog.Error("Redis rate limit error", "error", err, "identifier", identifier)
			c.Next()
			return
		}

		// 初回インクリメント時に有効期限を設定
		if count == 1 {
			rdb.Expire(ctx, key, window)
		}

		// 4. 制限チェック
		if int(count) > limit {
			// 制限に達したことを警告ログとして記録（攻撃検知に役立つ）
			slog.Warn("Rate limit exceeded",
				"identifier", identifier,
				"path", c.FullPath(),
				"count", count,
				"client_ip", c.ClientIP(),
			)

			// 他のハンドラーと統一したエラーレスポンス形式
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "リクエスト回数が制限を超えました。しばらく時間をおいてから再試行してください。",
			})
			return
		}

		c.Next()
	}
}
