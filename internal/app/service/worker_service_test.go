package service

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"my-portfolio-2025/internal/app/models"
	"my-portfolio-2025/internal/app/repository"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func init() {
	// テスト実行時に強制的にローカル環境（LocalStack/Post継承）を向くように設定
	os.Setenv("AWS_ENDPOINT", "http://localhost:4566")
	os.Setenv("AWS_REGION", "ap-northeast-1")
	os.Setenv("DATABASE_URL", "host=localhost user=portfolio_admin password=local_dev_password dbname=portfolio_db port=5432 sslmode=disable")

	// もしRedisもテストで使うなら
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")

	log.Println("--- Test Environment Initialized ---")
}

func TestIntegration_NotificationFlow(t *testing.T) {

	db := SetupTestDB(t)
	sqsClient := SetupTestSQS(t)

	// 依存関係をすべて実体で作成 (DI: Dependency Injection)
	taskRepo := repository.NewTaskRepository(db)
	notiRepo := repository.NewNotificationRepository(db)
	notiService := NewNotificationService(notiRepo)
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // init()で設定した環境変数を使ってもOK
	})
	hub := NewNotificationHub(rdb)

	// Hubを起動 (Redisとの通信待機状態にする)
	go hub.Run()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// WorkerService の作成
	workerService := NewWorkerService(sqsClient, taskRepo, notiService, hub)

	// テストデータの作成 (1分以内に期限が来るタスク)
	userID := uuid.New()
	testTask := models.Task{
		ID:             uuid.New(),
		UserID:         userID,
		Title:          "Integration Test Task",
		DueDate:        time.Now().Add(30 * time.Minute), // 1時間以内
		Status:         models.TaskStatusPending,
		LastNotifiedAt: nil, // 明示的にnilにする（GORMならデフォルトでnil）
	}
	db.Create(&testTask)

	// WorkerとWatcherをバックグラウンドで開始
	go workerService.StartWorker(ctx)
	go workerService.StartTaskWatcher(ctx)

	// アサーション (非同期処理を待機しながらDBを確認)
	t.Log("Waiting for notification to be processed...")

	var notification models.Notification
	success := false

	// 最大15秒間、1秒おきにDBを確認する(ポーリング)
	for i := 0; i < 15; i++ {
		err := db.Where("user_id = ?", userID).First(&notification).Error
		if err == nil {
			success = true
			break
		}
		time.Sleep(1 * time.Second)
	}

	if !success {
		t.Error("タイムアウト: 通知がDBに保存されませんでした")
	}
}
