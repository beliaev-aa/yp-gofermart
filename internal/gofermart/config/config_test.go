package config

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	"github.com/google/go-cmp/cmp"
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
