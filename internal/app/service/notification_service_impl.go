package service

import (
	"context"
	"my-portfolio-2025/internal/app/models"
	"my-portfolio-2025/internal/app/repository"

	"github.com/google/uuid"
)

type notificationServiceImpl struct {
	repo repository.NotificationRepository
}

func NewNotificationService(repo repository.NotificationRepository) NotificationService {
	return &notificationServiceImpl{repo: repo}
}

// GetNotifications はユーザーの通知を10件ずつ取得します
func (s *notificationServiceImpl) GetNotifications(ctx context.Context, userID uuid.UUID, page int) ([]models.Notification, error) {
	limit := 10
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	return s.repo.FindByUserID(ctx, userID, limit, offset)
}

// MarkAsRead は指定された通知を既読にします
func (s *notificationServiceImpl) MarkAsRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return s.repo.MarkAsRead(ctx, id, userID)
}
