package config

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	"errors"
	"flag"
	"os"
)

const (
	defaultAccrualSystemAddress = "http://localhost:8080"
	defaultDatabaseURI          = ""
	defaultJWTSecret            = "your-256-bit-secret-key"
	defaultRunAddress           = "localhost:8080"
)

var (
	ErrAccrualConfig    = errors.New("AccrualSystemAddress is not configured")
	ErrDatabaseConfig   = errors.New("DatabaseURI is not configured")
	ErrRunAddressConfig = errors.New("RunAddress is not configured")
)

// LoadConfig - загружает конфигурацию, отдает приоритет переменным окружения
func LoadConfig() (*domain.Config, error) {
	cfg := &domain.Config{}

	flag.StringVar(&cfg.RunAddress, "a", defaultRunAddress, "Address and port to run the HTTP server")
	flag.StringVar(&cfg.DatabaseURI, "d", defaultDatabaseURI, "PostgreSQL DSN")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", defaultAccrualSystemAddress, "Address of the accrual settlement system")
	flag.StringVar(&cfg.JWTSecret, "s", defaultJWTSecret, "Your JWT-secret key")
	flag.Parse()

	if envRunAddress := os.Getenv("RUN_ADDRESS"); envRunAddress != "" {
		cfg.RunAddress = envRunAddress
	}

	if envDatabaseURI := os.Getenv("DATABASE_URI"); envDatabaseURI != "" {
		cfg.DatabaseURI = envDatabaseURI
	}

	if envAccrualAddress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualAddress != "" {
		cfg.AccrualSystemAddress = envAccrualAddress
	}

	if envJWTSecret := os.Getenv("JWT_SECRET"); envJWTSecret != "" {
		cfg.JWTSecret = envJWTSecret
	}

	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validateConfig - проверяет обязательные параметры конфигурации
func validateConfig(cfg *domain.Config) error {
	if cfg.AccrualSystemAddress == "" {
		return ErrAccrualConfig
	}
	if cfg.DatabaseURI == "" {
		return ErrDatabaseConfig
	}
	if cfg.RunAddress == "" {
		return ErrRunAddressConfig
	}
	return nil
}
