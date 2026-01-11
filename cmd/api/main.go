package main

import (
	"context"
	"fmt"
	"log/slog"
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

// initLogger は環境に応じて slog を初期化します
func initLogger() {
	var handler slog.Handler
	if os.Getenv("APP_ENV") == "production" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}
	slog.SetDefault(slog.New(handler))
}

// loadEnv は .env ファイルを読み込みます
func loadEnv() {
	if err := godotenv.Load(); err != nil {
		slog.Info(".env file not found, using environment variables")
	} else {
		slog.Info(".env file loaded successfully")
	}
}

// setupDatabase はDB接続の確立、テスト、マイグレーションを行います
func setupDatabase() *gorm.DB {
	slog.Info("Starting database connection...")

	// 環境変数チェック
	requiredEnvs := []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE"}
	for _, env := range requiredEnvs {
		if os.Getenv(env) == "" {
			slog.Error("Required environment variable is missing", "variable", env)
			os.Exit(1)
		}
	}

	encodedPassword := url.QueryEscape(os.Getenv("DB_PASSWORD"))
	dbURI := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s&TimeZone=Asia/Tokyo",
		os.Getenv("DB_USER"), encodedPassword, os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"), os.Getenv("DB_NAME"), os.Getenv("DB_SSLMODE"),
	)

	var db *gorm.DB
	var err error
	maxRetries := 30

	for i := 1; i <= maxRetries; i++ {

		db, err = gorm.Open(postgres.Open(dbURI), &gorm.Config{})
		if err == nil {
			if sqlDB, pingErr := db.DB(); pingErr == nil {
				if pingErr = sqlDB.Ping(); pingErr == nil {
					slog.Info("PostgreSQL connected successfully")
					break
				}
				err = pingErr
			}
		}

		if i == maxRetries {
			slog.Error("Could not connect to database after maximum retries", "error", err)
			os.Exit(1)
		}

		slog.Warn("Database connection failed, retrying...", "attempt", i, "max", maxRetries, "error", err)
		time.Sleep(2 * time.Second)
	}

	// マイグレーション
	if err := db.AutoMigrate(&models.User{}, &models.Task{}, &models.Notification{}); err != nil {
		slog.Error("Database migration failed", "error", err)
		os.Exit(1)
	}
	slog.Info("Database migration completed")

	return db
}

func main() {
	// 1. ログと環境変数の初期設定
	initLogger()
	fmt.Println("INFO: API process started - Preparing for initialization...")
	loadEnv()

	// 2. 基盤 Context と DB 初期化
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db := setupDatabase()

	// 3. インフラ層の初期化
	// Redis (Pingなし版を想定)
	rdb, err := redis.NewRedisClient()
	if err != nil {
		slog.Error("Failed to initialize Redis", "error", err)
		// Redisが必須の構成（RateLimit等）であれば、ここで終了させる
		os.Exit(1)
	}

	// SQS
	queueName := os.Getenv("SQS_QUEUE_NAME")
	sqsClient, err := aws.NewSQSClient(ctx, queueName)
	if err != nil {
		slog.Error("SQS initialization failed", "error", err)
		// 本番ワーカーモードなら Exit、APIモードなら続行などの判断が可能
	}

	// 4. DI (依存性注入)
	// Repositories
	userRepo := repository.NewUserRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	notiRepo := repository.NewNotificationRepository(db)

	// Hub & Services
	hub := service.NewNotificationHub(rdb)
	go hub.Run(ctx)

	authService := service.NewAuthService(userRepo)
	notiService := service.NewNotificationService(notiRepo)

	// WorkerService
	workerService := service.NewWorkerService(sqsClient, taskRepo, notiService, hub)

	// Task/Auth Handler dependencies
	taskService := service.NewTaskService(taskRepo, workerService)

	authHandler := handler.NewAuthController(authService)
	taskHandler := handler.NewTaskHandler(taskService)
	notificationHandler := handler.NewNotificationHandler(notiService, hub)

	// 5. 実行モードの判定
	mode := os.Getenv("MODE")

	if mode == "worker" {
		slog.Info("Starting in WORKER mode")
		if sqsClient == nil {
			slog.Error("Worker mode requires a valid SQS client")
			os.Exit(1)
		}

		go workerService.StartTaskWatcher(ctx)
		slog.Info("Worker service is polling SQS")
		workerService.StartWorker(ctx) // 無限ループ

	} else {
		slog.Info("Starting in API server mode")

		// Gin の動作モード設定
		if os.Getenv("APP_ENV") == "production" {
			gin.SetMode(gin.ReleaseMode)
		}

		r := router.SetupRouter(authHandler, taskHandler, notificationHandler, rdb)

		// ヘルスチェック (slog を活用)
		r.GET("/health", func(c *gin.Context) {
			sqlDB, _ := db.DB()
			if err := sqlDB.Ping(); err != nil {
				slog.Error("Healthcheck failed: DB disconnected", "error", err)
				c.JSON(http.StatusInternalServerError, gin.H{"status": "error"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}

		slog.Info("API server starting", "port", port, "env", os.Getenv("APP_ENV"))
		if err := r.Run(":" + port); err != nil {
			slog.Error("Server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}
}
