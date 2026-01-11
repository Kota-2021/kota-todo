package repository

import (
	"context"
	"fmt"
	"my-portfolio-2025/internal/app/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type notificationRepositoryImpl struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepositoryImpl{db: db}
}

// Create は新しい通知をDBに保存します
func (r *notificationRepositoryImpl) Create(ctx context.Context, n *models.Notification) error {
	if err := r.db.WithContext(ctx).Create(n).Error; err != nil {
		return fmt.Errorf("notificationRepository.Create: %w", err)
	}
	return nil
}

// FindByUserID は特定のユーザーの通知を最新順に取得します
func (r *notificationRepositoryImpl) FindByUserID(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error

	if err != nil {
		return nil, fmt.Errorf("notificationRepository.FindByUserID (userID=%s): %w", userID, err)
	}
	return notifications, nil
}

// MarkAsRead は通知を既読(is_read = true)に更新します
// セキュリティのため、userIDも条件に含めて他人の通知を操作できないようにする
func (r *notificationRepositoryImpl) MarkAsRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true)

	if result.Error != nil {
		return fmt.Errorf("notificationRepository.MarkAsRead (id=%s, userID=%s): %w", id, userID, result.Error)
	}

	// 更新対象がなかった場合（ID間違いや他人の通知）はNotFoundとして扱う
	if result.RowsAffected == 0 {
		return fmt.Errorf("notificationRepository.MarkAsRead (id=%s): %w", id, gorm.ErrRecordNotFound)
	}
	return nil
}
