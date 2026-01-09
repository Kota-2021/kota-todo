// internal/app/middleware/middleware.go
package middleware

import (
	"net/http" // HTTPステータスコードのために必要
	"strings"  // ヘッダー文字列操作のために必要

	"my-portfolio-2025/pkg/auth"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware は、リクエストヘッダーからJWTを取得し、検証するGinミドルウェアです。
func AuthMiddleware() gin.HandlerFunc {
	const BEARER_SCHEMA = "Bearer "

	return func(c *gin.Context) {
		var tokenString string

		// 1. まずヘッダーを確認
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, BEARER_SCHEMA) {
			tokenString = strings.TrimPrefix(authHeader, BEARER_SCHEMA)
		} else {
			// 2. ヘッダーになければクエリパラメータを確認 (WebSocket用)
			tokenString = c.Query("token")
		}

		// トークンがどこにもなければエラー
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication token required"})
			return
		}

		// 検証
		userID, err := auth.ValidateToken(tokenString)
		if err != nil {
			// 💡 エラー内容をレスポンスに含める（デバッグ用）
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":       "Invalid token",
				"debug_cause": err.Error(), // ← これで「署名ミス」か「期限切れ」か判明します
			})
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}
