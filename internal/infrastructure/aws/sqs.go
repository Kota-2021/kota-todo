package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// SQSClient はSQS操作のラッパー構造体です
type SQSClient struct {
	Client   *sqs.Client
	QueueUrl string
}

// NewSQSClient はSQSクライアントを初期化して返します
func NewSQSClient(ctx context.Context, queueName string) (*SQSClient, error) {

	// aws config の読み込み
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("AWS設定の読み込みに失敗しました: %w", err)
	}

	// aws sqs client の初期化
	client := sqs.NewFromConfig(cfg)

	// キュー名からURLを取得
	result, err := client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: &queueName,
	})
	if err != nil {
		return nil, fmt.Errorf("キューURLの取得に失敗しました: %w", err)
	}

	// SQSClient を返す
	return &SQSClient{
		Client:   client,
		QueueUrl: *result.QueueUrl,
	}, nil
}
