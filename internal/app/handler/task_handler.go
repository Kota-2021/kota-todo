// internal/app/handler/task_handler.go
package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"my-portfolio-2025/internal/app/apperr"
	"my-portfolio-2025/internal/app/models"
	"my-portfolio-2025/internal/app/service" // Service層をインポート
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TaskHandler はタスク関連のHTTPリクエストを処理します。
type TaskHandler struct {
	taskService service.TaskService // TaskServiceへの依存性
}

// NewTaskHandler は TaskHandler の新しいインスタンスを作成します。
// DIコンテナやメイン関数からTaskService実装を受け取ります。
func NewTaskHandler(s service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: s}
}

// handleError: README要件に基づきエラーレスポンスを一元管理するヘルパー
func (h *TaskHandler) handleError(c *gin.Context, err error) {
	var status int
	var msg string

	// errors.Is を使用して Service 層から返されたカスタムエラーを判定
	switch {
	case errors.Is(err, apperr.ErrNotFound):
		status = http.StatusNotFound
		msg = "指定されたタスクが見つかりません"
	case errors.Is(err, apperr.ErrForbidden):
		// 認可違反はセキュリティ上重要なので、この段階でログを出力（項目6の対応）
		slog.Warn("Authorization violation attempt", "error", err)
		status = http.StatusForbidden
		msg = "この操作を行う権限がありません"
	case errors.Is(err, apperr.ErrValidation):
		status = http.StatusBadRequest
		msg = err.Error() // バリデーション内容はユーザーに伝える
	case errors.Is(err, apperr.ErrUnauthorized):
		status = http.StatusUnauthorized
		msg = "認証が必要です"
	default:
		// 内部エラーの詳細はユーザーに隠蔽しつつ、ログには残す（項目7の対応）
		slog.Error("Internal server error", "error", err)
		status = http.StatusInternalServerError
		msg = "サーバー内部でエラーが発生しました"
	}

	c.JSON(status, gin.H{"error": msg})
}

// ユーザーID取得ヘルパー関数 (重要！)
// JWTミドルウェアで c.Set("userID", userID) された値を取得します。
func getUserIDFromContext(c *gin.Context) uuid.UUID {
	// MustGet はキーがないとパニックになるため、Get を使用する
	val, exists := c.Get("userID")
	if !exists {
		return uuid.Nil
	}

	// 型アサーション
	userID, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return userID
}

// TaskHandler.CreateTask: POST /tasks
func (h *TaskHandler) CreateTask(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		h.handleError(c, apperr.ErrUnauthorized)
		return
	}

	var req models.TaskCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, fmt.Errorf("%w: %v", apperr.ErrValidation, err))
		return
	}

	task, err := h.taskService.CreateTask(userID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, task)
}

// TaskHandler.GetTasks: GET /tasks (リスト取得)
func (h *TaskHandler) GetTasks(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		h.handleError(c, apperr.ErrUnauthorized)
		return
	}

	tasks, err := h.taskService.GetTasks(userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// GetTaskByID: GET /tasks/:id (詳細取得)
func (h *TaskHandler) GetTaskByID(c *gin.Context) {
	userID := getUserIDFromContext(c)
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		h.handleError(c, fmt.Errorf("%w: invalid UUID", apperr.ErrValidation))
		return
	}

	task, err := h.taskService.GetTaskByID(userID, taskID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, task)
}

// UpdateTask: PUT /tasks/:id (更新)
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	userID := getUserIDFromContext(c)
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		h.handleError(c, fmt.Errorf("%w: invalid UUID", apperr.ErrValidation))
		return
	}

	var req models.TaskUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, fmt.Errorf("%w: %v", apperr.ErrValidation, err))
		return
	}

	updatedTask, err := h.taskService.UpdateTask(userID, taskID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, updatedTask)
}

// DeleteTask: DELETE /tasks/:id (削除)
func (h *TaskHandler) DeleteTask(c *gin.Context) {

	// デバッグログ：ここがターミナルに出れば、関数は呼ばれている
	fmt.Println("DEBUG: DeleteTask started")

	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		h.handleError(c, apperr.ErrUnauthorized)
		return
	}
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		h.handleError(c, fmt.Errorf("%w: invalid UUID", apperr.ErrValidation))
		return
	}

	if err := h.taskService.DeleteTask(userID, taskID); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
	c.Writer.WriteHeaderNow()
}
