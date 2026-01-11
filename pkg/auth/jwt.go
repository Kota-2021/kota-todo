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

// getJWTSecret は環境変数からシークレットキーを取得します
func getJWTSecret() ([]byte, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// 2026年本番環境に向け、デフォルト値は廃止しエラーを返します。
		// これにより、設定ミスによる脆弱な鍵の使用を防止します。
		return nil, errors.New("JWT_SECRET environment variable is required")
	}
	return []byte(secret), nil
}

// GenerateToken は指定されたユーザーIDのJWTを生成します
func GenerateToken(userID uuid.UUID) (string, error) {
	expirationTime := time.Now().Add(7 * 24 * time.Hour)

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secret, err := getJWTSecret()
	if err != nil {
		return "", fmt.Errorf("auth.GenerateToken: %w", err)
	}

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("auth.GenerateToken (signing): %w", err)
	}

	return tokenString, nil
}

// ValidateToken はJWT文字列を受け取り、検証してUserIDを返します。
func ValidateToken(tokenString string) (uuid.UUID, error) {
	secret, err := getJWTSecret()
	if err != nil {
		return uuid.Nil, fmt.Errorf("auth.ValidateToken (config): %w", err)
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 署名アルゴリズムの検証
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		// 期限切れや署名不一致などの具体的な理由はここでラッピングされる
		return uuid.Nil, fmt.Errorf("auth.ValidateToken: %w", err)
	}

	// クレームの抽出
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.UserID, nil
	}

	return uuid.Nil, errors.New("auth.ValidateToken: invalid token")
}
