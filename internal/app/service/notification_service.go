package service

import (
	"context"
	"my-portfolio-2025/internal/app/models"

	"github.com/google/uuid"
)

type NotificationService interface {

	// GetNotifications (ユーザーの通知を10件ずつ取得)
	GetNotifications(ctx context.Context, userID uuid.UUID, page int) ([]models.Notification, error)

	// MarkAsRead (指定された通知を既読にします)
	MarkAsRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	// 今後、SQSワーカーから呼ばれる「保存用メソッド」もここに追加予定です

}
