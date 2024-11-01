//go:build !test

package main

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/config"
	"beliaev-aa/yp-gofermart/internal/gofermart/http-server"
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"beliaev-aa/yp-gofermart/internal/gofermart/storage"
	"beliaev-aa/yp-gofermart/internal/gofermart/utils"
	"beliaev-aa/yp-gofermart/internal/gofermart/workers"
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
	orderService := services.NewOrderService(accrualService, store.OrderRepo, store.UserRepo, logger)

	// Создание сервисов приложения
	appServices := &services.AppServices{
		AuthService:  services.NewAuthService([]byte(cfg.JWTSecret), store.UserRepo, logger),
		OrderService: orderService,
		UserService:  services.NewUserService(store.UserRepo, store.WithdrawalRepo, logger),
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
	go workers.StartOrderStatusUpdater(ctx, orderService, logger, &wg)

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
