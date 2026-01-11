// internal/app/repository/user_repository.go
package repository

import (
	"fmt"
	"my-portfolio-2025/internal/app/models"

	"gorm.io/gorm"
)

// userRepositoryImpl は UserRepository インターフェースの具体的な実装です。
// DBはGormのDBインスタンス
type userRepositoryImpl struct {
	DB *gorm.DB
}

// NewUserRepository は UserRepository の新しいインスタンスを作成します
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepositoryImpl{DB: db}
}

func (r *userRepositoryImpl) CreateUser(user *models.User) error {
	if err := r.DB.Create(user).Error; err != nil {
		return fmt.Errorf("userRepository.CreateUser: %w", err)
	}
	return nil
}

// ユーザー名からユーザーを取得するメソッドを追加
func (r *userRepositoryImpl) FindByUsername(username string) (*models.User, error) {
	var user models.User
	if err := r.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, fmt.Errorf("userRepository.FindByUsername (username=%s): %w", username, err)
	}
	return &user, nil
}
