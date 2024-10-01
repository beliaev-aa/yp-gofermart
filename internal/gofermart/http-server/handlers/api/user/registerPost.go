package user

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"encoding/json"
	"errors"
	"go.uber.org/zap"
	"net/http"
)

// RegisterPostHandler — обработчик HTTP-запросов для регистрации пользователя
type RegisterPostHandler struct {
	authService *services.AuthService
	logger      *zap.Logger
}

// NewRegisterPostHandler — конструктор для создания обработчика регистрации
func NewRegisterPostHandler(authService *services.AuthService, logger *zap.Logger) *RegisterPostHandler {
	return &RegisterPostHandler{
		authService: authService,
		logger:      logger,
	}
}

// ServeHTTP — обрабатывает HTTP-запросы для регистрации пользователя.
func (h *RegisterPostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Декодируем запрос для получения данных пользователя
	var req domain.AuthenticationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Попытка зарегистрировать пользователя
	if err := h.authService.RegisterUser(req.Login, req.Password); err != nil {
		// Проверка на конфликт (если логин уже существует)
		if errors.Is(err, gofermartErrors.ErrLoginAlreadyExists) {
			h.logger.Warn("Registration failed", zap.String("login", req.Login))
			http.Error(w, "login already exist", http.StatusConflict)
		} else {
			h.logger.Error("Server error during registration", zap.Error(err))
			http.Error(w, "Server error", http.StatusInternalServerError)
		}
		return
	}

	// Генерация JWT токена для зарегистрированного пользователя
	token, err := h.authService.GenerateJWT(req.Login)
	if err != nil {
		h.logger.Error("Failed to generate JWT", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Успешная регистрация и авторизация
	h.logger.Info("User registered and authenticated", zap.String("login", req.Login))
	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}
