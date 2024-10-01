package balance

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"beliaev-aa/yp-gofermart/internal/gofermart/utils"
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
)

type (
	// IndexGetHandler - представляет HTTP-обработчик для получения текущего баланса пользователя.
	IndexGetHandler struct {
		logger            *zap.Logger
		userService       *services.UserService
		usernameExtractor utils.UsernameExtractor
	}
	// IndexGetResponse - представляет ответ API, содержащий информацию о балансе пользователя.
	IndexGetResponse struct {
		Current   float64 `json:"current"`   // Текущий баланс пользователя
		Withdrawn float64 `json:"withdrawn"` // Общая сумма выведенных средств
	}
)

// NewIndexGetHandler - создает новый экземпляр IndexGetHandler с указанными зависимостями.
func NewIndexGetHandler(userService *services.UserService, usernameExtractor utils.UsernameExtractor, logger *zap.Logger) *IndexGetHandler {
	return &IndexGetHandler{
		logger:            logger,
		userService:       userService,
		usernameExtractor: usernameExtractor,
	}
}

// ServeHTTP - обрабатывает HTTP-запросы для получения баланса пользователя.
func (h *IndexGetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	login, err := h.usernameExtractor.ExtractUsernameFromContext(r, h.logger)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	balance, err := h.userService.GetBalance(login)
	if err != nil {
		h.logger.Error("Failed to get balance", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	response := IndexGetResponse{
		Current:   balance.Current,
		Withdrawn: balance.Withdrawn,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		h.logger.Error("Failed to encode JSON response", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
