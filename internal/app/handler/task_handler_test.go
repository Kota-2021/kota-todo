package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"my-portfolio-2025/internal/app/apperr"
	"my-portfolio-2025/internal/app/models"
	"my-portfolio-2025/internal/testutils/mock" // 作成済みのMockを想定

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

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
