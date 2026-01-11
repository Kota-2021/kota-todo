package apperr

import (
	"errors"
	"fmt"
)

// カスタムエラーの定義
var (
	ErrNotFound     = errors.New("resource not found")    // 404用
	ErrForbidden    = errors.New("access forbidden")      // 403用
	ErrUnauthorized = errors.New("unauthorized")          // 401用
	ErrValidation   = errors.New("validation failed")     // 400用
	ErrInternal     = errors.New("internal server error") // 500用
)

// AppError は追加の文脈を持たせるための構造体（任意）
// 将来的に「ユーザー向けメッセージ」などを保持させたい場合に便利です
type AppError struct {
	Err     error
	Message string
}

// Error はエラーメッセージを返します
func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

// Unwrap はエラーを返します
func (e *AppError) Unwrap() error {
	return e.Err
}
