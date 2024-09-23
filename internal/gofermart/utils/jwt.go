package utils

import (
	"errors"
	"github.com/go-chi/jwtauth/v5"
	"go.uber.org/zap"
	"net/http"
)

// UsernameExtractor - интерфейс для извлечения имени пользователя из контекста
type UsernameExtractor interface {
	ExtractUsernameFromContext(r *http.Request, logger *zap.Logger) (string, error)
}

// RealUsernameExtractor - реальная реализация интерфейса для продакшн-кода
type RealUsernameExtractor struct{}

// ExtractUsernameFromContext - извлекает имя пользователя из контекста JWT токена
func (e *RealUsernameExtractor) ExtractUsernameFromContext(r *http.Request, logger *zap.Logger) (string, error) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	login, ok := claims["username"].(string)
	if !ok {
		logger.Warn("Unauthorized access attempt")
		return "", errors.New("unauthorized")
	}
	return login, nil
}
