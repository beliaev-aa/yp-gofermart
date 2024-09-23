package config

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	"flag"
	"github.com/caarlos0/env/v10"
)

// parseFlags - парсит флаги и возвращает конфигурацию
func parseFlags() *domain.Config {
	cfg := &domain.Config{}
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "http://localhost:8080", "Address of the accrual settlement system")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "PostgreSQL DSN")
	flag.StringVar(&cfg.RunAddress, "a", "localhost:8080", "Address and port to run the HTTP server")
	flag.StringVar(&cfg.JWTSecret, "s", "your-256-bit-secret-key", "Your JWT-secret key")
	flag.Parse()
	return cfg
}

// LoadConfig - загружает конфигурацию, отдает приоритет переменным окружения
func LoadConfig() (*domain.Config, error) {
	// Загружаем переменные окружения
	envConfig := &domain.Config{}
	err := env.Parse(envConfig)
	if err != nil {
		return nil, err // coverage:ignore
	}

	// Загружаем флаги командной строки
	flagConfig := parseFlags()

	// Создаем финальную конфигурацию, выбирая значения с приоритетом у переменных окружения
	finalConfig := &domain.Config{
		RunAddress:           selectCfgFromSource(flagConfig.RunAddress, envConfig.RunAddress),
		DatabaseURI:          selectCfgFromSource(flagConfig.DatabaseURI, envConfig.DatabaseURI),
		AccrualSystemAddress: selectCfgFromSource(flagConfig.AccrualSystemAddress, envConfig.AccrualSystemAddress),
		JWTSecret:            selectCfgFromSource(flagConfig.JWTSecret, envConfig.JWTSecret),
	}

	return finalConfig, nil
}

// selectCfgFromSource - выбирает значение из двух источников
func selectCfgFromSource(flagValue, envValue string) string {
	if envValue != "" {
		return envValue
	}
	return flagValue
}
