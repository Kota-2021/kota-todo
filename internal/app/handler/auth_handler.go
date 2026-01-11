// internal/app/handler/auth_handler.go
package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"my-portfolio-2025/internal/app/apperr"
	"my-portfolio-2025/internal/app/models"
	"my-portfolio-2025/internal/app/service"
	"net/http"

	"github.com/gin-gonic/gin"
	// PostgreSQL固有のエラーコードを処理するためにインポート
)

// AuthController は認証関連のエンドポイントを処理します
// AuthServiceはAuthServiceのインスタンス
type AuthController struct {
	AuthService service.AuthService
}

// SigninResponse: ログイン成功時に返すDTO（Data Transfer Object）
// 認証に成功したユーザー情報とJWTを含める
type SigninResponse struct {
	Token string `json:"token"`
	// 必要であればユーザー名などの情報もここに含める
	// User *models.UserResponse `json:"user"`
}

// NewAuthController は AuthController の新しいインスタンスを作成します
func NewAuthController(authService service.AuthService) *AuthController {
	return &AuthController{AuthService: authService}
}

// handleError: エラーレスポンス形式の統一と抽象化（TaskHandlerと同様のロジック）
func (c *AuthController) handleError(ctx *gin.Context, err error) {
	var status int
	var msg string

	switch {
	case errors.Is(err, apperr.ErrValidation):
		status = http.StatusBadRequest
		msg = err.Error() // バリデーション内容はユーザーに伝える
	case errors.Is(err, apperr.ErrUnauthorized):
		status = http.StatusUnauthorized
		msg = "ユーザー名またはパスワードが正しくありません"
	case errors.Is(err, apperr.ErrInternal):
		status = http.StatusInternalServerError
		msg = "サーバー内部エラーが発生しました"
	default:
		// 未定義のエラーは500として扱い、詳細はログにのみ記録
		slog.Error("Auth error occurred", "error", err)
		status = http.StatusInternalServerError
		msg = "予期せぬエラーが発生しました"
	}

	ctx.JSON(status, gin.H{"error": msg})
}

// Signup はユーザー登録のエンドポイントを処理します
func (c *AuthController) Signup(ctx *gin.Context) {
	var req models.SignupRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.handleError(ctx, fmt.Errorf("%w: %v", apperr.ErrValidation, err))
		return
	}

	user, err := c.AuthService.Signup(&req)
	if err != nil {
		c.handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user":    gin.H{"id": user.ID, "username": user.Username},
	})
}

// Signin は POST /signin のハンドラー関数です
func (c *AuthController) Signin(ctx *gin.Context) {
	var req models.SigninRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.handleError(ctx, fmt.Errorf("%w: invalid request format", apperr.ErrValidation))
		return
	}

	_, token, err := c.AuthService.AuthenticateUser(req.Username, req.Password)
	if err != nil {
		c.handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, SigninResponse{
		Token: token,
	})
}
