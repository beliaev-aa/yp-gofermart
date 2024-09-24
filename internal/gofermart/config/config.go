package config

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	"flag"
	"github.com/caarlos0/env/v10"
)

const (
	defaultRunAddress           = "localhost:8080"
	defaultDatabaseURI          = ""
	defaultAccrualSystemAddress = "http://localhost:8080"
	defaultJWTSecret            = "your-256-bit-secret-key"
)

// LoadConfig - загружает конфигурацию, отдает приоритет переменным окружения
func LoadConfig() (*domain.Config, error) {
	runAddressFlag := flag.String("a", defaultRunAddress, "Address and port to run the HTTP server")
	databaseURIFlag := flag.String("d", defaultDatabaseURI, "PostgreSQL DSN")
	accrualSystemAddressFlag := flag.String("r", defaultAccrualSystemAddress, "Address of the accrual settlement system")
	jwtSecretFlag := flag.String("s", defaultJWTSecret, "Your JWT-secret key")

	flag.Parse()

	cfg := domain.Config{}
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	if cfg.RunAddress == defaultRunAddress {
		cfg.RunAddress = *runAddressFlag
	}
	if cfg.DatabaseURI == defaultDatabaseURI {
		cfg.DatabaseURI = *databaseURIFlag
	}
	if cfg.AccrualSystemAddress == defaultAccrualSystemAddress {
		cfg.AccrualSystemAddress = *accrualSystemAddressFlag
	}
	if cfg.JWTSecret == defaultJWTSecret {
		cfg.JWTSecret = *jwtSecretFlag
	}

	return &cfg, nil
}
