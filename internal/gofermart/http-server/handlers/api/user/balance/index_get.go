package balance

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"encoding/json"
	"fmt"
	"github.com/go-chi/jwtauth/v5"
	"go.uber.org/zap"
	"net/http"
)

type IndexGetHandler struct {
	userService *services.UserService
	logger      *zap.Logger
}

func NewIndexGetHandler(userService *services.UserService, logger *zap.Logger) *IndexGetHandler {
	return &IndexGetHandler{
		userService: userService,
		logger:      logger,
	}
}

type IndexGetResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

func (r IndexGetResponse) MarshalJSON1() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"current": %.2f, "withdrawn": %.2f}`, r.Current, r.Withdrawn)), nil
}

func (h *IndexGetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	login, ok := claims["username"].(string)
	if !ok {
		h.logger.Warn("Unauthorized access attempt")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	balance, err := h.userService.GetBalance(login)
	if err != nil {
		h.logger.Error("Failed to get balance", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := IndexGetResponse{
		Current:   balance.Current,
		Withdrawn: balance.Withdrawn,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
