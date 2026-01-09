// internal/app/router/router.go
package router

import (
	"my-portfolio-2025/internal/app/handler"
	"my-portfolio-2025/internal/app/middleware"

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

	r := gin.Default()

	// --- 認証不要な公開ルート /auth ---
	public := r.Group("/auth")
	{
		public.POST("/signup", authHandler.Signup) // ユーザー登録
		public.POST("/signin", authHandler.Signin) // ログイン（JWT発行）
	}

	// authGroupへ移動させた。260108byKota
	// r.GET("/ws", notificationHandler.HandleWS)

	// --- 認証必須のルート共通設定 ---
	authGroup := r.Group("/")
	authGroup.Use(middleware.AuthMiddleware())
	// authGroup.Use(middleware.RateLimiter(redisClient, 5, time.Minute))
	{
		// タスク関連
		tasks := authGroup.Group("/tasks")
		{
			// 4.5 Create (POST /tasks)
			tasks.POST("", taskHandler.CreateTask)
			// 4.6 Read List (GET /tasks)
			tasks.GET("", taskHandler.GetTasks)
			// 4.7 Read Detail (GET /tasks/:id)
			tasks.GET("/:id", taskHandler.GetTaskByID)
			// 4.8 Update (PUT /tasks/:id)
			tasks.PUT("/:id", taskHandler.UpdateTask)
			// 4.9 Delete (DELETE /tasks/:id)
			tasks.DELETE("/:id", taskHandler.DeleteTask)
		}

		// WebSocket エンドポイント
		websocket := authGroup.Group("/ws")
		{
			websocket.GET("", notificationHandler.HandleWS)
		}

		// 通知関連
		notifications := authGroup.Group("/notifications")
		{
			// 通知エンドポイント
			// GET /notifications (ユーザーの通知を10件ずつ取得)
			notifications.GET("", notificationHandler.GetNotifications)
			// PATCH /notifications/:id/read (指定された通知を既読にします)
			notifications.PATCH("/:id/read", notificationHandler.MarkAsRead)
		}
	}

	return r
}
