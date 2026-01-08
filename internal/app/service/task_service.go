// internal/app/service/task_service.go
package service

import (
	"context"
	"my-portfolio-2025/internal/app/models"

	"github.com/google/uuid"
)

// TaskService はタスクに関するビジネスロジックを定義します。
type TaskService interface {
	// CreateTask: 新しいタスクを作成。UserIDを必須とする。
	CreateTask(userID uuid.UUID, req *models.TaskCreateRequest) (*models.Task, error)

	// GetTasks: 特定のユーザーのタスクリストを取得。
	GetTasks(userID uuid.UUID) ([]models.Task, error)

	// GetTaskByID: 特定のタスクを取得。**認可チェック**のためにUserIDとTaskIDの両方を受け取る。
	GetTaskByID(userID uuid.UUID, taskID uuid.UUID) (*models.Task, error)

	// UpdateTask: タスクを更新。認可チェックのためにUserIDとTaskIDを受け取る。
	UpdateTask(userID uuid.UUID, taskID uuid.UUID, req *models.TaskUpdateRequest) (*models.Task, error)

	// DeleteTask: タスクを削除。認可チェックのためにUserIDとTaskIDを受け取る。
	DeleteTask(userID uuid.UUID, taskID uuid.UUID) error

	// CheckAndQueueDeadlines: 期限切れのタスクをチェックしてSQSにキューイングする
	CheckAndQueueDeadlines(ctx context.Context) error
}
