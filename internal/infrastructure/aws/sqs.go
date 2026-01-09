// internal/infrastructure/aws/sqs.go
package aws

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type SQSClient struct {
	Client   *sqs.Client
	QueueUrl string
}

func NewSQSClient(ctx context.Context, queueName string) (*SQSClient, error) {

	// ローカル開発用
	endpoint := os.Getenv("AWS_ENDPOINT")
	region := os.Getenv("AWS_REGION")

	// Terraformから渡されるURL
	queueUrl := os.Getenv("SQS_QUEUE_URL")

	// オプションをスライスで管理
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
	}

	// LocalStack(endpointがある)の場合だけ、偽の認証情報を使う
	if endpoint != "" {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("test", "test", ""),
		))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("AWS設定の読み込みに失敗しました: %w", err)
	}

	client := sqs.NewFromConfig(cfg, func(o *sqs.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
	})

	// デバッグ用ログ
	fmt.Printf("✓ SQS client initialized. Target Queue: %s\n", queueUrl)

	return &SQSClient{
		Client:   client,
		QueueUrl: queueUrl,
	}, nil

}
