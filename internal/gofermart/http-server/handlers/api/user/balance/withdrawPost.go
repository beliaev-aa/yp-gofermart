package balance

import (
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"beliaev-aa/yp-gofermart/internal/gofermart/utils"
	"encoding/json"
	"errors"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"net/http"
)

type (
	// WithdrawPostHandler - представляет HTTP-обработчик для вывода средств пользователем.
	WithdrawPostHandler struct {
		logger            *zap.Logger
		userService       *services.UserService
		usernameExtractor utils.UsernameExtractor
	}
	// WithdrawPostRequest - представляет структуру запроса на вывод средств.
	WithdrawPostRequest struct {
		Order string  `json:"order"`
		Sum   float64 `json:"sum"`
	}
)

// NewWithdrawPostHandler - создает новый экземпляр WithdrawPostHandler.
func NewWithdrawPostHandler(userService *services.UserService, usernameExtractor utils.UsernameExtractor, logger *zap.Logger) *WithdrawPostHandler {
	return &WithdrawPostHandler{
		logger:            logger,
		userService:       userService,
		usernameExtractor: usernameExtractor,
	}
}

// ServeHTTP - обрабатывает HTTP-запрос POST для вывода средств.
func (h *WithdrawPostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	login, err := h.usernameExtractor.ExtractUsernameFromContext(r, h.logger)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var req WithdrawPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Invalid request format", zap.Error(err))
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if !utils.IsValidMoon(req.Order) {
		h.logger.Warn("Invalid order number format", zap.String("order", req.Order))
		http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	err = h.userService.Withdraw(login, req.Order, decimal.NewFromFloat(req.Sum))
	if err != nil {
		switch {
		case errors.Is(err, gofermartErrors.ErrInsufficientFunds):
			http.Error(w, "Insufficient funds", http.StatusPaymentRequired)
		default:
			h.logger.Error("Failed to process withdrawal", zap.Error(err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
