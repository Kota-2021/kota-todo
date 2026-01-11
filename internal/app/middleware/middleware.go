// internal/app/middleware/middleware.go
package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"my-portfolio-2025/pkg/auth"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware は、リクエストヘッダーからJWTを取得し、検証するGinミドルウェアです。
func AuthMiddleware() gin.HandlerFunc {
	const BEARER_SCHEMA = "Bearer "

	return func(c *gin.Context) {
		var tokenString string

		// 1. Authorization ヘッダーの確認
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, BEARER_SCHEMA) {
			tokenString = strings.TrimPrefix(authHeader, BEARER_SCHEMA)
		} else {
			// 2. クエリパラメータの確認 (WebSocket接続用)
			tokenString = c.Query("token")
		}

		// トークンが存在しない場合
		if tokenString == "" {
			// AbortWithStatusJSON を使い、他のハンドラーと共通のレスポンス形式を維持
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "認証トークンが必要です"})
			return
		}

		// 3. トークンの検証
		userID, err := auth.ValidateToken(tokenString)
		if err != nil {
			// 認証失敗はセキュリティ監査のために Warn レベルで記録する
			slog.Warn("Authentication failed",
				"error", err,
				"path", c.Request.URL.Path,
				"method", c.Request.Method,
				"client_ip", c.ClientIP(),
			)

			// セキュリティ上、詳細なエラー理由はクライアントに返さず固定メッセージにする
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "無効なトークンです"})
			return
		}

		// 以降の処理で userID を型安全に利用できるようにコンテキストに保存
		c.Set("userID", userID)

		// 正常終了して次の処理へ
		c.Next()
	}
}
