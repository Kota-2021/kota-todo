// internal/app/service/auth_service.go
package service

import (
	"errors"
	"my-portfolio-2025/internal/app/models"
	"my-portfolio-2025/internal/app/repository"
	"my-portfolio-2025/pkg/auth"

	"gorm.io/gorm"
)

// AuthService は認証関連のビジネスロジックを定義
// UserRepoはUserRepositoryのインスタンス
type AuthService struct {
	UserRepo *repository.UserRepository
}

// NewAuthService は AuthService の新しいインスタンスを作成します
func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{UserRepo: userRepo}
}

// Signup はユーザー登録のビジネスロジックを実行します
func (s *AuthService) Signup(req *models.SignupRequest) (*models.User, error) {
	// 1. パスワードのハッシュ化
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// 2. ユーザーモデルの作成とハッシュ化パスワードの設定
	user := &models.User{
		Username: req.Username,
		Password: hashedPassword, // ハッシュ化されたパスワードを格納
	}

	// 3. DBに保存
	if err := s.UserRepo.CreateUser(user); err != nil {
		// ここでユーザー名重複エラーなどを適切に処理する（例: Gormのエラーコードで判断）
		return nil, err // ひとまずエラーをそのまま返す
	}

	return user, nil
}

// AuthenticateUser はユーザー認証のビジネスロジックを実行します
func (s *AuthService) AuthenticateUser(username, password string) (*models.User, error) {
	// 1. ユーザー名からユーザーを取得
	user, err := s.UserRepo.FindByUsername(username)
	if err != nil {
		// レコードが見つからないエラーの場合、認証失敗として扱う
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("認証情報が正しくありません") // 認証失敗
		}
		return nil, err // その他のDBエラー
	}

	// 2. パスワード照合 (ステップ2-3)
	// CompareHashAndPassword(保存されているハッシュ, 平文のパスワード)
	if ok, err := auth.CheckPasswordHash(password, user.Password); !ok {
		return nil, err // 認証失敗
	}

	// 認証成功
	return user, nil

}
