// internal/testutils/test_helpers.go
package testutils

import (
	"my-portfolio-2025/internal/app/models"
	"my-portfolio-2025/pkg/utils"
	"time"

	// JWTやハッシュ化に使用するライブラリをインポート
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type TestClaims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

// TestConfig はテストで使用する最低限の設定項目を定義します
type TestConfig struct {
	JWTSecretKey string
	// 必要であれば、他の設定項目 (例: PasswordHashCost など) を追加
}

// GlobalTestConfig は GetTestConfig から返されるダミー設定値
// 実際のアプリケーションのconfigパッケージとは別に定義します
var GlobalTestConfig = &TestConfig{
	// テストでのみ使用する固定の秘密鍵を設定します。
	// アプリケーションの本番環境のシークレットキーとは異なる値にしてください。
	JWTSecretKey: "test-secret-key-for-portfolio-project",
}

// CreateTestUser はテスト用のダミーユーザーインスタンスを返します
func CreateTestUser(id uuid.UUID, username string, hashedPassword string) *models.User {
	return &models.User{
		ID:        id,
		Username:  username,
		Password:  hashedPassword, // ハッシュ済みのパスワードを設定
		CreatedAt: utils.NowJST(),
		UpdatedAt: utils.NowJST(),
	}
}

// HashPassword はプレーンテキストのパスワードをハッシュ化します
func HashPassword(password string) (string, error) {
	// 実際のアプリケーションで使用しているコスト（例: bcrypt.DefaultCost）を使用
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// GenerateTestToken は指定されたユーザーIDのテスト用有効なJWTを生成します
func GenerateTestToken(userID uuid.UUID, secretKey string) (string, error) {
	// 1. 有効期限を設定 (例: 7日間。本番のjwt.goと合わせる)
	expirationTime := utils.NowJST().Add(7 * 24 * time.Hour)

	// 2. クレームを作成 (jwt.go の Claims 構造体を使用)
	claims := &TestClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(utils.NowJST()),
			// その他のクレーム (Issuer, Subjectなど) も、jwt.goにあれば追加
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 3. シークレットキーで署名
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GenerateExpiredToken は指定されたユーザーIDの期限切れのテスト用JWTを生成します
func GenerateExpiredToken(userID uuid.UUID, secretKey string) (string, error) {
	// 1. 過去の有効期限を設定 (例: 1時間前)
	expirationTime := utils.NowJST().Add(-time.Hour)

	// 2. クレームを作成 (jwt.go の Claims 構造体を使用)
	claims := &TestClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	// 2. トークンを生成し、HS256で秘密鍵を使って署名
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 3. シークレットキーで署名
	expiredToken, _ := token.SignedString([]byte(secretKey))

	return expiredToken, nil
}

// GenerateInvalidSignatureToken は指定されたユーザーIDの不正な署名のテスト用JWTを生成します
func GenerateInvalidSignatureToken(userID uuid.UUID, secretKey string) (string, error) {
	// 1. 有効期限を設定 (例: 1時間後)
	claims := &TestClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(utils.NowJST().Add(time.Hour)), // 期限は有効
		},
	}
	// 2. トークンを生成し、HS256で秘密鍵を使って署名
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 3. シークレットキーで署名
	invalidSignatureToken, _ := token.SignedString([]byte(secretKey))

	return invalidSignatureToken, nil
}

// GetTestConfig はテスト実行用の設定構造体インスタンスを返します
func GetTestConfig() *TestConfig {
	return GlobalTestConfig
}
