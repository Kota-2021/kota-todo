package models

import (
	"time"

	"gorm.io/gorm"
)

// タスクのステータスを定数で定義
const (
	TaskStatusPending    = "pending"     // 未着手
	TaskStatusInProgress = "in_progress" // 進行中
	TaskStatusCompleted  = "completed"   // 完了
)

// Task はタスクのデータベースレコードを表します。
type Task struct {
	gorm.Model // ID, CreatedAt, UpdatedAt, DeletedAt を自動で追加
	// 外部キー: どのユーザーがこのタスクを作成したか
	UserID         uint       `gorm:"not null" json:"user_id"`
	Title          string     `gorm:"type:varchar(255);not null" json:"title"`
	Description    string     `gorm:"type:text" json:"description"`
	DueDate        time.Time  `gorm:"type:timestamp" json:"due_date"`
	LastNotifiedAt *time.Time `json:"last_notified_at" gorm:"type:timestamp"` // ポインタにしてNULLを許容
	Status         string     `gorm:"type:varchar(20);not null;default:pending" json:"status"`
	CreatedAt      time.Time  `gorm:"type:timestamp" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"type:timestamp" json:"updated_at"`
}

// TaskCreateRequest は、タスク作成リクエストの入力データ構造です。
// クライアントからの入力を受け取るために使用します。
type TaskCreateRequest struct {
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
}

// TaskUpdateRequest は、タスク更新リクエストの入力データ構造です。
type TaskUpdateRequest struct {
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	DueDate     *time.Time `json:"due_date"`
	// IsCompleted *bool      `json:"is_completed"`
	Status *string `json:"status"`
}
