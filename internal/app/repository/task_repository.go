package repository

import (
	"context"
	"my-portfolio-2025/internal/app/models" // Taskモデルをインポート
	"time"
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

	// FindUpcomingTasks: 指定した日付より前の期限のタスクを取得 (期限切れチェック用)
	FindUpcomingTasks(ctx context.Context, threshold time.Time) ([]models.Task, error)
}
