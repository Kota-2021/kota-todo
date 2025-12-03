// pkg/auth/jwt.go
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWT_SECRET は環境変数などから取得する（セキュリティ上の重要項目）
// AWS Secrets Managerから取得した秘密鍵を使用
const JWT_SECRET = "YOUR_SUPER_SECRET_KEY_MUST_BE_SECURELY_MANAGED"

// Claims はJWTのペイロードを定義
type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateToken は指定されたユーザーIDのJWTを生成します
func GenerateToken(userID uint) (string, error) {
	// 1. 有効期限を設定 (例: 7日間)
	expirationTime := time.Now().Add(7 * 24 * time.Hour)

	// 2. クレームを作成
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// 3. トークンを生成し、HS256で秘密鍵を使って署名
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(JWT_SECRET))
	if err != nil {
		return "", errors.New("JWT署名エラー")
	}

	return tokenString, nil
}
