package balance

import (
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"beliaev-aa/yp-gofermart/internal/gofermart/utils"
	"encoding/json"
	"errors"
	"github.com/go-chi/jwtauth/v5"
	"go.uber.org/zap"
	"net/http"
)

type WithdrawPostHandler struct {
	userService *services.UserService
	logger      *zap.Logger
}

func NewWithdrawPostHandler(userService *services.UserService, logger *zap.Logger) *WithdrawPostHandler {
	return &WithdrawPostHandler{
		userService: userService,
		logger:      logger,
	}
}

type WithdrawPostRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

func (h *WithdrawPostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	login, ok := claims["username"].(string)
	if !ok {
		h.logger.Warn("Unauthorized access attempt")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
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

	err := h.userService.Withdraw(login, req.Order, req.Sum)
	if err != nil {
		switch {
		case errors.Is(err, gofermartErrors.ErrInsufficientFunds):
			http.Error(w, "Insufficient funds", http.StatusPaymentRequired)
		default:
			h.logger.Error("Failed to process withdrawal", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
