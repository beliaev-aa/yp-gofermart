package config

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
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
		expectedConfig *domain.Config
		expectedErr    bool
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
			expectedConfig: &domain.Config{
				RunAddress:           "env_run_address",
				DatabaseURI:          "env_database_uri",
				AccrualSystemAddress: "env_accrual_system_address",
				JWTSecret:            "env_jwt_secret",
			},
			expectedErr: false,
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
			expectedConfig: &domain.Config{
				RunAddress:           "flag_run_address",
				DatabaseURI:          "flag_database_uri",
				AccrualSystemAddress: "flag_accrual_system_address",
				JWTSecret:            "flag_jwt_secret",
			},
			expectedErr: false,
		},
		{
			name:         "Default_Values_Used_If_Neither_Flags_Or_Env_Set",
			envVariables: map[string]string{},
			args:         []string{},
			expectedConfig: &domain.Config{
				RunAddress:           "localhost:8080",
				DatabaseURI:          "",
				AccrualSystemAddress: "http://localhost:8080",
				JWTSecret:            "your-256-bit-secret-key",
			},
			expectedErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Clearenv()
			os.Args = append([]string{"cmd"}, tc.args...)

			for key, value := range tc.envVariables {
				if err := os.Setenv(key, value); err != nil {
					t.Fatalf("failed to set env variable: %s", key)
				}
			}

			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

			cfg, err := LoadConfig()

			if (err != nil) != tc.expectedErr {
				t.Errorf("expected error: %v, got: %v", tc.expectedErr, err)
			}

			if diff := cmp.Diff(tc.expectedConfig, cfg); diff != "" {
				t.Fatalf("Unexpected config (-want +got):\n%s", diff)
			}
		})
	}
}
