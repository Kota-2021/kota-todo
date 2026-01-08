package service

import (
	"context"
	"my-portfolio-2025/internal/app/models"

	"github.com/google/uuid"
)

type NotificationService interface {

	// Create は新しい通知をDBに保存します (WorkerServiceから呼ばれます)
	Create(ctx context.Context, notification *models.Notification) error

	// GetNotifications はユーザーの通知をページネーション付きで取得します
	GetNotifications(ctx context.Context, userID uuid.UUID, page int) ([]models.Notification, error)

	// MarkAsRead は指定された通知を既読にします
	MarkAsRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}
