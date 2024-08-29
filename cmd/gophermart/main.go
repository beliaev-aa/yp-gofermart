package main

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/config"
	"beliaev-aa/yp-gofermart/internal/gofermart/http-server"
	"beliaev-aa/yp-gofermart/internal/gofermart/utils"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
)

func main() {
	logger := utils.NewLogger()
	cfg, err := config.LoadConfig()

	if err != nil {
		logger.Fatal("Server failed to load configuration.", zap.Error(err))
	}

	r := chi.NewRouter()
	httpserver.RegisterRoutes(r, cfg, logger)

	logger.Info("Starting server on :8080")
	err = http.ListenAndServe(cfg.RunAddress, r)
	if err != nil {
		logger.Fatal("Server failed to start", zap.Error(err))
	}
}
