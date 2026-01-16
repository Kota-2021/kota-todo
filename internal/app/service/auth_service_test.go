// internal/app/service/auth_service_test.go
package service

import (
	"my-portfolio-2025/internal/app/models"
	"my-portfolio-2025/internal/testutils"
	"my-portfolio-2025/internal/testutils/mock"
	"testing"

	"my-portfolio-2025/pkg/auth"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	mockPkg "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	// 独自のValidationErrorやErrUserAlreadyExistsなどのエラーがあればインポート
)

// AuthTestSuite は認証サービス (AuthService) のテストスイートです
type AuthTestSuite struct {
	suite.Suite
	mockUserRepo *mock.MockUserRepository
	authService  AuthService // auth_service.go で定義したインターフェース型
}

// SetupTest は各テストケースの前に実行されます
func (s *AuthTestSuite) SetupTest() {
	// 1. モックの初期化
	s.mockUserRepo = new(mock.MockUserRepository)

	// 2. サービスの実装にモックと設定を注入
	s.authService = NewAuthService(s.mockUserRepo)
}

// TestAuthServiceSuite はテストスイートを実行します
func TestAuthServiceSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}

// Singnupテスト(1)
// テストケース: ユーザー名が重複せず、パスワードがハッシュ化されてリポジトリに渡されることを確認
func (s *AuthTestSuite) TestSignup_Success() {
	t := s.T()
	username := "unique-user-for-signup-test"
	password := "strongpass123"

	// --- 1. テストデータの準備 (SignupRequest構造体を作成) ---
	signupRequest := &models.SignupRequest{
		Username: username,
		Password: password,
	}

	// --- 2. モックの期待値設定 ---

	// (1) FindByUsername: ユーザーが存在しないこと (nil, nil) をシミュレート
	// 'username'という引数で一度呼ばれることを期待
	s.mockUserRepo.On("FindByUsername", username).Return((*models.User)(nil), nil).Once()

	// (2) CreateUser: ユーザー作成が成功すること (nil error) をシミュレート
	// s.mockUserRepo.On("CreateUser", mock.AnythingOfType("*models.User")).
	s.mockUserRepo.On("CreateUser", mockPkg.AnythingOfType("*models.User")).
		Return(nil). // 戻り値はエラーなし
		Run(func(args mockPkg.Arguments) {
			// CreateUser に渡されたユーザーオブジェクトの検証
			user := args.Get(0).(*models.User)

			// パスワードがハッシュ化されているかを確認（コアロジックのテスト）
			// 実際のアプリケーションで利用しているハッシュ化ライブラリを使用
			// 例: bcrypt.CompareHashAndPassword
			err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
			assert.NoError(t, err, "CreateUserに渡されたパスワードはハッシュ化されているべき")
			assert.Equal(t, username, user.Username, "ユーザー名が正しくセットされているべき")
		}).
		Once()

	// --- 3. 実行と検証 ---
	// 戻り値2つ (*models.User と error) を受け取るように修正
	user, err := s.authService.Signup(signupRequest)

	// エラーがないことを検証
	assert.NoError(t, err, "正常な登録でエラーが発生してはならない")
	// userオブジェクトがnilでないことを確認する検証を追加するとより良い
	assert.NotNil(t, user, "登録成功時、ユーザーオブジェクトはnilであってはならない")

	// 4. モックの呼び出し検証
	s.mockUserRepo.AssertExpectations(t)
}

// Singnupテスト(2)
// テストケース: ユーザー名が重複している場合、エラーが発生することを確認
func (s *AuthTestSuite) TestSignup_UserAlreadyExists() {
	t := s.T()

	// 1. テストデータの準備 (FindByUsernameが返す既存ユーザー)
	username := "existinguser"
	password := "anypassword"

	signupRequest := &models.SignupRequest{
		Username: username,
		Password: password,
	}

	// ID 1, ハッシュ済みパスワードを持つ既存ユーザー
	existingUser := testutils.CreateTestUser(uuid.New(), username, "dummyhash")

	// --- 2. モックの期待値設定 ---

	// (1) FindByUsername: ユーザーが存在する (*models.User, nil) をシミュレート
	s.mockUserRepo.On("FindByUsername", username).Return(existingUser, nil).Once()

	// --- 3. 実行と検証 ---
	// 戻り値2つ (*models.User と error) を受け取るように修正
	user, err := s.authService.Signup(signupRequest)

	// エラーがあることを検証
	assert.Error(t, err, "重複ユーザー登録でエラーが発生すべき")
	// userオブジェクトがnilであることを確認する検証を追加するとより良い
	assert.Nil(t, user, "重複ユーザー登録時、ユーザーオブジェクトはnilであるべき")

	// 4. CreateUser が呼ばれていないことを検証
	s.mockUserRepo.AssertNotCalled(s.T(), "CreateUser")
}

