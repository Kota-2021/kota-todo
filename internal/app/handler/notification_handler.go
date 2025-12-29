// internal/app/handler/notification_handler.go
package handler

import (
	"log"
	"my-portfolio-2025/internal/app/service"
	"net/http"

	"github.com/gin-gonic/gin"
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
	hub *service.NotificationHub
}

func NewNotificationHandler(hub *service.NotificationHub) *NotificationHandler {
	return &NotificationHandler{hub: hub}
}

// HandleWS WebSocket接続の受付
func (h *NotificationHandler) HandleWS(c *gin.Context) {
	// 1. JWT等からユーザーIDを取得（W3-D13で実装した認証を利用）
	// userID, exists := c.Get("userID")
	// if !exists {
	//     c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	//     return
	// }

	// 2. HTTPをWebSocketへアップグレード
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// ユーザーIDを取得 (JWT等から)
	userID := uint(1) // 本来は authMiddleware 等から取得

	// Hubに登録依頼を出す
	h.hub.Register <- &service.ClientRegistration{
		UserID: userID,
		Conn:   conn,
	}

	// 切断時はHubに解除依頼を出す
	defer func() {
		h.hub.Unregister <- userID
	}()

	// 接続を維持するために、クライアントからの読み取りループを回す
	// (これがないと関数が終了して defer で切断されてしまいます)
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}

}
