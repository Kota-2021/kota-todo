package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"my-portfolio-2025/internal/app/models"
	"my-portfolio-2025/internal/app/repository"
	"my-portfolio-2025/internal/infrastructure/aws"
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
		log.Printf("Failed to update last_notified_at for task %s: %v", taskID, err)
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
			output, err := s.sqsClient.Client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
				QueueUrl:            &s.sqsClient.QueueUrl,
				MaxNumberOfMessages: 1,
				WaitTimeSeconds:     20,
			})

			if err != nil {
				log.Printf("Failed to receive message: %v", err)
				continue
			}

			for _, msg := range output.Messages {
				log.Printf("Processing message: %s", *msg.Body)

				var notifyData models.NotificationMessage
				if err := json.Unmarshal([]byte(*msg.Body), &notifyData); err != nil {
					log.Printf("Failed to unmarshal SQS message: %v", err)
					continue
				}

				// --- ★【新規追加】DBへの永続化ロジック ---
				// SQSから届いたデータを元に、DB保存用のモデルを作成
				newNoti := &models.Notification{
					ID:        uuid.New(),
					UserID:    notifyData.UserID,
					Message:   notifyData.Message,
					Type:      "task_deadline", // 通知種別
					IsRead:    false,
					CreatedAt: time.Now(),
				}

				// Optional: TaskIDがメッセージに含まれている場合はセットする
				// newNoti.TaskID = &notifyData.TaskID (モデル側が対応していれば)

				// DBに保存
				if err := s.notiService.Create(ctx, newNoti); err != nil {
					log.Printf("Failed to save notification to DB: %v", err)
					// DB保存に失敗しても、リアルタイム通知は試みるか検討
				} else {
					// 保存に成功した場合、Hub経由で配信するメッセージにDB上のIDをセットする
					notifyData.ID = newNoti.ID
				}

				// --- WebSocket/Redis連携 ---
				// Redisを通じて全サーバーへブロードキャスト
				if err := s.hub.PublishMessage(ctx, notifyData); err != nil {
					log.Printf("Failed to publish to Redis: %v", err)
				}

				// メッセージを削除
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
					log.Printf("Failed to send notification for task %s: %v", task.ID, err)
				} else {
					log.Printf("Successfully queued notification for task %s", task.ID)
				}
			}
		}
	}
}
