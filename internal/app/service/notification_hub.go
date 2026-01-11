package service

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"my-portfolio-2025/internal/app/models"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// Hub は全てのWebSocket接続を管理し、メッセージを配信します
type NotificationHub struct {
	// 接続中のクライアントを管理 (ユーザーID -> WebSocket接続)
	// 1ユーザーが複数デバイスで繋ぐ場合は [] *websocket.Conn にしますが、
	// 今回はシンプルに 1ユーザー1接続として実装します
	clients map[uuid.UUID]*websocket.Conn

	// クライアントからの新規接続通知用チャネル
	Register chan *ClientRegistration

	// クライアントの切断通知用チャネル
	Unregister chan uuid.UUID

	// 配信メッセージ用チャネル
	Broadcast chan *models.NotificationMessage

	// マップ操作時の排他制御用
	mu sync.Mutex

	// Redisクライアント
	redisClient *redis.Client
}

// 登録用構造体
type ClientRegistration struct {
	UserID uuid.UUID
	Conn   *websocket.Conn
}

// NewNotificationHub は新しいハブを作成します
func NewNotificationHub(redisClient *redis.Client) *NotificationHub {
	return &NotificationHub{
		clients:     make(map[uuid.UUID]*websocket.Conn),
		Register:    make(chan *ClientRegistration),
		Unregister:  make(chan uuid.UUID),
		Broadcast:   make(chan *models.NotificationMessage),
		redisClient: redisClient,
	}
}

// Run はハブのメインループを実行します（Goルーチンとして起動）
// Run はハブのメインループを実行します（Goルーチンとして起動）
func (h *NotificationHub) Run(ctx context.Context) {
	log.Println("Notification Hub is running...")

	pubsub := h.redisClient.Subscribe(ctx, "notifications")
	defer pubsub.Close()

	ch := pubsub.Channel()

	for {
		select {
		// 1. 終了信号の受信
		case <-ctx.Done():
			slog.Info("Notification Hub shutting down")
			return

		// 2. Redis（外部）からメッセージが届いた時
		case msg := <-ch:
			if msg == nil {
				continue
			}

			// ここで解析（Unmarshal）を行う
			var notif models.NotificationMessage
			if err := json.Unmarshal([]byte(msg.Payload), &notif); err != nil {
				slog.Error("Failed to unmarshal Redis message", "error", err)
				continue
			}

			// 内部の配信用チャネル（Broadcast）へ転送する
			// これにより、下の "case msg := <-h.Broadcast" が起動します
			h.Broadcast <- &notif

		// 3. クライアントの新規接続
		case reg := <-h.Register:
			h.mu.Lock()
			h.clients[reg.UserID] = reg.Conn
			h.mu.Unlock()
			log.Printf("User %s connected via WebSocket", reg.UserID)

		// 4. クライアントの切断
		case userID := <-h.Unregister:
			h.mu.Lock()
			if conn, ok := h.clients[userID]; ok {
				conn.Close()
				delete(h.clients, userID)
				log.Printf("User %s disconnected", userID)
			}
			h.mu.Unlock()

		// 5. クライアントへの実際の配信
		case msg := <-h.Broadcast:
			h.mu.Lock()
			if conn, ok := h.clients[msg.UserID]; ok {
				err := conn.WriteJSON(msg)
				if err != nil {
					log.Printf("Error sending message to user %s: %v", msg.UserID, err)
					conn.Close()
					delete(h.clients, msg.UserID)
				}
			}
			h.mu.Unlock()
		}
	}
}

// 1. Redisへの「出版」処理
func (h *NotificationHub) PublishMessage(ctx context.Context, msg models.NotificationMessage) error {
	payload, _ := json.Marshal(msg)
	// "notifications" チャンネルへ送信
	return h.redisClient.Publish(ctx, "notifications", payload).Err()
}

// 2. Redisからの「購読」ループ（Hub.Run() などから Goルーチンで起動）
func (h *NotificationHub) SubscribeRedis(ctx context.Context) {
	pubsub := h.redisClient.Subscribe(ctx, "notifications")
	defer pubsub.Close()

	ch := pubsub.Channel()
	log.Println("Subscribed to Redis notifications channel...")

	for msg := range ch {
		var notif models.NotificationMessage
		if err := json.Unmarshal([]byte(msg.Payload), &notif); err != nil {
			log.Printf("Failed to unmarshal Redis message: %v", err)
			continue
		}

		// ★ポイント: Redisから受け取ったメッセージを Broadcast チャネルに投げる
		// これにより、Run() メソッド内の case msg := <-h.Broadcast が反応し、
		// 適切な WebSocket 接続に送信される。
		h.Broadcast <- &notif
	}
}
