package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"my-portfolio-2025/internal/app/apperr"
	"my-portfolio-2025/internal/app/models"
	"my-portfolio-2025/internal/app/repository" // Repository層をインポート
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TaskServiceImpl は TaskService インターフェースの具体的な実装です。
type TaskServiceImpl struct {
	taskRepo      repository.TaskRepository // Repositoryへの依存性注入 (DI)
	workerService *WorkerService            // WorkerServiceへの依存性注入 (DI)
}

// NewTaskService は TaskService の新しいインスタンスを作成します。
func NewTaskService(repo repository.TaskRepository, workerService *WorkerService) TaskService {
	return &TaskServiceImpl{
		taskRepo:      repo,
		workerService: workerService,
	}
}

// CreateTask: タスク作成のビジネスロジック
func (s *TaskServiceImpl) CreateTask(userID uuid.UUID, req *models.TaskCreateRequest) (*models.Task, error) {

	// 入力バリデーション
	if req.Title == "" {
		return nil, fmt.Errorf("%w: title is required", apperr.ErrValidation)
	}

	task := &models.Task{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		DueDate:     req.DueDate,
		Status:      models.TaskStatusPending,
	}

	if err := s.taskRepo.Create(task); err != nil {
		return nil, fmt.Errorf("TaskService.CreateTask: %w", err)
	}
	return task, nil
}

// GetTaskByID: タスク詳細取得のビジネスロジックと認可チェック
func (s *TaskServiceImpl) GetTaskByID(userID uuid.UUID, taskID uuid.UUID) (*models.Task, error) {
	task, err := s.taskRepo.FindByID(taskID)
	if err != nil {
		// DBのNotFoundをapperr.ErrNotFoundに変換
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: taskID %s", apperr.ErrNotFound, taskID)
		}
		// その他のエラーはそのまま返す
		return nil, fmt.Errorf("TaskService.GetTaskByID: %w", err)
	}

	// 認可チェック: タスクの所有者か確認
	if task.UserID != userID {
		slog.Warn("Authorization violation attempt",
			"userID", userID,
			"taskID", taskID,
			"resourceType", "task",
			"action", "get",
			"ownerID", task.UserID,
		)
		return nil, fmt.Errorf("%w: user %s has no permission for task %s", apperr.ErrForbidden, userID, taskID)
	}

	return task, nil
}

// GetTasks: 特定のユーザーのタスクリストを取得。
func (s *TaskServiceImpl) GetTasks(userID uuid.UUID) ([]models.Task, error) {
	tasks, err := s.taskRepo.FindAllByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("TaskService.GetTasks: %w", err)
	}
	return tasks, nil
}

// UpdateTask: タスクの更新と認可チェック
func (s *TaskServiceImpl) UpdateTask(userID uuid.UUID, taskID uuid.UUID, req *models.TaskUpdateRequest) (*models.Task, error) {
	// GetTaskByIDを呼ぶことで、存在チェックと認可を一括で行う
	task, err := s.GetTaskByID(userID, taskID)
	if err != nil {
		return nil, err // apperr.ErrNotFound か apperr.ErrForbidden が返る
	}

	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.DueDate != nil {
		task.DueDate = *req.DueDate
	}
	if req.Status != nil {
		task.Status = *req.Status
	}

	if err := s.taskRepo.Update(task); err != nil {
		return nil, fmt.Errorf("TaskService.UpdateTask: %w", err)
	}

	return task, nil
}

// DeleteTask: タスクを削除します。認可チェックが必須です。
func (s *TaskServiceImpl) DeleteTask(userID uuid.UUID, taskID uuid.UUID) error {
	if _, err := s.GetTaskByID(userID, taskID); err != nil {
		return err
	}

	if err := s.taskRepo.Delete(taskID); err != nil {
		return fmt.Errorf("TaskService.DeleteTask: %w", err)
	}
	return nil
}

// CheckAndQueueDeadlines: 期限切れのタスクをチェックしてSQSにキューイングする
func (s *TaskServiceImpl) CheckAndQueueDeadlines(ctx context.Context) error {
	tasks, err := s.taskRepo.FindUpcomingTasks(ctx, time.Now().Add(1*time.Hour))
	if err != nil {
		return fmt.Errorf("TaskService.CheckAndQueueDeadlines: %w", err)
	}

	for _, task := range tasks {
		// ジョブ投入時のエラーは全体を止めないようログ出力（後にslogへ）
		if err := s.workerService.SendTaskNotification(ctx, task.ID, task.UserID, "期限が近づいています"); err != nil {
			fmt.Printf("failed to queue notification for task %s: %v\n", task.ID, err)
		}
	}
	return nil
}
