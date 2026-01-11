// internal/app/router/router.go
package router

import (
	"log/slog"
	"my-portfolio-2025/internal/app/handler"
	"my-portfolio-2025/internal/app/middleware"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// SetupRouter は全てのルートを設定します
func SetupRouter(
	authHandler *handler.AuthController,
	taskHandler *handler.TaskHandler,
	notificationHandler *handler.NotificationHandler,
	redisClient *redis.Client,
) *gin.Engine {

	// gin.Default() ではなく gin.New() を使用
	r := gin.New()

	// 1. パニックが起きてもサーバーを落とさないためのリカバリミドルウェア
	r.Use(gin.Recovery())

	// 2. 標準の Logger の代わりに、必要に応じて slog を使ったアクセスログを
	// 出力するようにすると、本番環境での解析がさらに楽になります。
	// ここではシンプルさを保つため標準の Logger を使用（または自作 slog ミドルウェア）
	r.Use(gin.Logger())

	// --- 認証不要な公開ルート /auth ---
	public := r.Group("/auth")
	{
		public.POST("/signup", authHandler.Signup) // ユーザー登録
		public.POST("/signin", authHandler.Signin) // ログイン（JWT発行）
	}

	// --- 認証必須のルート共通設定 ---
	authGroup := r.Group("/")
	authGroup.Use(middleware.AuthMiddleware())
	authGroup.Use(middleware.RateLimiter(redisClient, 5, time.Minute))
	{
		// タスク関連
		tasks := authGroup.Group("/tasks")
		{
			tasks.POST("", taskHandler.CreateTask)
			tasks.GET("", taskHandler.GetTasks)
			tasks.GET("/:id", taskHandler.GetTaskByID)
			tasks.PUT("/:id", taskHandler.UpdateTask)
			tasks.DELETE("/:id", taskHandler.DeleteTask)
		}

		// WebSocket エンドポイント
		authGroup.GET("/ws", notificationHandler.HandleWS)

		// 通知関連
		notifications := authGroup.Group("/notifications")
		{
			notifications.GET("", notificationHandler.GetNotifications)
			notifications.PATCH("/:id/read", notificationHandler.MarkAsRead)
		}
	}

	slog.Info("Router setup completed") // 正常にルートが組まれた記録を残す
	return r
}
