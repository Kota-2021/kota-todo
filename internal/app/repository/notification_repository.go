package repository

import (
	"context"
	"my-portfolio-2025/internal/app/models"

	"github.com/google/uuid"
)

type NotificationRepository interface {

	// Create (作成)
	Create(ctx context.Context, notification *models.Notification) error

	// FindByUserID (ユーザーIDに紐づく通知を最新順に取得)
	FindByUserID(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]models.Notification, error)

	// MarkAsRead (既読に更新)
	MarkAsRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}
