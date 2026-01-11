package service

import (
	"context"
	"encoding/json"
	"fmt"
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
func (h *NotificationHub) Run(ctx context.Context) {
	slog.Info("Notification Hub is running...")

	// --- 1. Redis監視ループを独立したGoroutineで動かす (以前の SubscribeRedis 相当) ---
	go func() {
		pubsub := h.redisClient.Subscribe(ctx, "notifications")
		defer pubsub.Close()
		ch := pubsub.Channel()
		slog.Info("Subscribed to Redis notifications")

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				// 受信ログ (Debug)
				fmt.Printf("DEBUG: Received something from Redis: %s\n", msg.Payload)

				var notif models.NotificationMessage
				if err := json.Unmarshal([]byte(msg.Payload), &notif); err != nil {
					slog.Error("Failed to unmarshal Redis message", "error", err)
					continue
				}

				// 独立した外側から Broadcast チャネルへ流し込む
				h.Broadcast <- &notif
			}
		}
	}()

	// --- 2. Hub管理ループ (以前の Run 相当) ---
	for {
		select {
		case <-ctx.Done():
			slog.Info("Notification Hub shutting down")
			return

		case reg := <-h.Register:
			h.mu.Lock()
			h.clients[reg.UserID] = reg.Conn
			h.mu.Unlock()
			slog.Info("User connected", "userID", reg.UserID)

		case userID := <-h.Unregister:
			h.mu.Lock()
			if conn, ok := h.clients[userID]; ok {
				conn.Close()
				delete(h.clients, userID)
				slog.Info("User disconnected", "userID", userID)
			}
			h.mu.Unlock()

		case msg := <-h.Broadcast:
			// 配信ログを追加
			slog.Info("Attempting to broadcast message", "targetUserID", msg.UserID)

			h.mu.Lock()
			if conn, ok := h.clients[msg.UserID]; ok {
				err := conn.WriteJSON(msg)
				if err != nil {
					slog.Error("Failed to send WebSocket message", "userID", msg.UserID, "error", err)
					conn.Close()
					delete(h.clients, msg.UserID)
				} else {
					slog.Info("✅ Notification sent successfully", "userID", msg.UserID)
				}
			} else {
				slog.Warn("Recipient not found in active connections", "userID", msg.UserID)
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
