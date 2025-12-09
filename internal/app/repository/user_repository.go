package repository

import (
	"my-portfolio-2025/internal/app/models" // Taskモデルをインポート
)

// UserRepository はTaskモデルのデータ永続化（CRUD）操作を抽象化します。
type UserRepository interface {
	// CreateUser (作成)
	CreateUser(user *models.User) error

	// FindByUsername (ユーザー名からユーザーを取得)
	FindByUsername(username string) (*models.User, error)
}
