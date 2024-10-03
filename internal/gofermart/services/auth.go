package services

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"beliaev-aa/yp-gofermart/internal/gofermart/storage"
	"errors"
	"github.com/go-chi/jwtauth/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type AuthService struct {
	logger    *zap.Logger
	tokenAuth *jwtauth.JWTAuth
	userRepo  storage.UserRepository
}

func NewAuthService(jwtSecret []byte, userRepo storage.UserRepository, logger *zap.Logger) *AuthService {
	tokenAuth := jwtauth.New("HS256", jwtSecret, nil)
	return &AuthService{
		logger:    logger,
		tokenAuth: tokenAuth,
		userRepo:  userRepo,
	}
}

func (s *AuthService) RegisterUser(login, password string) error {
	s.logger.Info("Attempting to register user", zap.String("login", login))

	user, _ := s.userRepo.GetUserByLogin(login)
	if user != nil {
		s.logger.Warn("Login already taken", zap.String("login", login))
		return gofermartErrors.ErrLoginAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Error generating password hash", zap.Error(err))
		return err
	}

	url := &domain.User{Login: login, Password: string(hashedPassword)}
	err = s.userRepo.SaveUser(*url)
	if err != nil {
		s.logger.Error("Error registering user", zap.String("login", login), zap.Error(err))
		return err
	}

	s.logger.Info("User registered successfully", zap.String("login", login))
	return nil
}

func (s *AuthService) AuthenticateUser(login, password string) (bool, error) {
	s.logger.Info("Attempting to authenticate user", zap.String("login", login))

	user, err := s.userRepo.GetUserByLogin(login)
	if err != nil {
		if errors.Is(err, gofermartErrors.ErrUserNotFound) {
			s.logger.Warn("User not found", zap.String("login", login))
			return false, nil
		}
		s.logger.Error("Error getting user", zap.Error(err))
		return false, err
	}

	if user == nil {
		s.logger.Warn("Login not found", zap.String("login", login))
		return false, nil
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		s.logger.Warn("Invalid password", zap.String("login", login))
		return false, nil
	}

	s.logger.Info("User authenticated successfully", zap.String("login", login))
	return true, nil
}

func (s *AuthService) GenerateJWT(username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	_, tokenString, err := s.tokenAuth.Encode(map[string]interface{}{
		"username": username,
		"exp":      expirationTime,
	})
	return tokenString, err
}

func (s *AuthService) GetTokenAuth() *jwtauth.JWTAuth {
	return s.tokenAuth
}
