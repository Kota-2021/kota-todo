package service

import (
	"context"
	"encoding/json"
	"log"
	"my-portfolio-2025/internal/infrastructure/aws"

	awsgo "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	// "github.com/aws/aws-sdk-go-v2/aws" as awsgo
)

type WorkerService struct {
	sqsClient *aws.SQSClient
}

func NewWorkerService(sqsClient *aws.SQSClient) *WorkerService {
	return &WorkerService{sqsClient: sqsClient}
}

// SendTaskNotification はタスク情報をSQSに送信します
func (s *WorkerService) SendTaskNotification(ctx context.Context, taskID uint, message string) error {
	body, _ := json.Marshal(map[string]interface{}{
		"task_id": taskID,
		"message": message,
	})

	_, err := s.sqsClient.Client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    &s.sqsClient.QueueUrl,
		MessageBody: awsgo.String(string(body)),
	})
	return err
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
				// ここで通知処理（後のW4-D17でのWebSocket連携等）を呼び出す
				log.Printf("Processing message: %s", *msg.Body)

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
