// internal/app/handler/notification_handler.go
package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"my-portfolio-2025/internal/app/apperr"
	"my-portfolio-2025/internal/app/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// WebSocketのアップグレーダー設定
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// すべての接続を許可する（開発用）
	// ブラウザからws接続する為に必要
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type NotificationHandler struct {
	svc service.NotificationService
	hub *service.NotificationHub
}

func NewNotificationHandler(svc service.NotificationService, hub *service.NotificationHub) *NotificationHandler {
	return &NotificationHandler{svc: svc, hub: hub}
}

// handleError: 他のハンドラーと共通のエラーハンドリング方針
func (h *NotificationHandler) handleError(c *gin.Context, err error) {
	var status int
	var msg string

	switch {
	case errors.Is(err, apperr.ErrNotFound):
		status = http.StatusNotFound
		msg = "指定された通知が見つかりません"
	case errors.Is(err, apperr.ErrUnauthorized):
		status = http.StatusUnauthorized
		msg = "認証が必要です"
	case errors.Is(err, apperr.ErrValidation):
		status = http.StatusBadRequest
		msg = "リクエストが不正です"
	default:
		slog.Error("Notification handler error", "error", err)
		status = http.StatusInternalServerError
		msg = "サーバー内部エラーが発生しました"
	}

	c.JSON(status, gin.H{"error": msg})
}

// HandleWS WebSocket接続の受付
func (h *NotificationHandler) HandleWS(c *gin.Context) {
	// 1. 認証チェック
	// AuthMiddlewareが正常に機能し、コンテキストにuserIDが入っていることが前提です
	val, ok := c.Get("userID")
	if !ok {
		slog.Warn("WebSocket unauthorized: userID not found in context")
		// 接続をアップグレードする前に 401 を返して拒否します
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	userID := val.(uuid.UUID)

	// 2. アップグレード
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("WebSocket upgrade failed", "error", err)
		return
	}

	// 3. Hubへの登録
	h.hub.Register <- &service.ClientRegistration{
		UserID: userID,
		Conn:   conn,
	}
	slog.Info("User registered to Hub", "userID", userID)

	defer func() {
		h.hub.Unregister <- userID
		slog.Info("User connection closed", "userID", userID)
		conn.Close()
	}()

	// 4. メッセージ待機ループ (接続維持のため)
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

// GetNotifications はログインユーザーの通知一覧を取得します
// GET /notifications?page=1
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		h.handleError(c, apperr.ErrUnauthorized)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	notifications, err := h.svc.GetNotifications(c.Request.Context(), userID.(uuid.UUID), page)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, notifications)
}

// MarkAsRead は特定の通知を既読にします
// PATCH /notifications/:id/read
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		h.handleError(c, apperr.ErrUnauthorized)
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.handleError(c, fmt.Errorf("%w: invalid notification id", apperr.ErrValidation))
		return
	}

	err = h.svc.MarkAsRead(c.Request.Context(), id, userID.(uuid.UUID))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "notification marked as read"})
}
