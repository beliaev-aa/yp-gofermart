package services

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"beliaev-aa/yp-gofermart/tests"
	"errors"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestNewAuthService(t *testing.T) {
	t.Run("NewAuthService_CreatesService", func(t *testing.T) {
		logger := zap.NewNop()
		mockStorage := &tests.MockStorage{}
		jwtSecret := []byte("secret")

		authService := NewAuthService(jwtSecret, logger, mockStorage)

		if authService == nil || authService.tokenAuth == nil {
			t.Errorf("Expected AuthService to be initialized with JWTAuth")
		}
		if authService.logger != logger {
			t.Errorf("Expected AuthService to be initialized with provided logger")
		}
		if authService.storage != mockStorage {
			t.Errorf("Expected AuthService to be initialized with provided storage")
		}
	})
}

func TestRegisterUser(t *testing.T) {
	testCases := []struct {
		Name          string
		MockReturn    func() (*domain.User, error)
		MockSave      func() error
		ExpectedError error
		Login         string
		Password      string
	}{
		{
			Name: "RegisterUser_Success",
			MockReturn: func() (*domain.User, error) {
				return nil, nil
			},
			MockSave: func() error {
				return nil
			},
			ExpectedError: nil,
			Login:         "new_user",
			Password:      "password123",
		},
		{
			Name: "RegisterUser_LoginAlreadyExists",
			MockReturn: func() (*domain.User, error) {
				return &domain.User{Login: "new_user"}, nil
			},
			MockSave:      func() error { return nil },
			ExpectedError: gofermartErrors.ErrLoginAlreadyExists,
			Login:         "new_user",
			Password:      "password123",
		},
		{
			Name: "RegisterUser_SaveUserError",
			MockReturn: func() (*domain.User, error) {
				return nil, nil
			},
			MockSave: func() error {
				return errors.New("failed to save user")
			},
			ExpectedError: errors.New("failed to save user"),
			Login:         "new_user",
			Password:      "password123",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mockStore := &tests.MockStorage{
				GetUserByLoginFn: func(login string) (*domain.User, error) {
					return tc.MockReturn()
				},
				SaveUserFn: func(user domain.User) error {
					return tc.MockSave()
				},
			}
			logger := zap.NewNop()
			authService := NewAuthService([]byte("secret"), logger, mockStore)

			err := authService.RegisterUser(tc.Login, tc.Password)

			if err != nil && tc.ExpectedError == nil {
				t.Errorf("Expected no error, got %v", err)
			} else if err == nil && tc.ExpectedError != nil {
				t.Errorf("Expected error, got none")
			} else if err != nil && err.Error() != tc.ExpectedError.Error() {
				t.Errorf("Expected error %v, got %v", tc.ExpectedError, err)
			}
		})
	}
}

func TestAuthenticateUser(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	testCases := []struct {
		Name          string
		MockReturn    func() (*domain.User, error)
		Login         string
		Password      string
		ExpectedAuth  bool
		ExpectedError error
	}{
		{
			Name: "AuthenticateUser_Success",
			MockReturn: func() (*domain.User, error) {
				return &domain.User{Login: "test_user", Password: string(hashedPassword)}, nil
			},
			Login:         "test_user",
			Password:      "password123",
			ExpectedAuth:  true,
			ExpectedError: nil,
		},
		{
			Name: "AuthenticateUser_UserNotFound",
			MockReturn: func() (*domain.User, error) {
				return nil, gofermartErrors.ErrUserNotFound
			},
			Login:         "test_user",
			Password:      "password123",
			ExpectedAuth:  false,
			ExpectedError: nil,
		},
		{
			Name: "AuthenticateUser_LoginNotFound",
			MockReturn: func() (*domain.User, error) {
				return nil, nil // Хранилище возвращает nil, что имитирует отсутствие пользователя
			},
			Login:         "test_user",
			Password:      "password123",
			ExpectedAuth:  false,
			ExpectedError: nil, // Ошибки быть не должно
		},
		{
			Name: "AuthenticateUser_InvalidPassword",
			MockReturn: func() (*domain.User, error) {
				return &domain.User{Login: "test_user", Password: string(hashedPassword)}, nil
			},
			Login:         "test_user",
			Password:      "wrong_password",
			ExpectedAuth:  false,
			ExpectedError: nil,
		},
		{
			Name: "AuthenticateUser_GetUserError",
			MockReturn: func() (*domain.User, error) {
				return nil, errors.New("db error")
			},
			Login:         "test_user",
			Password:      "password123",
			ExpectedAuth:  false,
			ExpectedError: errors.New("db error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mockStore := &tests.MockStorage{
				GetUserByLoginFn: func(login string) (*domain.User, error) {
					return tc.MockReturn()
				},
			}
			logger := zap.NewNop()
			authService := NewAuthService([]byte("secret"), logger, mockStore)

			authenticated, err := authService.AuthenticateUser(tc.Login, tc.Password)

			if authenticated != tc.ExpectedAuth {
				t.Errorf("Expected authenticated %v, got %v", tc.ExpectedAuth, authenticated)
			}

			if err != nil && tc.ExpectedError == nil {
				t.Errorf("Expected no error, got %v", err)
			} else if err == nil && tc.ExpectedError != nil {
				t.Errorf("Expected error, got none")
			} else if err != nil && err.Error() != tc.ExpectedError.Error() {
				t.Errorf("Expected error %v, got %v", tc.ExpectedError, err)
			}
		})
	}
}

func TestGenerateJWT(t *testing.T) {
	t.Run("GenerateJWT_Success", func(t *testing.T) {
		logger := zap.NewNop()
		mockStorage := &tests.MockStorage{}
		authService := NewAuthService([]byte("secret"), logger, mockStorage)

		token, err := authService.GenerateJWT("test_user")

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if token == "" {
			t.Errorf("Expected valid JWT token, got empty string")
		}
	})
}

func TestGetTokenAuth(t *testing.T) {
	t.Run("GetTokenAuth_Success", func(t *testing.T) {
		logger := zap.NewNop()
		mockStorage := &tests.MockStorage{}
		authService := NewAuthService([]byte("secret"), logger, mockStorage)

		tokenAuth := authService.GetTokenAuth()

		if tokenAuth == nil {
			t.Errorf("Expected non-nil tokenAuth, got nil")
		}
	})
}
