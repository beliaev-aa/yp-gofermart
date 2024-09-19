package main

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/config"
	"beliaev-aa/yp-gofermart/internal/gofermart/http-server"
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"beliaev-aa/yp-gofermart/internal/gofermart/storage"
	"beliaev-aa/yp-gofermart/internal/gofermart/utils"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func main() {
	logger := utils.NewLogger()
	cfg, err := config.LoadConfig()

	if err != nil {
		logger.Fatal("Server failed to load configuration.", zap.Error(err))
	}

	store, err := storage.NewPostgresStorage(cfg.DatabaseURI, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database.", zap.Error(err))
	}
	defer store.Close()

	orderService := services.NewOrderService(cfg.AccrualSystemAddress, store, logger)
	// Запуск фонового процесса для обновления статусов заказов
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				orderService.UpdateOrderStatuses()
			}
		}
	}()

	appServices := &services.AppServices{
		AuthService:  services.NewAuthService([]byte(cfg.JWTSecret), logger, store),
		OrderService: orderService,
		UserService:  services.NewUserService(store, logger),
	}

	r := chi.NewRouter()
	httpserver.RegisterRoutes(r, appServices, logger)

	logger.Info("Starting server on " + cfg.RunAddress)
	err = http.ListenAndServe(cfg.RunAddress, r)
	if err != nil {
		logger.Fatal("Server failed to start", zap.Error(err))
	}
}
