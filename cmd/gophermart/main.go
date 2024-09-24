//go:build !test

package main

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/config"
	"beliaev-aa/yp-gofermart/internal/gofermart/http-server"
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"beliaev-aa/yp-gofermart/internal/gofermart/storage"
	"beliaev-aa/yp-gofermart/internal/gofermart/utils"
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
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

	// Создание сервисов приложения
	appServices := &services.AppServices{
		AuthService:  services.NewAuthService([]byte(cfg.JWTSecret), logger, store),
		OrderService: orderService,
		UserService:  services.NewUserService(store, logger),
	}

	// Инициализация роутера Chi
	r := chi.NewRouter()
	httpserver.RegisterRoutes(r, appServices, logger)

	// Логирование запуска сервера
	logger.Info("Starting server on " + cfg.RunAddress)

	// Создаем канал для захвата системных сигналов
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	// Создаем контекст для управления goroutine обновления статусов заказов
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// Запуск фонового процесса для обновления статусов заказов
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		// Канал для отслеживания обработки
		processingOrders := make(chan struct{}, 1)

		for {
			select {
			case <-ticker.C:
				select {
				case processingOrders <- struct{}{}:
					orderService.UpdateOrderStatuses()
					<-processingOrders
				default:
					// Избегаем запуска новой обработки, если текущая не завершена
					logger.Warn("Skipping order update due to ongoing processing")
				}
			case <-ctx.Done():
				logger.Info("Waiting for ongoing order status updates to complete")
				processingOrders <- struct{}{}
				logger.Info("Shutting down order status update goroutine")
				return
			}
		}
	}()

	// Запуск HTTP-сервера в отдельной goroutine
	server := &http.Server{
		Addr:    cfg.RunAddress,
		Handler: r,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	<-stopChan
	logger.Info("Shutting down server...")

	cancel()

	wg.Wait()

	// Завершаем работу HTTP-сервера с таймаутом для завершения текущих запросов
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := server.Shutdown(ctxShutdown); err != nil {
		logger.Fatal("Server shutdown failed", zap.Error(err))
	} else {
		logger.Info("Server gracefully stopped")
	}
}
