package config

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	"flag"
	"github.com/google/go-cmp/cmp"
	"os"
	"testing"
)

// Тест для функции selectCfgFromSource
func TestSelectCfgFromSource(t *testing.T) {
	// Определяем структуру теста
	type testCase struct {
		name          string
		flagValue     string
		envValue      string
		expectedValue string
	}

	// Список тестов
	testCases := []testCase{
		{
			name:          "Env_value_has_priority",
			flagValue:     "flag-value",
			envValue:      "env-value",
			expectedValue: "env-value",
		},
		{
			name:          "Flag_value_is_used_when_env_is_empty",
			flagValue:     "flag-value",
			envValue:      "",
			expectedValue: "flag-value",
		},
		{
			name:          "Empty_flag_and_env",
			flagValue:     "",
			envValue:      "",
			expectedValue: "",
		},
	}

	// Запуск каждого теста
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := selectCfgFromSource(tc.flagValue, tc.envValue)
			if result != tc.expectedValue {
				t.Errorf("expected %v, got %v", tc.expectedValue, result)
			}
		})
	}
}

// Тест для функции LoadConfig
func TestLoadConfig(t *testing.T) {
	// Определяем структуру теста
	type testCase struct {
		name           string
		flagConfig     *domain.Config
		envConfig      *domain.Config
		expectedConfig *domain.Config
	}

	// Функция для имитации загрузки переменных окружения
	mockEnvParse := func(envConfig *domain.Config) {
		*envConfig = domain.Config{
			RunAddress:           "env-run-address",
			DatabaseURI:          "env-database-uri",
			AccrualSystemAddress: "env-accrual-address",
			JWTSecret:            "env-jwt-secret",
		}
	}

	// Создаем тестовые конфигурации флагов
	mockFlagConfig := &domain.Config{
		RunAddress:           "flag-run-address",
		DatabaseURI:          "flag-database-uri",
		AccrualSystemAddress: "flag-accrual-address",
		JWTSecret:            "flag-jwt-secret",
	}

	// Ожидаемая конфигурация
	expectedConfig := &domain.Config{
		RunAddress:           "env-run-address",     // из env
		DatabaseURI:          "env-database-uri",    // из env
		AccrualSystemAddress: "env-accrual-address", // из env
		JWTSecret:            "env-jwt-secret",      // из env
	}

	// Тест кейсы
	testCases := []testCase{
		{
			name:           "Env values have priority",
			flagConfig:     mockFlagConfig,
			envConfig:      &domain.Config{}, // начальная пустая конфигурация для env
			expectedConfig: expectedConfig,   // ожидаемый результат
		},
	}

	// Запуск тестов
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockEnvParse(tc.envConfig)

			// Запускаем тестируемую функцию с mock-данными
			config := &domain.Config{
				RunAddress:           selectCfgFromSource(tc.flagConfig.RunAddress, tc.envConfig.RunAddress),
				DatabaseURI:          selectCfgFromSource(tc.flagConfig.DatabaseURI, tc.envConfig.DatabaseURI),
				AccrualSystemAddress: selectCfgFromSource(tc.flagConfig.AccrualSystemAddress, tc.envConfig.AccrualSystemAddress),
				JWTSecret:            selectCfgFromSource(tc.flagConfig.JWTSecret, tc.envConfig.JWTSecret),
			}

			if diff := cmp.Diff(tc.expectedConfig, config); diff != "" {
				t.Fatalf("Unexpected config (-want +got):\n%s", diff)
			}
		})
	}
}

// Тест для функции parseFlags
func TestParseFlags(t *testing.T) { // Сбрасываем флаги перед тестом, чтобы избежать повторного объявления
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Сброс флагов для тестов
	os.Args = []string{"cmd", "-r", "http://test-accrual", "-d", "test-dsn", "-a", "localhost:9000", "-s", "test-secret"}

	// Ожидаемые значения
	expected := &domain.Config{
		AccrualSystemAddress: "http://test-accrual",
		DatabaseURI:          "test-dsn",
		RunAddress:           "localhost:9000",
		JWTSecret:            "test-secret",
	}

	// Парсинг флагов
	config := parseFlags()

	if diff := cmp.Diff(expected, config); diff != "" {
		t.Fatalf("Unexpected config (-want +got):\n%s", diff)
	}
}

// Тест для функции LoadConfig
func TestLoadConfigArgs(t *testing.T) {
	// Сбрасываем флаги перед тестом, чтобы избежать повторного объявления
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Настройка окружения для тестов
	err := os.Setenv("RUN_ADDRESS", "env-run-address")
	if err != nil {
		return
	}
	err = os.Setenv("DATABASE_URI", "env-database-uri")
	if err != nil {
		return
	}
	err = os.Setenv("ACCRUAL_SYSTEM_ADDRESS", "env-accrual-address")
	if err != nil {
		return
	}
	err = os.Setenv("JWT_SECRET", "env-jwt-secret")
	if err != nil {
		return
	}

	// Сброс флагов для тестов
	os.Args = []string{"cmd", "-r", "http://flag-accrual", "-d", "flag-dsn", "-a", "localhost:9000", "-s", "flag-secret"}

	// Ожидаемая конфигурация, в которой переменные окружения имеют приоритет
	expected := &domain.Config{
		RunAddress:           "env-run-address",
		DatabaseURI:          "env-database-uri",
		AccrualSystemAddress: "env-accrual-address",
		JWTSecret:            "env-jwt-secret",
	}

	// Загружаем конфигурацию
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned an error: %v", err)
	}

	if diff := cmp.Diff(expected, config); diff != "" {
		t.Fatalf("Unexpected config (-want +got):\n%s", diff)
	}

	// Удаляем переменные окружения после теста
	err = os.Unsetenv("RUN_ADDRESS")
	if err != nil {
		return
	}
	err = os.Unsetenv("DATABASE_URI")
	if err != nil {
		return
	}
	err = os.Unsetenv("ACCRUAL_SYSTEM_ADDRESS")
	if err != nil {
		return
	}
	err = os.Unsetenv("JWT_SECRET")
	if err != nil {
		return
	}
}
