package config

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	"flag"
	"github.com/caarlos0/env/v10"
)

var (
	FlagAccrualSystemAddress string
	FlagDatabaseURI          string
	FlagRunAddress           string
	FlagJWTSecret            string
	configEnv                = domain.Config{}
)

// parseFlags парсит аргументы командной строки и сохраняет значения в глобальные переменные.
func parseFlags() {
	flag.StringVar(&FlagAccrualSystemAddress, "r", "http://localhost:8080", "Address of the accrual settlement system")
	flag.StringVar(&FlagDatabaseURI, "d", "", "PostgreSQL DSN")
	flag.StringVar(&FlagRunAddress, "a", "localhost:8080", "Address and port to run the HTTP server")
	flag.StringVar(&FlagJWTSecret, "s", "your-256-bit-secret-key", "Your JWT-secret key")
	flag.Parse()
}

// parseEnv парсит переменные окружения и сохраняет значения в configEnv.
func parseEnv() error {
	err := env.Parse(&configEnv)
	if err != nil {
		return err
	}

	return nil
}

// selectCfgFromSource выбирает значение конфигурации между аргументом командной строки и переменной окружения.
func selectCfgFromSource(cfgFlag, cfgEnv *string) string {
	if len(*cfgEnv) > 0 {
		return *cfgEnv
	}

	return *cfgFlag
}

// LoadConfig загружает конфигурацию из переменных окружения и аргументов командной строки.
func LoadConfig() (*domain.Config, error) {
	err := parseEnv()

	if err != nil {
		return nil, err
	}

	parseFlags()

	var config = &domain.Config{}
	config.AccrualSystemAddress = selectCfgFromSource(&FlagAccrualSystemAddress, &configEnv.AccrualSystemAddress)
	config.DatabaseURI = selectCfgFromSource(&FlagDatabaseURI, &configEnv.DatabaseURI)
	config.RunAddress = selectCfgFromSource(&FlagRunAddress, &configEnv.RunAddress)
	config.JWTSecret = selectCfgFromSource(&FlagRunAddress, &configEnv.JWTSecret)

	return config, nil
}
