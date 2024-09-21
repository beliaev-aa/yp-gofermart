package user

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"encoding/json"
	"fmt"
	"github.com/go-chi/jwtauth/v5"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type WithdrawalsGetHandler struct {
	userService *services.UserService
	logger      *zap.Logger
}

func NewWithdrawalsGetHandler(userService *services.UserService, logger *zap.Logger) *WithdrawalsGetHandler {
	return &WithdrawalsGetHandler{
		userService: userService,
		logger:      logger,
	}
}

type WithdrawalResponse struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

func (w WithdrawalResponse) MarshalJSON() ([]byte, error) {
	jsonString := fmt.Sprintf(`{"order": "%s", "sum": %.2f, "processed_at": "%s"}`,
		w.Order,
		w.Sum,
		w.ProcessedAt,
	)

	return []byte(jsonString), nil
}

func (h *WithdrawalsGetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	login, ok := claims["username"].(string)
	if !ok {
		h.logger.Warn("Unauthorized access attempt")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	withdrawals, err := h.userService.GetWithdrawals(login)
	if err != nil {
		h.logger.Error("Failed to get withdrawals", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var response []WithdrawalResponse
	for _, wd := range withdrawals {
		item := WithdrawalResponse{
			Order:       wd.Order,
			Sum:         wd.Sum,
			ProcessedAt: wd.ProcessedAt.Format(time.RFC3339),
		}
		response = append(response, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
