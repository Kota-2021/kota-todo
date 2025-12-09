// internal/app/repository/user_repository.go
package repository

import (
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

// CreateUser は新しいユーザーレコードをDBに作成します
func (r *userRepositoryImpl) CreateUser(user *models.User) error {
	// DB操作。ユーザー名が重複している場合はエラー（Gormのエラーをそのまま返す）
	result := r.DB.Create(user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// ユーザー名からユーザーを取得するメソッドを追加
func (r *userRepositoryImpl) FindByUsername(username string) (*models.User, error) {
	var user models.User
	// GormでUsernameを条件に検索
	if err := r.DB.Where("username = ?", username).First(&user).Error; err != nil {
		// Gormのレコードが見つからないエラー (gorm.ErrRecordNotFound) をチェック
		return nil, err
	}
	return &user, nil
}
