package repository

import (
	"context"
	"fmt"
	"my-portfolio-2025/internal/app/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// taskRepositoryImpl は TaskRepository インターフェースの具体的な実装です。
type taskRepositoryImpl struct {
	db *gorm.DB // GormのDB接続インスタンス
}

// NewTaskRepository は TaskRepository の新しいインスタンスを作成します。
func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepositoryImpl{db: db}
}

// Create: 新しいタスクをDBに保存します。
// エラー発生時にコンテキストを付与して返す
func (r *taskRepositoryImpl) Create(task *models.Task) error {
	if err := r.db.Create(task).Error; err != nil {
		return fmt.Errorf("taskRepository.Create: %w", err)
	}
	return nil
}

// FindAllByUserID: 特定のユーザーIDに紐づく全てのタスクをリストで取得します。
func (r *taskRepositoryImpl) FindAllByUserID(userID uuid.UUID) ([]models.Task, error) {
	var tasks []models.Task
	// GORMのFindはレコードが見つからない場合に gorm.ErrRecordNotFound を返しません（空のスライスになる仕様）。
	// そのため、ここではDB接続エラー等の致命的なエラーのみをチェックします。
	if err := r.db.Where("user_id = ?", userID).Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("taskRepository.FindAllByUserID (userID=%s): %w", userID, err)
	}
	return tasks, nil
}

// FindByID: IDでタスクを検索します。
func (r *taskRepositoryImpl) FindByID(taskID uuid.UUID) (*models.Task, error) {
	var task models.Task
	// Firstはレコードが見つからない場合に gorm.ErrRecordNotFound を返します。
	if err := r.db.First(&task, taskID).Error; err != nil {
		return nil, fmt.Errorf("taskRepository.FindByID (taskID=%s): %w", taskID, err)
	}
	return &task, nil
}

// Update: Taskモデルの変更をDBに保存します。
func (r *taskRepositoryImpl) Update(task *models.Task) error {
	if err := r.db.Save(task).Error; err != nil {
		return fmt.Errorf("taskRepository.Update (taskID=%s): %w", task.ID, err)
	}
	return nil
}

// Delete: IDを指定してタスクを削除します。
func (r *taskRepositoryImpl) Delete(taskID uuid.UUID) error {
	if err := r.db.Delete(&models.Task{}, taskID).Error; err != nil {
		return fmt.Errorf("taskRepository.Delete (taskID=%s): %w", taskID, err)
	}
	return nil
}

// FindUpcomingTasks: 指定した日付より前の期限のタスクを取得 (期限切れチェック用)
func (r *taskRepositoryImpl) FindUpcomingTasks(ctx context.Context, threshold time.Time) ([]models.Task, error) {
	var tasks []models.Task
	now := time.Now()

	err := r.db.WithContext(ctx).
		Where("due_date <= ? AND status != ? AND (last_notified_at IS NULL OR last_notified_at < ?)",
			threshold, models.TaskStatusCompleted, now.Add(-1*time.Hour)).
		Find(&tasks).Error

	if err != nil {
		return nil, fmt.Errorf("taskRepository.FindUpcomingTasks: %w", err)
	}
	return tasks, nil
}

// UpdateLastNotifiedAt: 通知完了時刻を更新する
func (r *taskRepositoryImpl) UpdateLastNotifiedAt(ctx context.Context, taskID uuid.UUID, notifiedAt time.Time) error {
	err := r.db.WithContext(ctx).Model(&models.Task{}).
		Where("id = ?", taskID).
		Update("last_notified_at", notifiedAt).Error

	if err != nil {
		return fmt.Errorf("taskRepository.UpdateLastNotifiedAt (taskID=%s): %w", taskID, err)
	}
	return nil
}
