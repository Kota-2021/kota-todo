package repository

import (
	"context"
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
	return r.db.WithContext(ctx).Create(n).Error
}

// FindByUserID は特定のユーザーの通知を最新順に取得します（ページネーション対応）
func (r *notificationRepositoryImpl) FindByUserID(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
	return notifications, err
}

// MarkAsRead は通知を既読(is_read = true)に更新します
// セキュリティのため、userIDも条件に含めて他人の通知を操作できないようにします
func (r *notificationRepositoryImpl) MarkAsRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
