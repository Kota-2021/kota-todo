package mock

import (
	"context"
	"my-portfolio-2025/internal/app/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// TaskServiceMock は service.TaskService インターフェースのモックです
type TaskServiceMock struct {
	mock.Mock
}

func (m *TaskServiceMock) CreateTask(userID uuid.UUID, req *models.TaskCreateRequest) (*models.Task, error) {
	args := m.Called(userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *TaskServiceMock) GetTasks(userID uuid.UUID) ([]models.Task, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Task), args.Error(1)
}

func (m *TaskServiceMock) GetTaskByID(userID, taskID uuid.UUID) (*models.Task, error) {
	args := m.Called(userID, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *TaskServiceMock) UpdateTask(userID, taskID uuid.UUID, req *models.TaskUpdateRequest) (*models.Task, error) {
	args := m.Called(userID, taskID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *TaskServiceMock) DeleteTask(userID, taskID uuid.UUID) error {
	args := m.Called(userID, taskID)
	return args.Error(0)
}

func (m *TaskServiceMock) CheckAndQueueDeadlines(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
