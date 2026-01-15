package service

import (
	"my-portfolio-2025/internal/app/apperr"
	"my-portfolio-2025/internal/app/models"
	"my-portfolio-2025/internal/testutils/mock"

	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	mockPkg "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// TaskTestSuite はタスクサービス (TaskService) のテストスイートです
type TaskTestSuite struct {
	suite.Suite
	mockTaskRepo *mock.MockTaskRepository
	taskService  TaskService // task_service.go で定義したインターフェース型
}

// SetupTest は各テストケースの前に実行されます
func (s *TaskTestSuite) SetupTest() {
	// 1. モックの初期化
	s.mockTaskRepo = new(mock.MockTaskRepository)
	// 2. サービスの実装にモックと設定を注入
	s.taskService = NewTaskService(s.mockTaskRepo, nil)
}

// TestTaskServiceSuite はテストスイートを実行します
func TestTaskServiceSuite(t *testing.T) {
	suite.Run(t, new(TaskTestSuite))
}

// テストケース: タスクCRUD機能（コアロジック）のユニットテスト実装

// 1.正常系テスト
// (1)CreateTaskテスト
func (s *TaskTestSuite) TestCreateTask_Success() {
	t := s.T()

	userID := uuid.New()
	req := &models.TaskCreateRequest{
		Title:       "Test Task",
		Description: "This is a test task",
		DueDate:     time.Now().Add(time.Hour * 24),
	}

	// Runを使って詳細な検証を行うのは、ログに「どのフィールドが違ったか」出すため
	s.mockTaskRepo.On("Create", mockPkg.AnythingOfType("*models.Task")).
		Return(nil).
		Run(func(args mockPkg.Arguments) {
			task := args.Get(0).(*models.Task)
			assert.Equal(t, userID, task.UserID, "UserID matching")
			assert.Equal(t, req.Title, task.Title, "Title matching")
		}).Once()

	task, err := s.taskService.CreateTask(userID, req)

	// メッセージを英語に統一（CI環境の標準に合わせる）または意味のある内容に
	assert.NoError(t, err, "Should not return error on valid input")
	assert.NotNil(t, task, "Returned task should not be nil")
	s.mockTaskRepo.AssertExpectations(t)
}

// (1)-2 空のタイトルでエラーを返すことを確認
func (s *TaskTestSuite) TestCreateTask_EmptyTitle() {
	t := s.T()
	req := &models.TaskCreateRequest{Title: ""} // 空のタイトル

	task, err := s.taskService.CreateTask(uuid.New(), req)

	assert.ErrorIs(t, err, apperr.ErrValidation)
	assert.Nil(t, task)
	// Repository は呼ばれないはず
	s.mockTaskRepo.AssertNotCalled(t, "Create", mockPkg.Anything)
}

// (2)GetTaskByIDテスト
func (s *TaskTestSuite) TestGetTaskByID_Success() {
	t := s.T()

	// 1. テストデータの準備
	task := &models.Task{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		Title:       "Test Task",
		Description: "This is a test task",
		DueDate:     time.Now().Add(time.Hour * 24),
		Status:      models.TaskStatusPending,
	}

	// 2. モックの期待値設定
	s.mockTaskRepo.On("FindByID", task.ID).Return(task, nil).Once()

	// 3. 実行と検証
	task, err := s.taskService.GetTaskByID(task.UserID, task.ID)

	// エラーがないことを検証
	assert.NoError(t, err)
	// taskオブジェクトがnilでないことを検証
	assert.NotNil(t, task)

	// 4. モックの呼び出し検証
	s.mockTaskRepo.AssertExpectations(t)
}

// (3)GetTasksテスト
func (s *TaskTestSuite) TestGetTasks_Success() {
	t := s.T()

	// 1. テストデータの準備
	task := &models.Task{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		Title:       "Test Task",
		Description: "This is a test task",
		DueDate:     time.Now().Add(time.Hour * 24),
		Status:      models.TaskStatusPending,
	}
	tasks := []models.Task{*task}

	// 2. モックの期待値設定
	s.mockTaskRepo.On("FindAllByUserID", task.UserID).Return(tasks, nil).Once()

	// 3. 実行と検証
	tasks, err := s.taskService.GetTasks(task.UserID)

	// エラーがないことを検証
	assert.NoError(t, err)
	// tasksオブジェクトがnilでないことを検証
	assert.NotNil(t, tasks)

	// 4. モックの呼び出し検証
	s.mockTaskRepo.AssertExpectations(t)

}

// (4)GetTasksテスト
// UserIDによるリクエストで、データが無かった場合は空のリストを返す事。を確認する。
func (s *TaskTestSuite) TestGetTasks_NoTasksFound() {
	t := s.T()
	requestingUserID := uuid.New()

	s.mockTaskRepo.On("FindAllByUserID", requestingUserID).Return([]models.Task{}, nil).Once()

	tasks, err := s.taskService.GetTasks(requestingUserID)

	assert.NoError(t, err)
	// assert.NotNil と assert.Len を分けるより、Emptyアサーションを使うとログが綺麗です
	assert.Empty(t, tasks, "Should return an empty slice if no tasks found")

	s.mockTaskRepo.AssertExpectations(t)
}

// (5)UpdateTaskテスト
func (s *TaskTestSuite) TestUpdateTask_Success() {
	t := s.T()

	// 1. テストデータの準備
	task := &models.Task{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		Title:       "Test Task",
		Description: "This is a test task",
		DueDate:     time.Now().Add(time.Hour * 24),
		Status:      models.TaskStatusPending,
	}
	title := "Updated Test Task"
	description := "This is an updated test task"
	dueDate := time.Now().Add(time.Hour * 24)
	status := models.TaskStatusInProgress
	req := &models.TaskUpdateRequest{
		Title:       &title,
		Description: &description,
		DueDate:     &dueDate,
		Status:      &status,
	}

	// 2. モックの期待値設定
	s.mockTaskRepo.On("FindByID", task.ID).Return(task, nil).Once()
	s.mockTaskRepo.On("Update", mockPkg.AnythingOfType("*models.Task")).Return(nil).Once()

	// 3. 実行と検証
	task, err := s.taskService.UpdateTask(task.UserID, task.ID, req)

	// エラーがないことを検証
	assert.NoError(t, err)
	// taskオブジェクトがnilでないことを検証
	assert.NotNil(t, task)

	// 4. モックの呼び出し検証
	s.mockTaskRepo.AssertExpectations(t)
}

// (6)DeleteTaskテスト
func (s *TaskTestSuite) TestDeleteTask_Success() {
	t := s.T()

	// 1. テストデータの準備
	task := &models.Task{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		Title:       "Test Task",
		Description: "This is a test task",
		DueDate:     time.Now().Add(time.Hour * 24),
		Status:      models.TaskStatusPending,
	}

	// 2. モックの期待値設定
	s.mockTaskRepo.On("FindByID", task.ID).Return(task, nil).Once()
	s.mockTaskRepo.On("Delete", task.ID).Return(nil).Once()

	// 3. 実行と検証
	err := s.taskService.DeleteTask(task.UserID, task.ID)

	// エラーがないことを検証
	assert.NoError(t, err)

	// 4. モックの呼び出し検証
	s.mockTaskRepo.AssertExpectations(t)
}

// 2.認可テスト(異常系)
// リクエストを行ったユーザーIDが、タスクのuser_idの不一致でエラーを返すことを確認

// (1)GetTaskByIDテスト
func (s *TaskTestSuite) TestGetTaskByID_Authorization() {
	t := s.T()

	// 1. テストデータの準備
	task := &models.Task{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		Title:       "Test Task",
		Description: "This is a test task",
		DueDate:     time.Now().Add(time.Hour * 24),
		Status:      models.TaskStatusPending,
	}
	unauthorizedUserID := uuid.New()

	// 2. モックの期待値設定
	s.mockTaskRepo.On("FindByID", task.ID).Return(task, nil).Once()

	// 3. 実行と検証
	task, err := s.taskService.GetTaskByID(unauthorizedUserID, task.ID)

	// エラーが発生していることを検証
	assert.Error(t, err)
	// taskオブジェクトがnilであることを検証
	assert.Nil(t, task)

	// 4. モックの呼び出し検証
	s.mockTaskRepo.AssertExpectations(t)
}

// (2)UpdateTaskテスト
func (s *TaskTestSuite) TestUpdateTask_Authorization() {
	t := s.T()
	taskOwnerID := uuid.New()
	unauthorizedUserID := uuid.New()
	task := &models.Task{
		ID: uuid.New(), UserID: taskOwnerID, Title: "Forbidden Task",
	}
	req := &models.TaskUpdateRequest{} // フィールドはnilでOK

	s.mockTaskRepo.On("FindByID", task.ID).Return(task, nil).Once()

	updatedTask, err := s.taskService.UpdateTask(unauthorizedUserID, task.ID, req)

	// 失敗ログに期待するエラーが含まれているか明示
	assert.Error(t, err, "Should return error for unauthorized user")
	assert.Contains(t, err.Error(), "forbidden", "Error message should contain 'forbidden'")
	assert.Nil(t, updatedTask, "Task should be nil on auth failure")

	s.mockTaskRepo.AssertExpectations(t)
	s.mockTaskRepo.AssertNotCalled(t, "Update", mockPkg.Anything)
}

// (3)DeleteTaskテスト
func (s *TaskTestSuite) TestDeleteTask_Authorization() {
	t := s.T()

	// 1. テストデータの準備
	taskOwnerID := uuid.New()        // タスクの所有者ID
	unauthorizedUserID := uuid.New() // 権限のないユーザーID
	task := &models.Task{
		ID:          uuid.New(),
		UserID:      taskOwnerID,
		Title:       "Test Task",
		Description: "This is a test task",
		DueDate:     time.Now().Add(time.Hour * 24),
		Status:      models.TaskStatusPending,
	}

	// 2. モックの期待値設定
	// 認可チェックのため、Service層はまずタスクをDBから取得する（FindByIDは呼ばれる）
	s.mockTaskRepo.On("FindByID", task.ID).Return(task, nil).Once()

	// 3. 実行と検証
	// 権限のないユーザーID(101)で削除を試みる
	err := s.taskService.DeleteTask(unauthorizedUserID, task.ID)

	// エラーが発生し、かつそれが認可エラーであることを検証
	assert.Error(t, err)
	// エラーメッセージに "forbidden" など認可失敗を示す文字列が含まれていることを検証
	assert.Contains(t, err.Error(), "forbidden")

	// 4. モックの呼び出し検証
	// FindByIDは呼ばれたが、Deleteは呼ばれなかったことを明示的に検証する
	s.mockTaskRepo.AssertExpectations(t)                          // FindByIDの呼び出しを検証
	s.mockTaskRepo.AssertNotCalled(t, "Delete", mockPkg.Anything) // DB削除が実行されていないことを保証
}
