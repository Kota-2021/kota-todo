// pkg/auth/auth.go
package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword は平文のパスワードを bcrypt でハッシュ化します
func HashPassword(password string) (string, error) {
	// bcrypt のコストパラメータを設定（通常は DefaultCost で十分）
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPasswordHash は平文のパスワードとハッシュを比較します
// ログイン処理で利用します
func CheckPasswordHash(password, hash string) (bool, error) {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		// パスワードが一致しない場合
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, errors.New("認証情報が正しくありません") // 認証失敗
		}
		return false, errors.New("パスワード照合中にエラーが発生しました")
	}

	return true, nil
}
