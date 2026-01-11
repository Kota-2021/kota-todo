// internal/app/service/auth_service.go
package service

import (
	"errors"
	"fmt"
	"my-portfolio-2025/internal/app/apperr"
	"my-portfolio-2025/internal/app/models"
	"my-portfolio-2025/internal/app/repository"
	"my-portfolio-2025/pkg/auth"

	"gorm.io/gorm"
)

// AuthService は認証関連のビジネスロジックを定義
// UserRepoはUserRepositoryのインスタンス
type AuthServiceImpl struct {
	userRepo repository.UserRepository
}

// NewAuthService は AuthService の新しいインスタンスを作成します
func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &AuthServiceImpl{userRepo: userRepo}
}

// Signup はユーザー登録のビジネスロジックを実行します
func (s *AuthServiceImpl) Signup(req *models.SignupRequest) (*models.User, error) {
	// 1. ユーザー名の重複チェック
	_, err := s.userRepo.FindByUsername(req.Username)
	if err == nil {
		// エラーがない（＝ユーザーが見つかった）場合は重複エラーを返す
		return nil, fmt.Errorf("%w: username '%s' is already taken", apperr.ErrValidation, req.Username)
	}

	// レコード不在以外のDBエラーが発生した場合はそのまま返す
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("authService.Signup.FindByUsername: %w", err)
	}

	// 2. パスワードのハッシュ化
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("authService.Signup.HashPassword: %w", err)
	}

	// 3. ユーザーモデルの作成
	user := &models.User{
		Username: req.Username,
		Password: hashedPassword,
	}

	// 4. DBに保存
	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("authService.Signup.CreateUser: %w", err)
	}

	return user, nil
}

// AuthenticateUser はユーザー認証のビジネスロジックを実行します
func (s *AuthServiceImpl) AuthenticateUser(username, password string) (*models.User, string, error) {

	// 1. ユーザー名からユーザーを取得
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		// セキュリティ上、「ユーザーが存在しない」のか「パスワードが違う」のかを区別させないため
		// どちらも ErrUnauthorized として扱う
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", fmt.Errorf("%w: invalid username or password", apperr.ErrUnauthorized)
		}
		// その他のDBエラーはそのまま返す
		return nil, "", fmt.Errorf("authService.AuthenticateUser.FindByUsername: %w", err)
	}

	// 2. パスワード照合
	if ok, err := auth.CheckPasswordHash(password, user.Password); !ok {
		// 照合失敗時も ErrUnauthorized を返す
		if err != nil {
			return nil, "", fmt.Errorf("%w: %v", apperr.ErrUnauthorized, err)
		}
		// その他のエラーはそのまま返す
		return nil, "", fmt.Errorf("%w: invalid password", apperr.ErrUnauthorized)
	}

	// 3. JWT生成
	token, err := auth.GenerateToken(user.ID)

	if err != nil {
		return nil, "", fmt.Errorf("authService.AuthenticateUser.GenerateToken: %w", err)
	}

	return user, token, nil
}
