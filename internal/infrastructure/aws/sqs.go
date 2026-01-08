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

// internal/infrastructure/aws/sqs.go

func NewSQSClient(ctx context.Context, queueName string) (*SQSClient, error) {
	endpoint := os.Getenv("AWS_ENDPOINT")
	region := os.Getenv("AWS_REGION")

	staticProvider := credentials.NewStaticCredentialsProvider("test", "test", "")

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(staticProvider),
	)
	if err != nil {
		return nil, fmt.Errorf("AWS設定の読み込みに失敗しました: %w", err)
	}

	client := sqs.NewFromConfig(cfg, func(o *sqs.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
	})

	actualQueueUrl := "http://localhost:4566/000000000000/portfolio-notifications"
	fmt.Printf("✓ SQS Worker targeting: %s\n", actualQueueUrl)

	return &SQSClient{
		Client:   client,
		QueueUrl: actualQueueUrl,
	}, nil
}
