// pkg/auth/jwt.go
package auth

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims はJWTのペイロードを定義
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// ローカル開発用にデフォルト値を残しても良いですが、
		// 基本は ecs.tf の設定値が優先されます
		return []byte("YOUR_SUPER_SECRET_KEY_MUST_BE_SECURELY_MANAGED")
	}
	return []byte(secret)
}

// GenerateToken は指定されたユーザーIDのJWTを生成します
func GenerateToken(userID uuid.UUID) (string, error) {
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

	tokenString, err := token.SignedString(getJWTSecret())
	if err != nil {
		return "", errors.New("JWT署名エラー")
	}

	return tokenString, nil
}

// ValidateToken はJWT文字列を受け取り、検証してUserIDを返します。
func ValidateToken(tokenString string) (uuid.UUID, error) {
	// トークンのパースと検証
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// トークンの署名アルゴリズムが想定通りかチェック
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Method)
		}
		// シークレットキーを返します
		return getJWTSecret(), nil
	})

	if err != nil {
		// トークンの検証失敗 (署名無効、期限切れ、不正な形式など)
		return uuid.Nil, fmt.Errorf("token validation failed: %w", err)
	}

	// 検証に成功した場合、クレームを取得
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		// クレームの型変換失敗、またはトークンが有効でない
		return uuid.Nil, fmt.Errorf("invalid token claims or not valid")
	}

	// 期限切れチェック (ParseWithClaimsが通常処理しますが、念のため手動チェックも可能)
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return uuid.Nil, fmt.Errorf("token expired")
	}

	// 成功: 抽出したUserIDを返却
	return claims.UserID, nil // uuid.UUID型を返却
}
