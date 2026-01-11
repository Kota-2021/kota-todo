package service

import (
	"context"
	"errors"
	"fmt"
	"my-portfolio-2025/internal/app/apperr"
	"my-portfolio-2025/internal/app/models"
	"my-portfolio-2025/internal/app/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
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

	notifications, err := s.repo.FindByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("notificationService.GetNotifications: %w", err)
	}

	return notifications, nil
}

// MarkAsRead は指定された通知を既読にします
func (s *notificationServiceImpl) MarkAsRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	err := s.repo.MarkAsRead(ctx, id, userID)
	if err != nil {
		// Repository側で「ID不一致 または UserID不一致」の場合に ErrRecordNotFound が返る設計
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 通知が存在しないか、自分のものではないため NotFoundとして扱う
			return fmt.Errorf("%w: notificationID %s for userID %s", apperr.ErrNotFound, id, userID)
		}
		// その他のエラーはそのまま返す
		return fmt.Errorf("notificationService.MarkAsRead: %w", err)
	}
	return nil
}

// Create は通知を作成します
func (s *notificationServiceImpl) Create(ctx context.Context, notification *models.Notification) error {
	if notification.UserID == uuid.Nil {
		return fmt.Errorf("%w: userID is required for notification", apperr.ErrValidation)
	}

	if err := s.repo.Create(ctx, notification); err != nil {
		return fmt.Errorf("notificationService.Create: %w", err)
	}
	return nil
}
