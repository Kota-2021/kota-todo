package handler

import (
	"my-portfolio-2025/internal/app/service"
	"my-portfolio-2025/internal/app/repository"
	"github.com/gin-gonic/gin"
)

// ... 初期設定（DB接続、Gorm初期化）...

// 依存性の注入（DI）
userRepo := repository.NewUserRepository(db)
authService := service.NewAuthService(userRepo)
authController := handler.NewAuthController(authService)

// Ginルーターの設定
r := gin.Default()

// ルート定義
api := r.Group("/api")
{
    // ユーザー登録エンドポイント
    api.POST("/signup", authController.Signup)
    api.POST("/signin", authController.Signin)
    
    // ログインエンドポイント（次のステップで実装）
    // api.POST("/signin", authController.Signin) 
}

// r.Run(":8080")