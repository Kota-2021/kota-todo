package models

import (
	"time"

	"gorm.io/gorm"
)

// Task はタスクのデータベースレコードを表します。
type Task struct {
	gorm.Model // ID, CreatedAt, UpdatedAt, DeletedAt を自動で追加

	// 外部キー: どのユーザーがこのタスクを作成したか
	UserID uint `gorm:"not null" json:"user_id"`
	// User は belongs to User (Userモデルへのリレーション)
	// User User

	Title       string    `gorm:"type:varchar(255);not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	DueDate     time.Time `gorm:"type:timestamp" json:"due_date"`
	IsCompleted bool      `gorm:"default:false" json:"is_completed"`
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
	IsCompleted *bool      `json:"is_completed"`
}
