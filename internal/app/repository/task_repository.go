package repository

import (
	"my-portfolio-2025/internal/app/models" // Taskモデルをインポート
)

// TaskRepository はTaskモデルのデータ永続化（CRUD）操作を抽象化します。
type TaskRepository interface {
	// Create (作成)
	Create(task *models.Task) error

	// FindAllByUserID (リスト取得 - 認可チェックを含む)
	// 特定のユーザーIDに紐づく全てのタスクを取得
	FindAllByUserID(userID uint) ([]models.Task, error)

	// FindByID (詳細取得)
	FindByID(taskID uint) (*models.Task, error)

	// Update (更新)
	Update(task *models.Task) error

	// Delete (削除)
	Delete(taskID uint) error
}
