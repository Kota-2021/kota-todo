package service

import (
	"context"
	"os"
	"testing"

	"my-portfolio-2025/internal/app/models"
	"my-portfolio-2025/internal/infrastructure/aws"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupTestDB はテスト用のDB接続を初期化し、テーブルをクリーンアップします
func SetupTestDB(t *testing.T) *gorm.DB {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// デフォルトのテスト用接続先（docker-composeの設定に合わせる）
		dsn = "host=localhost user=portfolio_admin password=local_dev_password dbname=portfolio_db port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("テストDBへの接続に失敗しました: %v", err)
	}

	err = db.AutoMigrate(&models.Task{}, &models.Notification{}, &models.User{})
	if err != nil {
		t.Fatalf("マイグレーションに失敗しました: %v", err)
	}

	// テスト前にNotificationsテーブルを空にする（クリーンな状態でのテスト）
	db.Exec("DELETE FROM notifications")
	db.Exec("DELETE FROM tasks")

	return db
}

// SetupTestSQS はテスト用のSQSクライアントを初期化します
func SetupTestSQS(t *testing.T) *aws.SQSClient {
	ctx := context.Background()
	queueName := "portfolio-notifications" // テスト用キュー名

	// 先ほど修正したNewSQSClientを利用
	client, err := aws.NewSQSClient(ctx, queueName)
	if err != nil {
		t.Fatalf("テスト用SQSの初期化に失敗しました: %v", err)
	}

	return client
}
