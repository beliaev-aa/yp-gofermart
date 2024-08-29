package utils

import (
	"errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"syscall"
)

// NewLogger создает новый экземпляр Logger с конфигурацией для продакшн-окружения.
// Logger будет использовать ISO8601 формат времени для логов.
func NewLogger() *zap.Logger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, err := config.Build()

	if err != nil {
		zap.L().Fatal("Server failed to create logger instance", zap.Error(err))
	}

	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil && (!errors.Is(err, syscall.EBADF) && !errors.Is(err, syscall.ENOTTY)) {
			logger.Fatal("Server failed on called Sync()-method on zap.Logger.", zap.Error(err))
		}
	}(logger)

	return logger
}
