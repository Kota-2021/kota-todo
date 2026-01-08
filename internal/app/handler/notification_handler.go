// internal/app/handler/notification_handler.go
package handler

import (
	"log"
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
	// CheckOrigin: オリジンのチェック
	// ブラウザからの接続を許可するかを判断する。
	// **開発中は全てのオリジンを許可（本番環境では適切に制限が必要）**
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

// HandleWS WebSocket接続の受付
func (h *NotificationHandler) HandleWS(c *gin.Context) {
	log.Println("--- WebSocket ハンドシェイク開始 ---")

	// 1. HTTPをWebSocketへアップグレード
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocketアップグレード失敗: %v", err)
		return
	}
	log.Println("WebSocketアップグレード成功")

	// **テスト用にユーザーIDを固定（UserID: 1）**
	// userID := uint(1) // 260108byKota
	userID := uuid.New()

	// 2. Hubに登録
	h.hub.Register <- &service.ClientRegistration{
		UserID: userID,
		Conn:   conn,
	}
	log.Printf("ユーザー %d が Hub に登録されました", userID)

	// 切断時の処理
	defer func() {
		h.hub.Unregister <- userID
		log.Printf("ユーザー %d の接続が終了しました", userID)
		conn.Close()
	}()

	// 3. 読み取りループ（これがないと即座に終了してしまいます）
	log.Println("クライアントからのメッセージ待機中...")
	for {
		// クライアントが切断するか、エラーが発生するまでここで待機
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Printf("接続終了 (ReadMessage): %v", err)
			break
		}
		log.Printf("📩 メッセージ受信: type=%d, payload=%s", messageType, string(p))
	}
}

// GetNotifications はログインユーザーの通知一覧を取得します
// GET /notifications?page=1
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	// ミドルウェアからUserIDを取得 (JWT認証済み前提)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDStr.(uuid.UUID)

	// クエリパラメータからページ番号を取得
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	notifications, err := h.svc.GetNotifications(c.Request.Context(), userID, page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch notifications"})
		return
	}

	c.JSON(http.StatusOK, notifications)
}

// MarkAsRead は特定の通知を既読にします
// PATCH /notifications/:id/read
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDStr.(uuid.UUID)

	// URLパスから通知IDを取得
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification id"})
		return
	}

	err = h.svc.MarkAsRead(c.Request.Context(), id, userID)
	if err != nil {
		// リポジトリ層でNotFoundだった場合
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "notification marked as read"})
}
