// internal/testutils/mock/user_mock.go
package mock

import (
	"my-portfolio-2025/internal/app/models" // モデルパッケージへのパスは適宜修正してください

	"github.com/stretchr/testify/mock"
)

// MockUserRepository は repository.UserRepository インターフェースのモックです
type MockUserRepository struct {
	mock.Mock
}

// SaveUser は UserRepository.SaveUser のモック実装です
func (m *MockUserRepository) CreateUser(user *models.User) error {
	args := m.Called(user)

	// var savedUser *models.User
	// if args.Get(0) != nil {
	// 	savedUser = args.Get(0).(*models.User)
	// }

	return args.Error(0)
}

// FindUserByUsername は UserRepository.FindUserByUsername のモック実装です
func (m *MockUserRepository) FindByUsername(username string) (*models.User, error) {
	// Mockオブジェクトに設定された期待値（引数と戻り値）に基づいて処理を実行します
	args := m.Called(username)

	var user *models.User
	if args.Get(0) != nil {
		user = args.Get(0).(*models.User)
	}

	return user, args.Error(1)
}
