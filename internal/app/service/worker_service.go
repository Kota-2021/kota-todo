package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"my-portfolio-2025/internal/app/repository"
	"my-portfolio-2025/internal/infrastructure/aws"
	"time"

	awsgo "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	// "github.com/aws/aws-sdk-go-v2/aws" as awsgo
)

type WorkerService struct {
	// SQSクライアント
	sqsClient *aws.SQSClient
	// タスク管理リポジトリ
	taskRepo repository.TaskRepository
	// Hubを依存注入(DI)できるように追加
	hub *NotificationHub
}

func NewWorkerService(sqsClient *aws.SQSClient, taskRepo repository.TaskRepository, hub *NotificationHub) *WorkerService {
	return &WorkerService{sqsClient: sqsClient, taskRepo: taskRepo, hub: hub}
}

// SendTaskNotification はタスク情報をSQSに送信します
func (s *WorkerService) SendTaskNotification(ctx context.Context, taskID uint, userID uint, message string) error {
	body, _ := json.Marshal(map[string]interface{}{
		"task_id": taskID,
		"user_id": userID,
		"message": message,
	})

	// 1. SQSへメッセージを送信
	_, err := s.sqsClient.Client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    &s.sqsClient.QueueUrl,
		MessageBody: awsgo.String(string(body)),
	})
	if err != nil {
		return err
	}

	// 2. 送信成功後、DBの通知時刻を更新（二重送信防止の要）
	// リポジトリに追加した UpdateLastNotifiedAt を呼び出す
	err = s.taskRepo.UpdateLastNotifiedAt(ctx, taskID, time.Now())
	if err != nil {
		// ここでエラーになっても、SQSには飛んでいるのでログ出力に留めるのが一般的
		log.Printf("Failed to update last_notified_at for task %v: %v", taskID, err)
	}

	return nil
}

// StartWorker はGoルーチンで実行されるポーリングループです
func (s *WorkerService) StartWorker(ctx context.Context) {
	log.Println("SQS Worker started...")
	for {
		select {
		case <-ctx.Done():
			log.Println("Worker shutting down...")
			return
		default:
			// ロングポーリングでメッセージを受信
			output, err := s.sqsClient.Client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
				QueueUrl:            &s.sqsClient.QueueUrl,
				MaxNumberOfMessages: 1,
				WaitTimeSeconds:     20, // ロングポーリング
			})

			if err != nil {
				log.Printf("Failed to receive message: %v", err)
				continue
			}

			for _, msg := range output.Messages {
				log.Printf("Processing message: %s", *msg.Body)

				// --- WebSocket連携 ---
				// 1. SQSから届いたJSONを構造体にデコード
				var notifyData NotificationMessage
				if err := json.Unmarshal([]byte(*msg.Body), &notifyData); err != nil {
					log.Printf("Failed to unmarshal SQS message: %v", err)
					// 解析に失敗した場合は削除して良いか検討が必要。
					// 開発の現段階では一旦ログを出してスキップします
				} else {
					// 2. Hubのチャネルにメッセージを送信（ここでWebSocket配信が実行される）
					s.hub.Broadcast <- &notifyData
				}

				// 処理成功後にメッセージを削除（重要！）
				_, err := s.sqsClient.Client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
					QueueUrl:      &s.sqsClient.QueueUrl,
					ReceiptHandle: msg.ReceiptHandle,
				})
				if err != nil {
					log.Printf("Failed to delete message: %v", err)
				}
			}
		}
	}
}

// StartTaskWatcher はGoルーチンで実行されるタスク監視ループです
// 定期的にDBをチェックし、期限間近なタスクをSQSへ送ります
func (s *WorkerService) StartTaskWatcher(ctx context.Context) {
	// 1分ごとに実行するタイマーを設定
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	log.Println("Task Watcher (Producer) started...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Task Watcher shutting down...")
			return
		case <-ticker.C:
			// 1. 期限チェックの基準時間を設定（例：現在から1時間以内）
			threshold := time.Now().Add(1 * time.Hour)

			// 2. リポジトリを使って条件に合うタスクを取得
			// ここで前回修正した「二重送信防止クエリ」が活きます
			tasks, err := s.taskRepo.FindUpcomingTasks(ctx, threshold)
			if err != nil {
				log.Printf("Error finding upcoming tasks: %v", err)
				continue
			}

			// 3. 見つかったタスクを1つずつSQSに送信
			for _, task := range tasks {
				message := fmt.Sprintf("タスク「%s」の期限が近づいています（期限: %s）",
					task.Title, task.DueDate.Format("15:04"))

				// SendTaskNotification 内で SQS送信 ＋ DBの更新が行われる
				err := s.SendTaskNotification(ctx, task.ID, task.UserID, message)
				if err != nil {
					log.Printf("Failed to send notification for task %d: %v", task.ID, err)
				} else {
					log.Printf("Successfully queued notification for task %d", task.ID)
				}
			}
		}
	}
}