// Signin(AuthenticateUser)テスト(1)
// テストケース: 正しいパスワードの場合、ユーザーオブジェクトとJWTトークンが返されることを確認
func (s *AuthTestSuite) TestSignin_Success() {
	t := s.T()
	username := "testuser"
	password := "correctpass"

	// 1. テストデータの準備: 正しいパスワードのハッシュ化
	hashedPassword, _ := testutils.HashPassword(password)
	authenticatedUser := testutils.CreateTestUser(uuid.New(), username, hashedPassword)

	// --- 2. モックの期待値設定 ---

	// (1) FindByUsername: ユーザーが存在すること (*models.User, nil) をシミュレート
	// 認証ロジックはまずユーザー名でDBを検索するため、これを設定
	s.mockUserRepo.On("FindByUsername", username).Return(authenticatedUser, nil).Once()

	// --- 3. 実行と検証 ---

	// AuthenticateUser を呼び出し、3つの戻り値を受け取る
	user, token, err := s.authService.AuthenticateUser(username, password)

	// (1) エラーがないことを検証
	assert.NoError(t, err, "正常なサインインでエラーが発生してはならない")

	// (2) ユーザーオブジェクトが返されていることを検証
	assert.NotNil(t, user, "認証成功時、ユーザーオブジェクトはnilであってはならない")
	// assert.NotEqual(t, uuid.Nil, userID, "返されたユーザーIDは空ではないはず")

	// (3) JWTトークンが生成されていることを検証
	assert.NotEmpty(t, token, "認証成功時、JWTトークンは空であってはならない")

	// 4. モックの呼び出し検証
	s.mockUserRepo.AssertExpectations(t)

	// 💡 より高度な検証:
	// ここで testutils.ParseToken(token) のようなヘルパーを使い、
	// 生成されたトークンのペイロードにユーザーID (1) が含まれていることを検証すると完璧です。
}

// Signin(AuthenticateUser)テスト(2)
// テストケース: 誤ったパスワードの場合、エラーが発生することを確認
func (s *AuthTestSuite) TestSignin_PasswordMismatch() {
	t := s.T()

	username := "testuser"
	correctPassword := "correctpass"
	wrongPassword := "incorrectpass" // 意図的に間違ったパスワード

	// 1. テストデータの準備: 正しいパスワードでハッシュ化されたユーザーを用意
	hashedPassword, _ := testutils.HashPassword(correctPassword)
	existingUser := testutils.CreateTestUser(uuid.New(), username, hashedPassword)

	// --- 2. モックの期待値設定 ---

	// (1) FindByUsername: ユーザーが存在すること (*models.User, nil) をシミュレート
	s.mockUserRepo.On("FindByUsername", username).Return(existingUser, nil).Once()

	// --- 3. 実行と検証 ---

	// 誤ったパスワードで認証を試みる
	user, token, err := s.authService.AuthenticateUser(username, wrongPassword)

	// (1) 期待されるエラーを検証 (例: ErrInvalidCredentials)
	assert.Error(t, err, "パスワード不一致でエラーが発生すべき")
	// assert.Equal(s.T(), service.ErrInvalidCredentials, err, "エラータイプが不一致") // カスタムエラーの場合

	// (2) ユーザーオブジェクトとトークンが返されていないことを検証
	assert.Nil(t, user, "認証失敗時、ユーザーオブジェクトはnilであるべき")
	assert.Empty(t, token, "認証失敗時、JWTトークンは空であるべき")

	// 4. モックの呼び出し検証
	s.mockUserRepo.AssertExpectations(t)
}

// Signin(AuthenticateUser)テスト(3)
// テストケース: ユーザーが見つからない場合、エラーが発生することを確認
func (s *AuthTestSuite) TestSignin_UserNotFound() {
	t := s.T()
	username := "nonexistentuser"
	password := "anypass"

	// --- 1. モックの期待値設定 ---

	// (1) FindByUsername: ユーザーが見つからないことをシミュレート
	// FindByUsernameがgorm.ErrRecordNotFoundなどのエラーを返すように設定
	// サービス層はこれをErrUserNotFoundまたはErrInvalidCredentialsに変換するはず
	s.mockUserRepo.On("FindByUsername", username).Return((*models.User)(nil), gorm.ErrRecordNotFound).Once()
	// ※ gorm.ErrRecordNotFound は repository層が返す具体的なエラー。
	//   service層がこれを何に変換するかによって、assert.Error の期待値が変わります。

	// --- 2. 実行と検証 ---
	user, token, err := s.authService.AuthenticateUser(username, password)

	// (1) 期待されるエラーを検証
	// セキュリティ上の理由から、ユーザー不在でもパスワード不一致と同じErrInvalidCredentialsを返すことが推奨されます
	assert.Error(t, err, "ユーザー不在でエラーが発生すべき")
	// assert.Equal(s.T(), service.ErrInvalidCredentials, err, "エラータイプが不一致")

	// (2) ユーザーオブジェクトとトークンが返されていないことを検証
	assert.Nil(t, user, "ユーザー不在時、ユーザーオブジェクトはnilであるべき")
	assert.Empty(t, token, "ユーザー不在時、JWTトークンは空であるべき")

	// 3. モックの呼び出し検証
	s.mockUserRepo.AssertExpectations(t)
}

