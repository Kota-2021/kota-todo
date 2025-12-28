// internal/testutils/mock/task_mock.go
package mock

import (
	"context"
	"my-portfolio-2025/internal/app/models" // モデルパッケージへのパスは適宜修正してください
	"time"

	"github.com/stretchr/testify/mock"
)

// MockTaskRepository は repository.TaskRepository インターフェースのモックです
type MockTaskRepository struct {
	mock.Mock
}

// CreateTask は TaskRepository.CreateTask のモック実装です
func (m *MockTaskRepository) Create(task *models.Task) error {
	args := m.Called(task)

	// var createdTask *models.Task
	// if args.Get(0) != nil {
	// 	createdTask = args.Get(0).(*models.Task)
	// }

	return args.Error(0)
}

// FindAllByUserID は TaskRepository.FindAllByUserID のモック実装です
func (m *MockTaskRepository) FindAllByUserID(userID uint) ([]models.Task, error) {
	args := m.Called(userID)

	var tasks []models.Task
	if args.Get(0) != nil {
		tasks = args.Get(0).([]models.Task)
	}

	return tasks, args.Error(1)
}

// FindByID は TaskRepository.FindByID のモック実装です
func (m *MockTaskRepository) FindByID(taskID uint) (*models.Task, error) {
	args := m.Called(taskID)

	var task *models.Task
	if args.Get(0) != nil {
		task = args.Get(0).(*models.Task)
	}

	return task, args.Error(1)
}

// Update は TaskRepository.Update のモック実装です
func (m *MockTaskRepository) Update(task *models.Task) error {
	// Mockオブジェクトに設定された期待値（引数と戻り値）に基づいて処理を実行します
	args := m.Called(task)

	// var user *models.User
	// if args.Get(0) != nil {
	// 	user = args.Get(0).(*models.User)
	// }

	return args.Error(0)
}

// Delete は TaskRepository.Delete のモック実装です
func (m *MockTaskRepository) Delete(taskID uint) error {
	// Mockオブジェクトに設定された期待値（引数と戻り値）に基づいて処理を実行します
	args := m.Called(taskID)

	// var user *models.User
	// if args.Get(0) != nil {
	// 	user = args.Get(0).(*models.User)
	// }

	return args.Error(0)
}

// MockTaskRepository 構造体にメソッドを追加
func (m *MockTaskRepository) FindUpcomingTasks(ctx context.Context, threshold time.Time) ([]models.Task, error) {
	args := m.Called(ctx, threshold)

	// args.Get(0) を []models.Task 型にキャストして返す
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Task), args.Error(1)
}

// UpdateLastNotifiedAt は TaskRepository.UpdateLastNotifiedAt のモック実装です
func (m *MockTaskRepository) UpdateLastNotifiedAt(ctx context.Context, taskID uint, notifiedAt time.Time) error {
	// テスト時に呼び出しを記録したり、エラーをシミュレートしたりするためのロジック
	// testify/mock を使っている場合は以下のように記述
	args := m.Called(ctx, taskID, notifiedAt)
	return args.Error(0)
}
