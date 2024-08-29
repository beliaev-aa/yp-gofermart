package services

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"sync"
	"time"
)

const ErrorLoginAlreadyExist = "login already exist"

type (
	AuthService struct {
		users      map[string]string
		usersMutex sync.Mutex
		jwtSecret  []byte
		logger     *zap.Logger
	}
	Claims struct {
		Username string `json:"username"`
		jwt.RegisteredClaims
	}
)

func NewAuthService(jwtSecret []byte, logger *zap.Logger) *AuthService {
	return &AuthService{
		users:     make(map[string]string),
		jwtSecret: jwtSecret,
		logger:    logger,
	}
}

func (s *AuthService) RegisterUser(login, password string) error {
	s.logger.Info("Attempting to register user", zap.String("login", login))

	s.usersMutex.Lock()
	defer s.usersMutex.Unlock()

	if _, exists := s.users[login]; exists {
		s.logger.Warn("Login already taken", zap.String("login", login))
		return errors.New(ErrorLoginAlreadyExist)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Error generating password hash", zap.Error(err))
		return err
	}

	s.users[login] = string(hashedPassword)
	s.logger.Info("User registered successfully", zap.String("login", login))
	return nil
}

func (s *AuthService) AuthenticateUser(login, password string) (bool, error) {
	s.logger.Info("Attempting to authenticate user", zap.String("login", login))

	s.usersMutex.Lock()
	defer s.usersMutex.Unlock()

	hashedPassword, exists := s.users[login]
	if !exists {
		s.logger.Warn("Login not found", zap.String("login", login))
		return false, nil
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		s.logger.Warn("Invalid password", zap.String("login", login))
		return false, nil
	}

	s.logger.Info("User authenticated successfully", zap.String("login", login))
	return true, nil
}

func (s *AuthService) GenerateJWT(username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
