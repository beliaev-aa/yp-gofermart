package config

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	"flag"
	"github.com/google/go-cmp/cmp"
	"os"
	"testing"
)

func TestSelectCfgFromSource(t *testing.T) {
	type testCase struct {
		name          string
		flagValue     string
		envValue      string
		expectedValue string
	}

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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := selectCfgFromSource(tc.flagValue, tc.envValue)
			if result != tc.expectedValue {
				t.Errorf("expected %v, got %v", tc.expectedValue, result)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	type testCase struct {
		name           string
		flagConfig     *domain.Config
		envConfig      *domain.Config
		expectedConfig *domain.Config
	}

	mockEnvParse := func(envConfig *domain.Config) {
		*envConfig = domain.Config{
			RunAddress:           "env-run-address",
			DatabaseURI:          "env-database-uri",
			AccrualSystemAddress: "env-accrual-address",
			JWTSecret:            "env-jwt-secret",
		}
	}

	mockFlagConfig := &domain.Config{
		RunAddress:           "flag-run-address",
		DatabaseURI:          "flag-database-uri",
		AccrualSystemAddress: "flag-accrual-address",
		JWTSecret:            "flag-jwt-secret",
	}

	expectedConfig := &domain.Config{
		RunAddress:           "env-run-address",
		DatabaseURI:          "env-database-uri",
		AccrualSystemAddress: "env-accrual-address",
		JWTSecret:            "env-jwt-secret",
	}

	testCases := []testCase{
		{
			name:           "Env values have priority",
			flagConfig:     mockFlagConfig,
			envConfig:      &domain.Config{},
			expectedConfig: expectedConfig,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockEnvParse(tc.envConfig)

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

func TestParseFlags(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Args = []string{"cmd", "-r", "http://test-accrual", "-d", "test-dsn", "-a", "localhost:9000", "-s", "test-secret"}

	expected := &domain.Config{
		AccrualSystemAddress: "http://test-accrual",
		DatabaseURI:          "test-dsn",
		RunAddress:           "localhost:9000",
		JWTSecret:            "test-secret",
	}

	config := parseFlags()

	if diff := cmp.Diff(expected, config); diff != "" {
		t.Fatalf("Unexpected config (-want +got):\n%s", diff)
	}
}

func TestLoadConfigArgs(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

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

	os.Args = []string{"cmd", "-r", "http://flag-accrual", "-d", "flag-dsn", "-a", "localhost:9000", "-s", "flag-secret"}

	expected := &domain.Config{
		RunAddress:           "env-run-address",
		DatabaseURI:          "env-database-uri",
		AccrualSystemAddress: "env-accrual-address",
		JWTSecret:            "env-jwt-secret",
	}

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned an error: %v", err)
	}

	if diff := cmp.Diff(expected, config); diff != "" {
		t.Fatalf("Unexpected config (-want +got):\n%s", diff)
	}

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
