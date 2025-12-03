// internal/app/models/user.go
package models

import "time"

// User はデータベースの users テーブルに対応する構造体
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"unique;not null" json:"username"` // 一意性とNOT NULL制約
	Password  string    `gorm:"not null" json:"-"`               // ハッシュ化されたパスワード。レスポンスには含めないため `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SignupRequest はユーザー登録時にクライアントから受け取るデータ
type SignupRequest struct {
	Username string `json:"username" binding:"required"`       // Ginのバリデーションタグ
	Password string `json:"password" binding:"required,min=8"` // パスワードの最小長を8文字に設定
}

// SigninRequest はログイン時にクライアントから受け取るデータ
type SigninRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
