// internal/infrastructure/aws/sqs.go
package aws

import (
	"context"
	"fmt"
	"log/slog"
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
	// 環境変数の取得
	endpoint := os.Getenv("AWS_ENDPOINT")  // LocalStack用
	region := os.Getenv("AWS_REGION")      // 基本必須
	queueUrl := os.Getenv("SQS_QUEUE_URL") // Terraform/環境変数から渡される

	// 必須パラメータのチェック
	if queueUrl == "" {
		return nil, fmt.Errorf("SQS_QUEUE_URL is not set in environment variables")
	}

	// SDK設定のオプション構築
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
	}

	// LocalStack(endpointがある)の場合のモック認証設定
	if endpoint != "" {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("test", "test", ""),
		))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := sqs.NewFromConfig(cfg, func(o *sqs.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
	})

	// 構造化ログによる初期化完了の記録
	slog.Info("SQS client initialized",
		"queue_url", queueUrl,
		"region", region,
		"is_localstack", endpoint != "",
	)

	return &SQSClient{
		Client:   client,
		QueueUrl: queueUrl,
	}, nil
}
