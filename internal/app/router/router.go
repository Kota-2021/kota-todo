// internal/app/router/router.go
package router

import (
	"my-portfolio-2025/internal/app/handler"
	"my-portfolio-2025/internal/app/middleware"

	"github.com/gin-gonic/gin"
)

// ... 初期設定（DB接続、Gorm初期化）...

// SetupRouter は全てのルートを設定します
func SetupRouter(
	authController *handler.AuthController,
	taskHandler *handler.TaskHandler,
) *gin.Engine {

	r := gin.Default()

	// --- 認証不要な公開ルート /auth ---
	public := r.Group("/auth")
	{
		public.POST("/signup", authController.Signup) // ユーザー登録
		public.POST("/signin", authController.Signin) // ログイン（JWT発行）
	}

	// --- 認証必須のプライベートルート /tasks ---
	// .Use(middleware.AuthMiddleware()) を使ってミドルウェアを適用します。
	tasks := r.Group("/tasks")
	tasks.Use(middleware.AuthMiddleware()) // このグループ以下の全てのエンドポイントにJWT検証を強制

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

	return r
}
