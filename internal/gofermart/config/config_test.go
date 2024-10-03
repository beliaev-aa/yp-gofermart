package config

import (
	"errors"
	"flag"
	"github.com/google/go-cmp/cmp"
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	type testCase struct {
		name           string
		envVariables   map[string]string
		args           []string
		expectedConfig *Config
		expectedErr    error
	}

	testCases := []testCase{
		{
			name: "Env_Variables_Have_Priority_Over_Flags",
			envVariables: map[string]string{
				"RUN_ADDRESS":            "env_run_address",
				"DATABASE_URI":           "env_database_uri",
				"ACCRUAL_SYSTEM_ADDRESS": "env_accrual_system_address",
				"JWT_SECRET":             "env_jwt_secret",
			},
			args: []string{
				"-a", "flag_run_address",
				"-d", "flag_database_uri",
				"-r", "flag_accrual_system_address",
				"-s", "flag_jwt_secret",
			},
			expectedConfig: &Config{
				RunAddress:           "env_run_address",
				DatabaseURI:          "env_database_uri",
				AccrualSystemAddress: "env_accrual_system_address",
				JWTSecret:            "env_jwt_secret",
			},
		},
		{
			name:         "Flags_Used_If_Env_Variables_Not_Set",
			envVariables: map[string]string{},
			args: []string{
				"-a", "flag_run_address",
				"-d", "flag_database_uri",
				"-r", "flag_accrual_system_address",
				"-s", "flag_jwt_secret",
			},
			expectedConfig: &Config{
				RunAddress:           "flag_run_address",
				DatabaseURI:          "flag_database_uri",
				AccrualSystemAddress: "flag_accrual_system_address",
				JWTSecret:            "flag_jwt_secret",
			},
		},
		{
			name: "Config_Validation_Error_If_Database_URI_Not_Set",
			envVariables: map[string]string{
				"RUN_ADDRESS":            "env_run_address",
				"ACCRUAL_SYSTEM_ADDRESS": "env_accrual_system_address",
				"JWT_SECRET":             "env_jwt_secret",
			},
			args: []string{
				"-a", "flag_run_address",
				"-r", "flag_accrual_system_address",
				"-s", "flag_jwt_secret",
			},
			expectedConfig: nil,
			expectedErr:    ErrDatabaseConfig,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Clearenv()

			for key, value := range tc.envVariables {
				if err := os.Setenv(key, value); err != nil {
					t.Fatalf("failed to set env variable: %s", key)
				}
			}

			os.Args = append([]string{"cmd"}, tc.args...)

			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

			cfg, err := LoadConfig()

			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("expected error: %v, got: %v", tc.expectedErr, err)
			}

			if tc.expectedErr != nil {
				if diff := cmp.Diff(tc.expectedConfig, cfg); diff != "" {
					t.Fatalf("Unexpected config (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	type testCase struct {
		name        string
		config      *Config
		expectedErr error
	}

	testCases := []testCase{
		{
			name: "Valid_Config",
			config: &Config{
				RunAddress:           "localhost:8080",
				DatabaseURI:          "postgres://user:pass@localhost/db",
				AccrualSystemAddress: "http://localhost:8080",
			},
			expectedErr: nil,
		},
		{
			name: "Missing_AccrualSystemAddress",
			config: &Config{
				RunAddress:  "localhost:8080",
				DatabaseURI: "postgres://user:pass@localhost/db",
			},
			expectedErr: ErrAccrualConfig,
		},
		{
			name: "Missing_DatabaseURI",
			config: &Config{
				RunAddress:           "localhost:8080",
				AccrualSystemAddress: "http://localhost:8080",
			},
			expectedErr: ErrDatabaseConfig,
		},
		{
			name: "Missing_RunAddress",
			config: &Config{
				DatabaseURI:          "postgres://user:pass@localhost/db",
				AccrualSystemAddress: "http://localhost:8080",
			},
			expectedErr: ErrRunAddressConfig,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateConfig(tc.config)

			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("expected error: %v, got: %v", tc.expectedErr, err)
			}
		})
	}
}
