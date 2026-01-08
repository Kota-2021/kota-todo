package models

import (
	"time"

	"github.com/google/uuid"
)

// Notification は通知情報を表すモデルです
// DB保存用モデル
type Notification struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	TaskID    *uuid.UUID `gorm:"type:uuid" json:"task_id"`     // システム通知の場合はnullを許容
	Type      string     `gorm:"size:20;not null" json:"type"` // overdue, system など
	Message   string     `gorm:"type:text;not null" json:"message"`
	IsRead    bool       `gorm:"not null;default:false" json:"is_read"`
	CreatedAt time.Time  `gorm:"not null" json:"created_at"`
}

// 配信（WebSocket/Redis）用
type NotificationMessage struct {
	ID      uuid.UUID `json:"id"`
	UserID  uuid.UUID `json:"user_id"`
	Type    string    `json:"type"`
	Message string    `json:"message"`
}
