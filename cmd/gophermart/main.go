//go:build !test

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
	// Инициализация сервиса логирования
	logger := utils.NewLogger()

	// Загрузка конфигурации приложения
	cfg, err := config.LoadConfig()
	if err != nil {
		// Завершение работы приложения с ошибкой, если конфигурация не загружена
		logger.Fatal("Server failed to load configuration.", zap.Error(err))
	}

	// Создание подключения к базе данных
	store, err := storage.NewStorage(cfg.DatabaseURI, logger)
	if err != nil {
		// Завершение работы приложения с ошибкой при подключении к базе данных
		logger.Fatal("Failed to connect to database.", zap.Error(err))
	}

	// Инициализация сервиса для работы с внешним сервисом заказов
	accrualService := services.NewAccrualService(cfg.AccrualSystemAddress, logger)

	// Инициализация сервиса для работы с заказами
	orderService := services.NewOrderService(accrualService, store, logger)

	// Запуск фонового процесса для обновления статусов заказов
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		// Используем for range для перебора значений из канала тикера
		for range ticker.C {
			// Обновление статусов заказов каждую секунду
			orderService.UpdateOrderStatuses()
		}
	}()

	// Создание сервисов приложения: аутентификация, заказы и пользователи
	appServices := &services.AppServices{
		AuthService:  services.NewAuthService([]byte(cfg.JWTSecret), logger, store),
		OrderService: orderService,
		UserService:  services.NewUserService(store, logger),
	}

	// Инициализация роутера Chi
	r := chi.NewRouter()

	// Регистрация маршрутов приложения
	httpserver.RegisterRoutes(r, appServices, logger)

	// Логирование запуска сервера
	logger.Info("Starting server on " + cfg.RunAddress)

	// Запуск HTTP-сервера
	err = http.ListenAndServe(cfg.RunAddress, r)
	if err != nil {
		// Завершение работы приложения с ошибкой при старте сервера
		logger.Fatal("Server failed to start", zap.Error(err))
	}
}
