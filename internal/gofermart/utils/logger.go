package utils

import (
	"errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"syscall"
)

// NewLogger - создает новый экземпляр Logger с конфигурацией для вывода в stdout.
func NewLogger() *zap.Logger {
	// Настройка конфигурации для вывода в stdout
	config := zap.NewProductionConfig()

	// Изменение выходного потока на stdout
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stdout"}

	// Задание формата времени в ISO8601 для читаемости
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Создание сервиса логирования с конфигурацией
	logger, err := config.Build()
	if err != nil {
		// В случае ошибки, выходим с фатальной ошибкой и делаем запись в лог
		zap.L().Fatal("Server failed to create logger instance", zap.Error(err))
	}

	// Отложенная функция для синхронизации буферов записи
	defer func(logger *zap.Logger) {
		// Выполнение синхронизации (например, для flush)
		err := logger.Sync()

		// Обработка ошибок, игнорируем некоторые системные ошибки (например, EBADF или ENOTTY)
		if err != nil && (!errors.Is(err, syscall.EBADF) && !errors.Is(err, syscall.ENOTTY)) {
			// В случае серьёзной ошибки, делаем запись в лог
			logger.Fatal("Server failed on Sync()-method call on zap.Logger.", zap.Error(err))
		}
	}(logger)

	return logger
}
