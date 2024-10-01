package user

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"beliaev-aa/yp-gofermart/internal/gofermart/utils"
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type (
	// WithdrawalsGetHandler — обработчик HTTP-запросов для получения списка выводов пользователя
	WithdrawalsGetHandler struct {
		logger            *zap.Logger
		userService       *services.UserService
		usernameExtractor utils.UsernameExtractor
	}
	// WithdrawalResponse — структура для представления ответа о выводе средств
	WithdrawalResponse struct {
		Order       string  `json:"order"`
		Sum         float64 `json:"sum"`
		ProcessedAt string  `json:"processed_at"`
	}
)

// NewWithdrawalsGetHandler — конструктор для создания обработчика WithdrawalsGetHandler
func NewWithdrawalsGetHandler(userService *services.UserService, usernameExtractor utils.UsernameExtractor, logger *zap.Logger) *WithdrawalsGetHandler {
	return &WithdrawalsGetHandler{
		userService:       userService,
		usernameExtractor: usernameExtractor,
		logger:            logger,
	}
}

// ServeHTTP — основной метод для обработки входящих HTTP-запросов
func (h *WithdrawalsGetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	login, err := h.usernameExtractor.ExtractUsernameFromContext(r, h.logger)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	withdrawals, err := h.userService.GetWithdrawals(login)
	if err != nil {
		h.logger.Error("Failed to get withdrawals", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var response []WithdrawalResponse
	for _, wd := range withdrawals {
		item := WithdrawalResponse{
			Order:       wd.OrderNumber,
			Sum:         wd.Amount,
			ProcessedAt: wd.ProcessedAt.Format(time.RFC3339),
		}
		response = append(response, item)
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		h.logger.Error("Failed to encode JSON response", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
