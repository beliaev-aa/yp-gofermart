package config

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	"flag"
	"github.com/caarlos0/env/v10"
)

var (
	configEnv = domain.Config{}
)

func parseFlags() *domain.Config {
	cfg := &domain.Config{}
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "http://localhost:8080", "Address of the accrual settlement system")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "PostgreSQL DSN")
	flag.StringVar(&cfg.RunAddress, "a", "localhost:8080", "Address and port to run the HTTP server")
	flag.StringVar(&cfg.JWTSecret, "s", "your-256-bit-secret-key", "Your JWT-secret key")
	flag.Parse()
	return cfg
}

func LoadConfig() (*domain.Config, error) {
	err := env.Parse(&configEnv)
	if err != nil {
		return nil, err
	}

	cfgFlags := parseFlags()

	// Приоритет у флагов командной строки
	config := &domain.Config{
		RunAddress:           selectCfgFromSource(&cfgFlags.RunAddress, &configEnv.RunAddress),
		DatabaseURI:          selectCfgFromSource(&cfgFlags.DatabaseURI, &configEnv.DatabaseURI),
		AccrualSystemAddress: selectCfgFromSource(&cfgFlags.AccrualSystemAddress, &configEnv.AccrualSystemAddress),
		JWTSecret:            selectCfgFromSource(&cfgFlags.JWTSecret, &configEnv.JWTSecret),
	}

	return config, nil
}

func selectCfgFromSource(cfgFlag, cfgEnv *string) string {
	if len(*cfgFlag) > 0 {
		return *cfgFlag
	}
	return *cfgEnv
}
