package models

import (
	"time"

	"github.com/google/uuid"
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
	ID     uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	// UserID         uint       `gorm:"not null" json:"user_id"` // 260108byKota
	Title          string         `gorm:"type:varchar(255);not null" json:"title"`
	Description    string         `gorm:"type:text" json:"description"`
	DueDate        time.Time      `gorm:"type:timestamp" json:"due_date"`
	LastNotifiedAt *time.Time     `json:"last_notified_at" gorm:"type:timestamp"` // ポインタにしてNULLを許容
	Status         string         `gorm:"type:varchar(20);not null;default:pending" json:"status"`
	CreatedAt      time.Time      `gorm:"type:timestamp" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"type:timestamp" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
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

// BeforeCreate GORMのフックを使用して、作成時に自動でUUIDを付与する
func (t *Task) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return
}
