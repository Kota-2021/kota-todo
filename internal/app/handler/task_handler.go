// internal/app/handler/task_handler.go
package handler

import (
	"my-portfolio-2025/internal/app/models"
	"my-portfolio-2025/internal/app/service" // Service層をインポート
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
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

// ユーザーID取得ヘルパー関数 (重要！)
// JWTミドルウェアで c.Set("userID", userID) された値を取得します。
func getUserIDFromContext(c *gin.Context) uint {
	// c.MustGetは値が設定されていない場合にパニックを引き起こしますが、
	// AuthMiddlewareが先に実行されるため、安全に利用できます。
	// MustGetの結果はinterface{}型なので、uint型にキャストが必要です。
	userID, ok := c.MustGet("userID").(uint)
	if !ok {
		// 通常はAuthMiddlewareで止まるため、ここは発生しない想定だが、念のためログ出力やエラー処理を検討
		return 0
	}
	return userID
}

// TaskHandler.CreateTask: POST /tasks
func (h *TaskHandler) CreateTask(c *gin.Context) {
	// 1. JWTミドルウェアから認証済みユーザーIDを取得 (最も重要)
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	// 2. リクエストボディをDTOにバインド（入力検証）
	var req models.TaskCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 3. Service層を呼び出し、タスク作成処理を委譲
	task, err := h.taskService.CreateTask(userID, &req)

	// 4. エラー処理
	if err != nil {
		// Service層からのエラーに応じて適切なHTTPステータスを返す
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	// 5. 成功レスポンスを返却
	// HTTP 201 Created ステータスと、作成されたタスクデータを返す
	c.JSON(http.StatusCreated, task)
}

// TaskHandler.GetTasks: GET /tasks (リスト取得)
func (h *TaskHandler) GetTasks(c *gin.Context) {
	// 1. JWTミドルウェアから認証済みユーザーIDを取得
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	// 2. Service層を呼び出し、ユーザーIDに紐づくタスクリストを取得
	tasks, err := h.taskService.GetTasks(userID)

	// 3. エラー処理
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}

	// 4. 成功レスポンスを返却
	// HTTP 200 OK ステータスと、タスクリストを返す
	c.JSON(http.StatusOK, tasks)
}

// GetTaskByID: GET /tasks/:id (詳細取得)
func (h *TaskHandler) GetTaskByID(c *gin.Context) {
	// 1. JWTミドルウェアから認証済みユーザーIDを取得
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	// 2. URLパスパラメータからタスクIDを取得
	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 64) // stringをuintに変換
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID format"})
		return
	}

	// 3. Service層を呼び出し、タスクを取得（Service層で認可チェックが行われる）
	task, err := h.taskService.GetTaskByID(userID, uint(taskID))

	// 4. エラー処理
	if err != nil {
		// タスクが存在しない場合 ('task not found' error)
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()}) // 404 Not Found
			return
		}
		// 認可エラーの場合 ('forbidden' error)
		if strings.Contains(err.Error(), "forbidden") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to this task"}) // 403 Forbidden
			return
		}
		// その他の内部エラー
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch task"})
		return
	}

	// 5. 成功レスポンスを返却
	c.JSON(http.StatusOK, task)
}

// UpdateTask: PUT /tasks/:id (更新)
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	// 1. 認証済みユーザーIDを取得
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	// 2. URLパスパラメータからタスクIDを取得
	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID format"})
		return
	}

	// 3. リクエストボディをDTOにバインド
	var req models.TaskUpdateRequest
	// ShouldBindJSONは、空のリクエストボディでもエラーを返さないように設計されることが多いです。
	// models.TaskUpdateRequestのフィールドをポインタ型にすることで、nilチェックが可能になります。
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 4. Service層を呼び出し（Service層で認可と更新処理が行われる）
	updatedTask, err := h.taskService.UpdateTask(userID, uint(taskID), &req)

	// 5. エラー処理
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()}) // 404 Not Found
			return
		}
		if strings.Contains(err.Error(), "forbidden") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to update this task"}) // 403 Forbidden
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	// 6. 成功レスポンスを返却
	c.JSON(http.StatusOK, updatedTask) // 更新成功は 200 OK
}

// DeleteTask: DELETE /tasks/:id (削除)
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	// 1. 認証済みユーザーIDを取得
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	// 2. URLパスパラメータからタスクIDを取得
	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID format"})
		return
	}

	// 3. Service層を呼び出し（Service層で認可と削除処理が行われる）
	err = h.taskService.DeleteTask(userID, uint(taskID))

	// 4. エラー処理
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()}) // 404 Not Found
			return
		}
		if strings.Contains(err.Error(), "forbidden") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to delete this task"}) // 403 Forbidden
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	// 5. 成功レスポンスを返却
	// 削除成功時は、コンテンツなし (204 No Content) を返すのがベストプラクティスです。
	c.Status(http.StatusNoContent)
}
