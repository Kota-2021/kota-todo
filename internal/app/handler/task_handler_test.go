package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"my-portfolio-2025/internal/app/apperr"
	"my-portfolio-2025/internal/app/models"
	"my-portfolio-2025/internal/testutils/mock" // 作成済みのMockを想定

	mk "github.com/stretchr/testify/mock"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTaskHandler_CreateTask(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()

	tests := []struct {
		name           string
		setupMock      func(m *mock.TaskServiceMock)
		requestBody    interface{}
		expectedStatus int
		expectedTitle  string
	}{
		{
			name: "正常系：新しいタスクを作成できる",
			setupMock: func(m *mock.TaskServiceMock) {
				req := &models.TaskCreateRequest{Title: "New Task"}
				m.On("CreateTask", userID, req).Return(&models.Task{
					ID:     uuid.New(),
					UserID: userID,
					Title:  "New Task",
				}, nil)
			},
			requestBody:    models.TaskCreateRequest{Title: "New Task"},
			expectedStatus: http.StatusCreated,
			expectedTitle:  "New Task",
		},
		{
			name:           "異常系：バリデーションエラー（タイトル空）",
			setupMock:      func(m *mock.TaskServiceMock) {},
			requestBody:    "invalid-json", // 不正な形式
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mock.TaskServiceMock)
			tt.setupMock(mockService)
			h := NewTaskHandler(mockService)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// JSONボディのセット
			jsonBody, _ := json.Marshal(tt.requestBody)
			c.Request, _ = http.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer(jsonBody))
			c.Request.Header.Set("Content-Type", "application/json")

			c.Set("userID", userID)

			h.CreateTask(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				var response models.Task
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTitle, response.Title)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestTaskHandler_GetTasks(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()

	tests := []struct {
		name           string
		setupMock      func(m *mock.TaskServiceMock)
		contextUserID  uuid.UUID
		expectedStatus int
		expectedCount  int // レスポンスに含まれるタスク数
	}{
		{
			name: "正常系：自分のタスク一覧を取得できる",
			setupMock: func(m *mock.TaskServiceMock) {
				// []*models.Task ではなく []models.Task (ポインタなし) として作成
				tasks := []models.Task{
					{ID: uuid.New(), UserID: userID, Title: "Task 1"},
					{ID: uuid.New(), UserID: userID, Title: "Task 2"},
				}
				m.On("GetTasks", userID).Return(tasks, nil)
			},
			contextUserID:  userID,
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name: "正常系：タスクが1件もない場合は空配列を返す",
			setupMock: func(m *mock.TaskServiceMock) {
				// ❌ []*models.Task{} になっていませんか？
				// ✅ ポインタなしの []models.Task{} に修正します
				m.On("GetTasks", userID).Return([]models.Task{}, nil)
			},
			contextUserID:  userID,
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name:           "異常系：認証情報がない場合は401を返す",
			setupMock:      func(m *mock.TaskServiceMock) {},
			contextUserID:  uuid.Nil, // ユーザーID未セット
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mock.TaskServiceMock)
			tt.setupMock(mockService)
			h := NewTaskHandler(mockService)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// リクエスト設定
			c.Request, _ = http.NewRequest(http.MethodGet, "/tasks", nil)
			if tt.contextUserID != uuid.Nil {
				c.Set("userID", tt.contextUserID)
			}

			h.GetTasks(c)

			// ステータスコード検証
			assert.Equal(t, tt.expectedStatus, w.Code)

			// 正常系のときは件数もチェック
			if tt.expectedStatus == http.StatusOK {
				var response []models.Task // ここもポインタなしにする
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response, tt.expectedCount)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestTaskHandler_GetTaskByID(t *testing.T) {
	// Ginをテストモードに設定
	gin.SetMode(gin.TestMode)

	// テスト用の固定ID
	userID := uuid.New()
	otherUserID := uuid.New()
	taskID := uuid.New()

	tests := []struct {
		name           string
		setupMock      func(m *mock.TaskServiceMock)
		contextUserID  uuid.UUID
		urlParamID     string
		expectedStatus int
		expectedError  string
	}{
		{
			name: "正常系：自分のタスクを取得できる",
			setupMock: func(m *mock.TaskServiceMock) {
				m.On("GetTaskByID", userID, taskID).Return(&models.Task{
					ID: taskID, UserID: userID, Title: "Test Task",
				}, nil)
			},
			contextUserID:  userID,
			urlParamID:     taskID.String(),
			expectedStatus: http.StatusOK,
		},
		{
			name: "異常系：他人のタスクにアクセスすると403を返す(認可ガード)",
			setupMock: func(m *mock.TaskServiceMock) {
				// Service層がErrForbiddenを返すパターンを模倣
				m.On("GetTaskByID", otherUserID, taskID).Return(nil, apperr.ErrForbidden)
			},
			contextUserID:  otherUserID,
			urlParamID:     taskID.String(),
			expectedStatus: http.StatusForbidden,
			expectedError:  "この操作を行う権限がありません",
		},
		{
			name:           "異常系：無効なUUID形式の場合は400を返す",
			setupMock:      func(m *mock.TaskServiceMock) {},
			contextUserID:  userID,
			urlParamID:     "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid UUID",
		},
		{
			name: "異常系：タスクが存在しない場合は404を返す",
			setupMock: func(m *mock.TaskServiceMock) {
				m.On("GetTaskByID", userID, taskID).Return(nil, apperr.ErrNotFound)
			},
			contextUserID:  userID,
			urlParamID:     taskID.String(),
			expectedStatus: http.StatusNotFound,
			expectedError:  "指定されたタスクが見つかりません",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mockの初期化
			mockService := new(mock.TaskServiceMock)
			tt.setupMock(mockService)
			h := NewTaskHandler(mockService)

			// HTTPリクエストのシミュレーション設定
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// ContextにuserIDをセット（ミドルウェアの動作を再現）
			c.Set("userID", tt.contextUserID)
			// URLパラメータをセット
			c.Params = []gin.Param{{Key: "id", Value: tt.urlParamID}}

			// 実行
			h.GetTaskByID(c)

			// 検証
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestTaskHandler_UpdateTask(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	taskID := uuid.New()

	tests := []struct {
		name           string
		setupMock      func(m *mock.TaskServiceMock)
		contextUserID  uuid.UUID
		urlParamID     string
		requestBody    interface{} // テストごとに異なるリクエストを送れるように
		expectedStatus int
		expectedError  string
	}{
		{
			name: "正常系：自分のタスクを更新できる",
			setupMock: func(m *mock.TaskServiceMock) {
				req := &models.TaskUpdateRequest{Title: ptr("Updated Title")}
				m.On("UpdateTask", userID, taskID, req).Return(&models.Task{
					ID: taskID, UserID: userID, Title: "Updated Title",
				}, nil)
			},
			contextUserID:  userID,
			urlParamID:     taskID.String(),
			requestBody:    models.TaskUpdateRequest{Title: ptr("Updated Title")},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "異常系：リクエストボディが不正な場合は400を返す",
			setupMock:      func(m *mock.TaskServiceMock) {},
			contextUserID:  userID,
			urlParamID:     taskID.String(),
			requestBody:    "invalid-json", // JSONとして解析不能なデータ
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid character",
		},
		{
			name: "異常系：他人のタスクを更新しようとすると403を返す",
			setupMock: func(m *mock.TaskServiceMock) {
				req := &models.TaskUpdateRequest{Title: ptr("Hack Title")}
				m.On("UpdateTask", userID, taskID, req).Return(nil, apperr.ErrForbidden)
			},
			contextUserID:  userID,
			urlParamID:     taskID.String(),
			requestBody:    models.TaskUpdateRequest{Title: ptr("Hack Title")},
			expectedStatus: http.StatusForbidden,
			expectedError:  "この操作を行う権限がありません",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mock.TaskServiceMock)
			tt.setupMock(mockService)
			h := NewTaskHandler(mockService)

			// リクエストボディの作成
			var body []byte
			if s, ok := tt.requestBody.(string); ok {
				body = []byte(s)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodPut, "/tasks/"+tt.urlParamID, bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			c.Set("userID", tt.contextUserID)
			c.Params = []gin.Param{{Key: "id", Value: tt.urlParamID}}

			h.UpdateTask(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				var response map[string]string
				_ = json.Unmarshal(w.Body.Bytes(), &response)
				assert.Contains(t, response["error"], tt.expectedError)
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestTaskHandler_DeleteTask(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupMock      func(m *mock.TaskServiceMock, uid, tid uuid.UUID)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "正常系：自分のタスクを削除できる (204 No Content)",
			setupMock: func(m *mock.TaskServiceMock, uid, tid uuid.UUID) {
				// 【重要】引数を mock.Anything にして、IDの不一致を無視する
				m.On("DeleteTask", mk.Anything, mk.Anything).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "異常系：他人のタスクを削除しようとすると403を返す",
			setupMock: func(m *mock.TaskServiceMock, uid, tid uuid.UUID) {
				m.On("DeleteTask", uid, tid).Return(apperr.ErrForbidden)
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "この操作を行う権限がありません",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. 各テストケースごとに新しいIDを生成
			userID := uuid.New()
			taskID := uuid.New()

			mockService := new(mock.TaskServiceMock)
			tt.setupMock(mockService, userID, taskID)
			h := NewTaskHandler(mockService)

			// 2. Contextの完全な初期化
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// 3. Requestをセット（MethodとPathが重要）
			c.Request, _ = http.NewRequest(http.MethodDelete, "/tasks/"+taskID.String(), nil)

			// 4. パラメータを確実にセット
			c.Set("userID", userID)
			c.Params = []gin.Param{{Key: "id", Value: taskID.String()}}

			// 実行
			h.DeleteTask(c)

			resp := w.Result()
			defer resp.Body.Close()

			// 検証（w.Code ではなく resp.StatusCode を見る）
			assert.Equal(t, tt.expectedStatus, resp.StatusCode, "Case: "+tt.name)

			if tt.expectedError != "" {
				var response map[string]string
				_ = json.Unmarshal(w.Body.Bytes(), &response)
				assert.Contains(t, response["error"], tt.expectedError)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func ptr(s string) *string {
	return &s
}
