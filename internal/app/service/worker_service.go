package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"my-portfolio-2025/internal/app/models"
	"my-portfolio-2025/internal/app/repository"
	"my-portfolio-2025/internal/infrastructure/aws"
	"my-portfolio-2025/pkg/utils"
	"time"

	awsgo "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
	// "github.com/aws/aws-sdk-go-v2/aws" as awsgo
)

type WorkerService struct {
	// SQSクライアント
	sqsClient *aws.SQSClient
	// タスク管理リポジトリ
	taskRepo repository.TaskRepository
	// 通知管理機能
	notiService NotificationService
	// Hubを依存注入(DI)できるように追加
	hub *NotificationHub
}

func NewWorkerService(sqsClient *aws.SQSClient, taskRepo repository.TaskRepository, notiService NotificationService, hub *NotificationHub) *WorkerService {
	return &WorkerService{sqsClient: sqsClient, taskRepo: taskRepo, notiService: notiService, hub: hub}
}

// SendTaskNotification はタスク情報をSQSに送信します
func (s *WorkerService) SendTaskNotification(ctx context.Context, taskID uuid.UUID, userID uuid.UUID, message string) error {
	body, err := json.Marshal(map[string]interface{}{
		"task_id": taskID,
		"user_id": userID,
		"message": message,
	})
	if err != nil {
		return fmt.Errorf("WorkerService.SendTaskNotification (marshal): %w", err)
	}

	// 1. SQSへメッセージを送信
	_, err = s.sqsClient.Client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    &s.sqsClient.QueueUrl,
		MessageBody: awsgo.String(string(body)),
	})
	if err != nil {
		return fmt.Errorf("WorkerService.SendTaskNotification (SQS): %w", err)
	}

	// 2. 送信成功後、DBの通知時刻を更新
	if err := s.taskRepo.UpdateLastNotifiedAt(ctx, taskID, utils.NowJST()); err != nil {
		// SQSには送信済みのため、エラーを返さずログに留める（構造化ログの活用）
		slog.Error("Failed to update last_notified_at in DB",
			"taskID", taskID,
			"error", err,
		)
	}

	return nil
}

// StartWorker はGoルーチンで実行されるポーリングループです
func (s *WorkerService) StartWorker(ctx context.Context) {
	slog.Info("SQS Worker started")
	for {
		select {
		case <-ctx.Done():
			slog.Info("Worker shutting down")
			return
		default:
			output, err := s.sqsClient.Client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
				QueueUrl:            &s.sqsClient.QueueUrl,
				MaxNumberOfMessages: 1,
				WaitTimeSeconds:     20,
			})

			if err != nil {
				// 【修正ポイント】ctx.Done()によるエラーなら、エラーログを出さずに終了する
				if ctx.Err() != nil {
					slog.Info("Worker loop stopped by context cancellation")
					return
				}
				slog.Error("Failed to receive message from SQS", "error", err)

				// エラー時の待機中もキャンセルを検知できるようにする
				select {
				case <-ctx.Done():
					return
				case <-time.After(5 * time.Second):
					continue
				}
			}

			for _, msg := range output.Messages {
				slog.Debug("Processing SQS message", "body", *msg.Body)

				var notifyData models.NotificationMessage
				if err := json.Unmarshal([]byte(*msg.Body), &notifyData); err != nil {
					slog.Error("Failed to unmarshal SQS message", "error", err)
					continue
				}

				newNoti := &models.Notification{
					ID:        uuid.New(),
					UserID:    notifyData.UserID,
					Message:   notifyData.Message,
					Type:      "task_deadline",
					IsRead:    false,
					CreatedAt: utils.NowJST(),
				}

				// DBに保存
				if err := s.notiService.Create(ctx, newNoti); err != nil {
					slog.Error("Failed to save notification to DB", "error", err)
				} else {
					notifyData.ID = newNoti.ID
				}

				// WebSocket/Redis経由でブロードキャスト
				if err := s.hub.PublishMessage(ctx, notifyData); err != nil {
					slog.Error("Failed to publish to Redis", "error", err)
				}

				// 処理完了したメッセージを削除
				s.sqsClient.Client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
					QueueUrl:      &s.sqsClient.QueueUrl,
					ReceiptHandle: msg.ReceiptHandle,
				})
			}
		}
	}
}

// StartTaskWatcher はGoルーチンで実行されるタスク監視ループです
// 定期的にDBをチェックし、期限間近なタスクをSQSへ送ります
func (s *WorkerService) StartTaskWatcher(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	slog.Info("Task Watcher (Producer) started")

	// 共通の処理ロジック
	runWatcher := func() {
		threshold := utils.NowJST().Add(1 * time.Hour)
		tasks, err := s.taskRepo.FindUpcomingTasks(ctx, threshold)
		if err != nil {
			slog.Error("Error finding upcoming tasks", "error", err)
			return
		}

		for _, task := range tasks {
			message := fmt.Sprintf("タスク「%s」の期限が近づいています（期限: %s）",
				task.Title, task.DueDate.In(utils.JST).Format("15:04"))

			if err := s.SendTaskNotification(ctx, task.ID, task.UserID, message); err != nil {
				slog.Error("Failed to send notification", "taskID", task.ID, "error", err)
			} else {
				slog.Info("Successfully queued notification", "taskID", task.ID)
			}
		}
	}

	runWatcher() // 初回実行

	for {
		select {
		case <-ctx.Done():
			slog.Info("Task Watcher shutting down")
			return
		case <-ticker.C:
			runWatcher()
		}
	}
}
