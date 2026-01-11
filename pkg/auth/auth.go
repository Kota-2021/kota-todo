// pkg/auth/auth.go
package auth

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword は平文のパスワードを bcrypt でハッシュ化します
func HashPassword(password string) (string, error) {
	// コストパラメータは DefaultCost (10) が一般的です
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("auth.HashPassword: %w", err)
	}
	return string(bytes), nil
}

// CheckPasswordHash は平文のパスワードとハッシュを比較します
// ログイン処理で利用します
func CheckPasswordHash(password, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		// パスワードの不一致は「エラー」ではなく「不一致（false）」という結果として返す
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		// それ以外のエラー（ハッシュ形式が壊れている等）はエラーとして返す
		return false, fmt.Errorf("auth.CheckPasswordHash: %w", err)
	}

	return true, nil
}
