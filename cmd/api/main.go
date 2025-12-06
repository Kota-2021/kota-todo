package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"my-portfolio-2025/internal/app/handler"
	"my-portfolio-2025/internal/app/models"
	"my-portfolio-2025/internal/app/repository"
	"my-portfolio-2025/internal/app/router"
	"my-portfolio-2025/internal/app/service"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupDatabase はDB接続の確立、テスト、マイグレーションを行います
func setupDatabase() *gorm.DB {
	log.Println("=== データベース接続開始 ===")

	// 環境変数から接続情報を取得
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbSSLMode := os.Getenv("DB_SSLMODE")

	if dbHost == "" || dbPort == "" || dbUser == "" || dbPassword == "" || dbName == "" || dbSSLMode == "" {
		log.Fatalf("環境変数が設定されていません (DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, DB_SSLMODE)")
	}

	// URI形式の接続文字列を構築
	// パスワードをURLエンコードすることで、特殊文字を含む場合でもGormで安全に扱える
	encodedPassword := url.QueryEscape(dbPassword)
	dbURI := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s&TimeZone=Asia/Tokyo",
		dbUser, encodedPassword, dbHost, dbPort, dbName, dbSSLMode,
	)

	maxRetries := 30
	retryInterval := 1 * time.Second
	var db *gorm.DB
	var err error

	// 接続テストと確立（リトライロジック）
	for i := 0; i < maxRetries; i++ {
		// Gormを使ってDB接続を試みる
		db, err = gorm.Open(postgres.Open(dbURI), &gorm.Config{})
		if err == nil {
			// 接続に成功したら、Pingで生存確認
			sqlDB, pingErr := db.DB()
			if pingErr == nil {
				pingErr = sqlDB.Ping()
			}

			if pingErr == nil {
				log.Println("✓ PostgreSQLへの接続に成功しました！")
				break
			}
			err = pingErr
		}

		if i < maxRetries-1 {
			log.Printf("接続試行 %d/%d 失敗: %v (再試行します...)", i+1, maxRetries, err)
			time.Sleep(retryInterval)
		} else {
			log.Fatalf("接続試行 %d/%d 失敗: %v", i+1, maxRetries, err)
		}
	}

	// データベースマイグレーション
	// データベースが存在しない場合は自動作成される。
	err = db.AutoMigrate(&models.User{}, &models.Task{})
	if err != nil {
		log.Fatalf("Failed to perform database migration: %v", err)
	}
	log.Println("Database migration completed.")

	return db
}

func main() {

	// 1. DB接続の確立とマイグレーション
	db := setupDatabase()

	// 2. 依存性の注入（DI）と各層の初期化
	// Task Handlerを先に初期化できるように、UserとTaskの両方の依存性をここで定義

	// 認証機能の依存性
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo)
	authController := handler.NewAuthController(authService)

	// タスク管理機能の依存性
	taskRepo := repository.NewTaskRepository(db)
	taskService := service.NewTaskService(taskRepo)
	taskHandler := handler.NewTaskHandler(taskService)

	// 3. ルーター設定とハンドラーの紐づけ
	r := router.SetupRouter(authController, taskHandler)

	// ヘルスチェックエンドポイントの追加（ALB/ECS用）
	r.GET("/health", func(c *gin.Context) {
		// DB接続もテスト
		sqlDB, _ := db.DB()
		if sqlDB.Ping() != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "db_connected": false})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "db_connected": true})
	})

	// 4. サーバー起動
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	serverAddr := ":" + port

	log.Printf("Starting API server on http://localhost%s", serverAddr)
	if err := r.Run(serverAddr); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server stopped unexpectedly: %v", err)
	}
}
