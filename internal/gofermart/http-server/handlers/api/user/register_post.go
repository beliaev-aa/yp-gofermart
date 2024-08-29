package user

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
)

type RegisterPostHandler struct {
	authService *services.AuthService
	logger      *zap.Logger
}

func NewRegisterPostHandler(authService *services.AuthService, logger *zap.Logger) *RegisterPostHandler {
	return &RegisterPostHandler{
		authService: authService,
		logger:      logger,
	}
}

// ServeHTTP обрабатывает HTTP-запросы для регистрации пользователя.
func (h *RegisterPostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req domain.AuthenticationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if err := h.authService.RegisterUser(req.Login, req.Password); err != nil {
		if err.Error() == services.ErrorLoginAlreadyExist {
			h.logger.Warn("Registration failed", zap.String("login", req.Login))
			http.Error(w, "login already exist", http.StatusConflict)
		} else {
			h.logger.Error("Server error during registration", zap.Error(err))
			http.Error(w, "Server error", http.StatusInternalServerError)
		}
		return
	}

	token, err := h.authService.GenerateJWT(req.Login)
	if err != nil {
		h.logger.Error("Failed to generate JWT", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("User registered and authenticated", zap.String("login", req.Login))
	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}
