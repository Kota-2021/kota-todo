package service

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Hub は全てのWebSocket接続を管理し、メッセージを配信します
type NotificationHub struct {
	// 接続中のクライアントを管理 (ユーザーID -> WebSocket接続)
	// 1ユーザーが複数デバイスで繋ぐ場合は [] *websocket.Conn にしますが、
	// 今回はシンプルに 1ユーザー1接続として実装します
	clients map[uint]*websocket.Conn

	// クライアントからの新規接続通知用チャネル
	Register chan *ClientRegistration

	// クライアントの切断通知用チャネル
	Unregister chan uint

	// 配信メッセージ用チャネル
	Broadcast chan *NotificationMessage

	// マップ操作時の排他制御用
	mu sync.Mutex
}

// 登録用構造体
type ClientRegistration struct {
	UserID uint
	Conn   *websocket.Conn
}

// 配信メッセージ構造体
type NotificationMessage struct {
	UserID  uint   `json:"user_id"`
	Message string `json:"message"`
}

// NewNotificationHub は新しいハブを作成します
func NewNotificationHub() *NotificationHub {
	return &NotificationHub{
		clients:    make(map[uint]*websocket.Conn),
		Register:   make(chan *ClientRegistration),
		Unregister: make(chan uint),
		Broadcast:  make(chan *NotificationMessage),
	}
}

// Run はハブのメインループを実行します（Goルーチンとして起動）
func (h *NotificationHub) Run() {
	log.Println("Notification Hub is running...")
	for {
		select {
		// クライアントからの新規接続通知用チャネルを受け取った場合の処理
		case reg := <-h.Register:
			h.mu.Lock()
			// クライアントをマップに追加
			h.clients[reg.UserID] = reg.Conn
			h.mu.Unlock()
			log.Printf("User %d connected via WebSocket", reg.UserID)

		// クライアントの切断通知用チャネルを受け取った場合の処理
		case userID := <-h.Unregister:
			h.mu.Lock()
			// クライアントをマップから削除
			if conn, ok := h.clients[userID]; ok {
				conn.Close()
				delete(h.clients, userID)
				log.Printf("User %d disconnected", userID)
			}
			h.mu.Unlock()

		// 配信メッセージ用チャネルを受け取った場合の処理
		case msg := <-h.Broadcast:
			h.mu.Lock()
			// クライアントにメッセージを送信
			if conn, ok := h.clients[msg.UserID]; ok {
				err := conn.WriteJSON(msg)
				if err != nil {
					log.Printf("Error sending message to user %d: %v", msg.UserID, err)
					conn.Close()
					// クライアントをマップから削除
					delete(h.clients, msg.UserID)
				}
			}
			h.mu.Unlock()
		}
	}
}
