package main

import (
	"context"
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
	"my-portfolio-2025/internal/infrastructure/aws"
	"my-portfolio-2025/internal/infrastructure/redis"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupDatabase はDB接続の確立、テスト、マイグレーションを行います
func setupDatabase() *gorm.DB {

	log.Println("=== データベース接続開始 ===")

	// ローカルでの開発用.envファイルの読み込み
	currentPath, errEnv := os.Getwd()
	if errEnv != nil {
		log.Fatal("Error getting current path")
	}
	envFilePath := currentPath + "/.env"

	if _, err := os.Stat(envFilePath); err == nil {
		errEnv := godotenv.Load(envFilePath)
		if errEnv != nil {
			log.Printf("Notice: .env file found at %s but could not be loaded: %v", envFilePath, errEnv)
		} else {
			log.Println("✓ .env file loaded successfully")
		}
	} else {
		log.Println("Notice: .env file not found, skipping. (This is normal in production)")
	}
	// ここまでがローカルでの開発用.envファイルの読み込みの処理。

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
	err = db.AutoMigrate(&models.User{}, &models.Task{}, &models.Notification{})
	if err != nil {
		log.Fatalf("Failed to perform database migration: %v", err)
	}
	log.Println("Database migration completed.")

	return db
}

func main() {

	// 1. DB接続の確立とマイグレーション
	db := setupDatabase()

	// --- 基盤となる Context の生成 ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// --- Redisクライアントの初期化 ---
	rdb := redis.NewRedisClient()

	// --- NotificationHub の初期化 (Redisクライアントを渡す) ---
	hub := service.NewNotificationHub(rdb)

	go hub.Run()               // Hubのイベントループを開始
	go hub.SubscribeRedis(ctx) // Redisの購読ループをバックグラウンドで開始

	// SQSクライアントの初期化
	// ローカルでの開発用.envファイルの読み込み
	currentPath, err := os.Getwd()
	if err != nil {
		log.Fatal("Error getting current path")
	}
	envFilePath := currentPath + "/.env"
	if _, err := os.Stat(envFilePath); err == nil {
		godotenv.Load(envFilePath)
	}
	// ここまでがローカルでの開発用.envファイルの読み込みの処理。

	// --- SQS関連の環境変数チェック (共通) ---
	region := os.Getenv("AWS_REGION")
	queueURL := os.Getenv("SQS_QUEUE_URL")
	queueName := os.Getenv("SQS_QUEUE_NAME")
	if region == "" || queueURL == "" {
		log.Println("Warning: AWS_REGION or SQS_QUEUE_URL is not set. Worker may not function correctly.")
	}

	// --- 依存性の注入（DI）と各層の初期化 ---
	userRepo := repository.NewUserRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	notiRepo := repository.NewNotificationRepository(db)

	authService := service.NewAuthService(userRepo)
	authHandler := handler.NewAuthController(authService)

	notiService := service.NewNotificationService(notiRepo)
	notificationHandler := handler.NewNotificationHandler(notiService, hub)

	// --- 非同期ワーカーの依存性 ---
	// SQSクライアントを初期化
	sqsClient, err := aws.NewSQSClient(ctx, queueName)
	if err != nil {
		log.Printf("SQS初期化失敗: %v", err)
		// 本番環境では Fatalf にする検討も必要ですが、まずは実行を優先
	}

	// WorkerService を初期化 (taskRepoを渡すことで、二重送信防止の更新を可能にする)
	workerService := service.NewWorkerService(sqsClient, taskRepo, notiService, hub)
	taskService := service.NewTaskService(taskRepo, workerService)
	taskHandler := handler.NewTaskHandler(taskService)

	// ==========================================
	// ここから分岐処理
	// ==========================================
	mode := os.Getenv("MODE")

	if mode == "worker" {
		// --- 【Workerモード】 ---
		log.Println("🚾 Starting in WORKER mode...")

		if sqsClient != nil {
			// WatcherとWorkerをバックグラウンドではなく、メインスレッドを維持する形で実行
			go workerService.StartTaskWatcher(ctx)

			log.Println("✓ Worker service is polling SQS...")
			// StartWorker は中で無限ループしている想定のため、ここでプロセスをブロックする
			workerService.StartWorker(ctx)
		} else {
			log.Fatal("Worker mode failed: SQS client is nil")
		}

	} else {
		// --- 【API/Defaultモード】 ---
		log.Println("🅰️ Starting in API server mode...")

		// ルーター設定
		r := router.SetupRouter(authHandler, taskHandler, notificationHandler, rdb)

		// ヘルスチェック
		r.GET("/health", func(c *gin.Context) {
			sqlDB, _ := db.DB()
			if sqlDB.Ping() != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "db_connected": false})
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": "ok", "db_connected": true})
		})

		// サーバー起動
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
}
