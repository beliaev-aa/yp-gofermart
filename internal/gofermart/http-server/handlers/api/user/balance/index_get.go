package balance

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"beliaev-aa/yp-gofermart/internal/gofermart/utils"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"net/http"
)

type IndexGetHandler struct {
	logger            *zap.Logger
	userService       *services.UserService
	usernameExtractor utils.UsernameExtractor
}

func NewIndexGetHandler(userService *services.UserService, usernameExtractor utils.UsernameExtractor, logger *zap.Logger) *IndexGetHandler {
	return &IndexGetHandler{
		logger:            logger,
		userService:       userService,
		usernameExtractor: usernameExtractor,
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
