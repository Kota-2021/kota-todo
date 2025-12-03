// internal/app/handler/auth_handler.go
package handler

import (
	"my-portfolio-2025/internal/app/models"
	"my-portfolio-2025/internal/app/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq" // PostgreSQL固有のエラーコードを処理するためにインポート
)

// AuthController は認証関連のエンドポイントを処理します
// AuthServiceはAuthServiceのインスタンス
type AuthController struct {
	AuthService *service.AuthService
}

// NewAuthController は AuthController の新しいインスタンスを作成します
func NewAuthController(authService *service.AuthService) *AuthController {
	return &AuthController{AuthService: authService}
}

// Signup はユーザー登録のエンドポイントを処理します
func (c *AuthController) Signup(ctx *gin.Context) {
	var req models.SignupRequest

	// 1. リクエストのバインディングとバリデーション
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body or validation failed", "details": err.Error()})
		return
	}

	// 2. Service層の呼び出し
	user, err := c.AuthService.Signup(&req)
	if err != nil {
		// ユーザー名重複エラーのハンドリング (PostgreSQL/pq のエラーコードを使用)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
			ctx.JSON(http.StatusConflict, gin.H{"error": "Username already taken"})
			return
		}
		// その他のエラー
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user", "details": err.Error()})
		return
	}

	// 3. 成功レスポンスの返却 (201 Created)
	// セキュリティのため、パスワード情報を含まないレスポンスを返す
	// models.User に `json:"-"` を付けていれば自動的に除外される
	ctx.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user":    gin.H{"id": user.ID, "username": user.Username},
	})
}

// Signin は POST /signin のハンドラー関数です
func (c *AuthController) Signin(ctx *gin.Context) {
	var req models.SigninRequest

	// 1. リクエストボディのバインドとバリデーション (ステップ2-1)
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "リクエスト形式が不正です"})
		return
	}

	// 2. Service層での認証処理を呼び出し (ステップ2-2, 2-3)
	// ここで認証が成功し、JWTが生成される想定です
	// Service層からJWTトークンとユーザー情報を取得するように調整が必要です
	// 例: token, user, err := h.AuthService.AuthenticateAndGenerateToken(req.Username, req.Password)

	// 現在は認証ロジックのみを呼び出し、トークン生成は次のステップで行います
	user, err := c.AuthService.AuthenticateUser(req.Username, req.Password)

	if err != nil {
		// 認証失敗時 (ユーザーNotFoundやパスワード不一致) のエラーハンドリング
		if err.Error() == "認証情報が正しくありません" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()}) // 401 Unauthorized
			return
		}
		// その他のサービス層エラー (DB接続エラーなど)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "ログイン処理中にエラーが発生しました"})
		return
	}

	// --- ここから JWT 生成 (ステップ2-4) と レスポンス返却 (ステップ2-5) が続きます ---

	// 認証が成功した時点 (JWT生成前)
	// c.JSON(http.StatusOK, gin.H{"message": "認証成功 (JWT生成と返却待ち)"})

	// JWT生成に進むため、一旦コメントアウト
	// ...
}
