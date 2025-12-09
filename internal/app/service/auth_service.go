// internal/app/service/auth_service.go
package service

import (
	"my-portfolio-2025/internal/app/models"
)

// AuthService は認証に関するビジネスロジックを定義します。
type AuthService interface {
	// CreateTask: 新しいタスクを作成。UserIDを必須とする。
	Signup(req *models.SignupRequest) (*models.User, error)

	// AuthenticateUser: ユーザーを認証。
	AuthenticateUser(username, password string) (*models.User, string, error)
}
