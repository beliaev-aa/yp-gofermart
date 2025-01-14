package services

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"beliaev-aa/yp-gofermart/tests/mocks"
	"errors"
	"github.com/golang/mock/gomock"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"testing"
)

func TestNewAuthService(t *testing.T) {
	t.Run("NewAuthService_CreatesService", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockUserRepo := mocks.NewMockUserRepository(ctrl)
		logger := zap.NewNop()
		jwtSecret := []byte("secret")

		authService := NewAuthService(jwtSecret, mockUserRepo, logger)

		if authService == nil || authService.tokenAuth == nil {
			t.Errorf("Expected AuthService to be initialized with JWTAuth")
		}
		if authService.logger != logger {
			t.Errorf("Expected AuthService to be initialized with provided logger")
		}
		if authService.userRepo != mockUserRepo {
			t.Errorf("Expected AuthService to be initialized with provided storage")
		}
	})
}

func TestRegisterUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	logger := zap.NewNop()
	jwtSecret := []byte("secret")

	testCases := []struct {
		name          string
		setupMocks    func()
		expectedError error
		login         string
		password      string
	}{
		{
			name: "RegisterUser_Success",
			setupMocks: func() {
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), gomock.Any()).Return(nil, nil)
				mockUserRepo.EXPECT().SaveUser(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedError: nil,
			login:         "new_user",
			password:      "password123",
		},
		{
			name: "RegisterUser_LoginAlreadyExists",
			setupMocks: func() {
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), gomock.Any()).Return(&domain.User{Login: "new_user"}, nil)
			},
			expectedError: gofermartErrors.ErrLoginAlreadyExists,
			login:         "new_user",
			password:      "password123",
		},
		{
			name: "RegisterUser_SaveUserError",
			setupMocks: func() {
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), gomock.Any()).Return(nil, nil)
				mockUserRepo.EXPECT().SaveUser(gomock.Any(), gomock.Any()).Return(errors.New("failed to save user"))
			},
			expectedError: errors.New("failed to save user"),
			login:         "new_user",
			password:      "password123",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()

			authService := NewAuthService(jwtSecret, mockUserRepo, logger)

			err := authService.RegisterUser(tc.login, tc.password)

			if err != nil && tc.expectedError == nil {
				t.Errorf("Expected no error, got %v", err)
			} else if err == nil && tc.expectedError != nil {
				t.Errorf("Expected error, got none")
			} else if err != nil && err.Error() != tc.expectedError.Error() {
				t.Errorf("Expected error %v, got %v", tc.expectedError, err)
			}
		})
	}
}

func TestAuthenticateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	logger := zap.NewNop()
	jwtSecret := []byte("secret")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	testCases := []struct {
		name          string
		mockReturn    func(tx *gorm.DB, login string) (*domain.User, error)
		login         string
		password      string
		expectedAuth  bool
		expectedError error
	}{
		{
			name: "AuthenticateUser_Success",
			mockReturn: func(tx *gorm.DB, login string) (*domain.User, error) {
				return &domain.User{Login: "test_user", Password: string(hashedPassword)}, nil
			},
			login:         "test_user",
			password:      "password123",
			expectedAuth:  true,
			expectedError: nil,
		},
		{
			name: "AuthenticateUser_UserNotFound",
			mockReturn: func(tx *gorm.DB, login string) (*domain.User, error) {
				return nil, gofermartErrors.ErrUserNotFound
			},
			login:         "test_user",
			password:      "password123",
			expectedAuth:  false,
			expectedError: nil,
		},
		{
			name: "AuthenticateUser_LoginNotFound",
			mockReturn: func(tx *gorm.DB, login string) (*domain.User, error) {
				return nil, nil
			},
			login:         "test_user",
			password:      "password123",
			expectedAuth:  false,
			expectedError: nil,
		},
		{
			name: "AuthenticateUser_InvalidPassword",
			mockReturn: func(tx *gorm.DB, login string) (*domain.User, error) {
				return &domain.User{Login: "test_user", Password: string(hashedPassword)}, nil
			},
			login:         "test_user",
			password:      "wrong_password",
			expectedAuth:  false,
			expectedError: nil,
		},
		{
			name: "AuthenticateUser_GetUserError",
			mockReturn: func(tx *gorm.DB, login string) (*domain.User, error) {
				return nil, errors.New("db error")
			},
			login:         "test_user",
			password:      "password123",
			expectedAuth:  false,
			expectedError: errors.New("db error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), gomock.Any()).DoAndReturn(tc.mockReturn)

			authService := NewAuthService(jwtSecret, mockUserRepo, logger)

			authenticated, err := authService.AuthenticateUser(tc.login, tc.password)

			if authenticated != tc.expectedAuth {
				t.Errorf("Expected authenticated %v, got %v", tc.expectedAuth, authenticated)
			}

			if err != nil && tc.expectedError == nil {
				t.Errorf("Expected no error, got %v", err)
			} else if err == nil && tc.expectedError != nil {
				t.Errorf("Expected error, got none")
			} else if err != nil && err.Error() != tc.expectedError.Error() {
				t.Errorf("Expected error %v, got %v", tc.expectedError, err)
			}
		})
	}
}

func TestGenerateJWT(t *testing.T) {
	t.Run("GenerateJWT_Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockUserRepo := mocks.NewMockUserRepository(ctrl)
		logger := zap.NewNop()
		jwtSecret := []byte("secret")

		authService := NewAuthService(jwtSecret, mockUserRepo, logger)

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
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockUserRepo := mocks.NewMockUserRepository(ctrl)
		logger := zap.NewNop()
		jwtSecret := []byte("secret")

		authService := NewAuthService(jwtSecret, mockUserRepo, logger)

		tokenAuth := authService.GetTokenAuth()

		if tokenAuth == nil {
			t.Errorf("Expected non-nil tokenAuth, got nil")
		}
	})
}
