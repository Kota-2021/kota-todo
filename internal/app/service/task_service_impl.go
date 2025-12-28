package service

import (
	"context"
	"errors"
	"my-portfolio-2025/internal/app/models"
	"my-portfolio-2025/internal/app/repository" // Repository層をインポート
	"time"
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
func (s *TaskServiceImpl) CreateTask(userID uint, req *models.TaskCreateRequest) (*models.Task, error) {
	// 1. DTOからModelへの変換とUserIDのセット
	task := &models.Task{
		UserID:      userID, // JWTミドルウェアから渡されたUserIDを設定
		Title:       req.Title,
		Description: req.Description,
		DueDate:     req.DueDate,
		Status:      models.TaskStatusPending,
	}

	// 2. Repositoryを呼び出し、DBに保存
	err := s.taskRepo.Create(task)
	if err != nil {
		return nil, err
	}
	return task, nil
}

// GetTaskByID: タスク詳細取得のビジネスロジックと認可チェック
func (s *TaskServiceImpl) GetTaskByID(userID uint, taskID uint) (*models.Task, error) {
	// 1. Repositoryを呼び出し、タスクを検索
	task, err := s.taskRepo.FindByID(taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("task not found") // タスクが存在しない
	}

	// 2. 認可チェック: タスクの所有者か確認（最も重要！）
	if task.UserID != userID {
		// ログ出力: 認可違反の試行
		// エラーを返すことで、タスク詳細の取得を拒否します
		return nil, errors.New("forbidden: task does not belong to user")
	}

	return task, nil
}

// GetTasks: 特定のユーザーのタスクリストを取得。
func (s *TaskServiceImpl) GetTasks(userID uint) ([]models.Task, error) {
	// 認可チェックはRepository層のFindAllByUserIDに委譲（UserIDでフィルタリングされるため）
	// Service層の役割はシンプルにRepositoryを呼び出すこと
	tasks, err := s.taskRepo.FindAllByUserID(userID)
	if err != nil {
		return nil, err
	}

	// 必要に応じてここでタスクのソートや整形などのビジネスロジックを適用

	return tasks, nil
}

// UpdateTask: タスクの更新と認可チェック
func (s *TaskServiceImpl) UpdateTask(userID uint, taskID uint, req *models.TaskUpdateRequest) (*models.Task, error) {
	// 1. タスクの存在確認と認可チェック（GetTaskByIDと同様の処理が必要）
	task, err := s.taskRepo.FindByID(taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("task not found")
	}

	// 2. 認可チェック: タスクの所有者か確認（最重要！）
	if task.UserID != userID {
		return nil, errors.New("forbidden: task does not belong to user")
	}

	// 3. 更新データの適用: リクエストがあったフィールドのみをモデルに適用（Go特有の処理）
	// DTOのフィールドがnilでない場合（つまりリクエストに含まれていた場合）のみ、モデルの値を更新します。
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

	// 4. Repositoryを呼び出し、DBを更新
	err = s.taskRepo.Update(task) // taskオブジェクト全体を渡して更新
	if err != nil {
		return nil, err
	}

	return task, nil // 更新されたタスクを返す
}

// DeleteTask: タスクを削除します。認可チェックが必須です。
func (s *TaskServiceImpl) DeleteTask(userID uint, taskID uint) error {
	// 1. 認可チェックのためにタスクを取得
	task, err := s.taskRepo.FindByID(taskID)
	if err != nil {
		return err // DBエラー
	}
	if task == nil {
		return errors.New("task not found")
	}

	// 2. 認可チェック: タスクの所有者か確認（最重要！）
	if task.UserID != userID {
		// 他人のタスクは削除できない
		return errors.New("forbidden: task does not belong to user")
	}

	// 3. 認可OK: Repositoryを呼び出して削除を実行
	return s.taskRepo.Delete(taskID)
}

// CheckAndQueueDeadlines: 期限切れのタスクをチェックしてSQSにキューイングする
func (s *TaskServiceImpl) CheckAndQueueDeadlines(ctx context.Context) error {
	// 1時間以内に期限が来るタスクを取得する
	tasks, err := s.taskRepo.FindUpcomingTasks(ctx, time.Now().Add(1*time.Hour))
	if err != nil {
		return err
	}

	for _, task := range tasks {
		// SQSにジョブを投入
		s.workerService.SendTaskNotification(ctx, task.ID, "期限が近づいています")
	}
	return nil
}
