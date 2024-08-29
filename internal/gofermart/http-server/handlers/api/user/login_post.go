package user

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
)

type LoginPostHandler struct {
	authService *services.AuthService
	logger      *zap.Logger
}

func NewLoginPostHandler(authService *services.AuthService, logger *zap.Logger) *LoginPostHandler {
	return &LoginPostHandler{
		authService: authService,
		logger:      logger,
	}
}

// ServeHTTP обрабатывает HTTP-запросы для аутентификации пользователя.
func (h *LoginPostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req domain.AuthenticationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	authenticated, err := h.authService.AuthenticateUser(req.Login, req.Password)
	if err != nil {
		h.logger.Error("Server error during authentication", zap.Error(err))
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	if !authenticated {
		h.logger.Warn("Authentication failed", zap.String("login", req.Login))
		http.Error(w, "Invalid login/password", http.StatusUnauthorized)
		return
	}

	token, err := h.authService.GenerateJWT(req.Login)
	if err != nil {
		h.logger.Error("Failed to generate JWT", zap.Error(err))
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("User authenticated", zap.String("login", req.Login))
	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}
