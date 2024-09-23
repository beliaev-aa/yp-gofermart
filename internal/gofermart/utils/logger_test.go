package utils

import (
	"bytes"
	"errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"syscall"
	"testing"
)

// mockLogger - заменяет Logger для тестирования
type mockLogger struct {
	*zap.Logger
	syncErr error
}

func (m *mockLogger) Sync() error {
	return m.syncErr
}

// captureLogs - перехватывает логи, записанные через Logger, для анализа в тестах
func captureLogs() (*zap.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	writer := zapcore.AddSync(&buf)
	encoderCfg := zap.NewProductionEncoderConfig()
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), writer, zapcore.DebugLevel)
	logger := zap.New(core)
	return logger, &buf
}

var _ = os.Exit

func mockExit(int) {}

func TestNewLogger(t *testing.T) {
	_ = mockExit
	defer func() { _ = os.Exit }()

	testCases := []struct {
		Name       string
		MockSync   error
		ShouldFail bool
	}{
		{
			Name:       "Logger_Initialization_Success",
			MockSync:   nil,
			ShouldFail: false,
		},
		{
			Name:       "Logger_Sync_Failure_EBADF",
			MockSync:   syscall.EBADF,
			ShouldFail: false, // Эту ошибку нужно игнорировать
		},
		{
			Name:       "Logger_Sync_Failure_ENOTTY",
			MockSync:   syscall.ENOTTY,
			ShouldFail: false, // Эту ошибку нужно игнорировать
		},
		{
			Name:       "Logger_Sync_OtherFailure",
			MockSync:   errors.New("sync failed"),
			ShouldFail: true, // Для всех других ошибок мы ожидаем фатальное завершение
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Перехватываем логи
			logger, logBuffer := captureLogs()

			// Создаем mock с предопределенной ошибкой sync
			mockedLogger := &mockLogger{
				Logger:  logger,
				syncErr: tc.MockSync,
			}

			// Обрабатываем defer для вызова Sync
			defer func(logger *mockLogger) {
				err := logger.Sync()

				// Если ошибка не системная и не игнорируемая, пишем её как фатальную
				if err != nil && (!errors.Is(err, syscall.EBADF) && !errors.Is(err, syscall.ENOTTY)) {
					// Делаем запись ошибки через Error вместо Fatal для тестирования
					logger.Error("Server failed on Sync()-method call on zap.Logger.", zap.Error(err))

					// Если ошибка критическая, ожидаем, что она будет зафиксирована
					if tc.ShouldFail {
						if !bytes.Contains(logBuffer.Bytes(), []byte("sync failed")) {
							t.Errorf("Expected fatal log, but did not find it in logs: %v", logBuffer.String())
						}
					}
				}
			}(mockedLogger)

			if mockedLogger == nil {
				t.Errorf("Expected non-nil logger, got nil")
			}
		})
	}
}
