// internal/app/middleware/middleware.go
package middleware

import (
	"fmt"
	"net/http" // HTTPステータスコードのために必要
	"strings"  // ヘッダー文字列操作のために必要

	"my-portfolio-2025/pkg/auth"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware は、リクエストヘッダーからJWTを取得し、検証するGinミドルウェアです。
func AuthMiddleware() gin.HandlerFunc {
	const BEARER_SCHEMA = "Bearer " // プレフィックスを定数として定義

	// gin.HandlerFunc を返す関数として定義します。
	return func(c *gin.Context) {
		// --- ステップ 3.3: トークン抽出の準備 ---
		// リクエストヘッダーから Authorization フィールドの値を取得します。
		authHeader := c.GetHeader("Authorization")

		// --- ステップ 3.5: エラーハンドリング（トークンが存在しない場合） ---
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			return
		}

		// 2. トークンの形式チェックと抽出（ステップ 3.3 の実装）
		// Authorizationヘッダーが "Bearer " で始まっているか確認
		if !strings.HasPrefix(authHeader, BEARER_SCHEMA) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token format. Must be 'Bearer <token>'",
			})
			return
		}

		// トークン文字列を抽出
		// strings.TrimPrefixを使用して "Bearer " の部分を取り除く
		tokenString := strings.TrimPrefix(authHeader, BEARER_SCHEMA)

		// 3. JWT検証ロジックの呼び出し (3.1で作成)
		userID, err := auth.ValidateToken(tokenString)

		// 4. エラーハンドリング (3.5) - 検証失敗の場合
		if err != nil {
			// 詳細なエラーメッセージをログに出力しつつ、クライアントには一般的なエラーを返すのがベストプラクティスだが、
			// ここでは分かりやすさのためエラー内容を返す様にしておく。
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": fmt.Sprintf("Token validation failed: %s", err.Error()),
			})
			return
		}

		// 5. 認証成功: UserIDをGinのコンテキストに格納
		// このUserIDは、後続のハンドラー（タスク作成、参照など）で c.MustGet("userID") を使って取得されます。
		c.Set("userID", userID)

		// 認証が成功した場合、後続のハンドラーを実行します。
		c.Next()
	}
}