// JWTVerificationテスト(1)
// テストケース: 正常に生成されたJWTトークンが、サービスの認証ロジックまたは検証ヘルパーによって正しくパースされ、ユーザーIDなどのクレームを取り出せることを確認
func (s *AuthTestSuite) TestJWTVerification_Success() {
	t := s.T()
	username := "verifieduser"
	password := "testpass"
	userID := uuid.New()

	// 1. テストデータの準備: 認証成功時のシミュレーション
	// ※ ユーザーモデルにJWTトークン生成に必要なフィールド（例: ID）が含まれていることを前提とします。
	hashedPassword, _ := testutils.HashPassword(password)
	authenticatedUser := testutils.CreateTestUser(userID, username, hashedPassword)

	// --- 2. モックの期待値設定 ---

	// AuthenticateUser内部で呼ばれる FindByUsername をモック
	s.mockUserRepo.On("FindByUsername", username).Return(authenticatedUser, nil).Once()

	// --- 3. 実行 (AuthenticateUserでトークンを生成) ---
	_, token, err := s.authService.AuthenticateUser(username, password)

	// トークンが生成され、エラーがないことを確認 (念のため)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// --- 4. JWT検証ロジックの検証 ---

	// 認可ミドルウェアが使用するロジックを直接テストする
	// トークンからユーザーIDを取り出す関数を呼び出す
	extractedUserID, validationErr := auth.ValidateToken(token)

	// (1) 検証エラーがないことを確認
	assert.NoError(t, validationErr, "有効なトークンの検証でエラーが発生してはならない")

	// (2) ユーザーIDが正しく取り出されていることを確認
	assert.Equal(t, userID, extractedUserID, "トークンから抽出されたユーザーIDが一致するべき")

	s.mockUserRepo.AssertExpectations(t)
}

// JWTVerificationテスト(2)
// テストケース: 期限切れのJWTトークンが、エラーを返すことを確認
func (s *AuthTestSuite) TestJWTVerification_ExpiredToken() {
	t := s.T()
	userID := uuid.New()

	// --- 1. 期限切れトークンの作成 ---
	//期限切れトークンを生成するためのヘルパー関数GenerateExpiredTokenを使用
	expiredToken, _ := testutils.GenerateExpiredToken(userID, testutils.GlobalTestConfig.JWTSecretKey)

	// --- 3. 実行と検証 ---

	// 期限切れトークンで検証を試みる
	extractedUserID, err := auth.ValidateToken(expiredToken)

	// (1) エラーが発生していることを検証
	assert.Error(t, err, "期限切れトークンはエラーを返す必要がある")

	// (2) 期待されるエラータイプを検証
	// トークンライブラリが出すエラー、またはそれを変換したカスタムエラー(例: service.ErrTokenExpired)
	// assert.True(t, errors.Is(err, service.ErrTokenExpired), "エラーはErrTokenExpiredであるべき")

	// (3) ユーザーIDが0または無効な値であることを検証
	assert.Equal(t, uuid.Nil, extractedUserID, "抽出されるユーザーIDは0であるべき")
}

// JWTVerificationテスト(3)
// テストケース: 不正な署名のJWTトークンが、エラーを返すことを確認
func (s *AuthTestSuite) TestJWTVerification_InvalidSignature() {
	t := s.T()
	userID := uuid.New()

	// --- 1. 異なる秘密鍵で署名されたトークンの作成 ---

	// ⚠️ WARNING: 必ず設定ファイルにある秘密鍵とは異なる、ダミーの秘密鍵を使用します
	wrongSecret := "ThisIsNotTheRealJWTSecretKey12345"

	// JWTライブラリ(例: github.com/dgrijalva/jwt-go/v4)を使用した場合の生成ロジックの例:
	// claims := models.Claims{
	//     UserID: userID,
	//     RegisteredClaims: jwt.RegisteredClaims{
	//         ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)), // 期限は有効
	//     },
	// }
	// token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// invalidSignatureToken, _ := token.SignedString([]byte(wrongSecret)) // 異なる鍵で署名！

	// --- 2. 代わりに、異なる秘密鍵で署名されたトークンを格納 ---
	// invalidSignatureToken := "your_generated_invalid_signature_token_string" // 異なる鍵で署名されたトークン文字列を格納
	invalidSignatureToken, _ := testutils.GenerateInvalidSignatureToken(userID, wrongSecret)

	// --- 3. 実行と検証 ---

	// 不正な署名のトークンで検証を試みる
	extractedUserID, err := auth.ValidateToken(invalidSignatureToken)

	// (1) エラーが発生していることを検証
	assert.Error(t, err, "無効な署名のトークンはエラーを返す必要がある")

	// (2) 期待されるエラータイプを検証
	// assert.True(t, errors.Is(err, service.ErrInvalidSignature), "エラーはErrInvalidSignatureであるべき")

	// (3) ユーザーIDが0または無効な値であることを検証
	assert.Equal(t, uuid.Nil, extractedUserID, "抽出されるユーザーIDは0であるべき")
}
