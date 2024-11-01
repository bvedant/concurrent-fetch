package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Test default values
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", cfg.Server.Port)
	}

	// Test environment variable override
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("API_RETRY_ATTEMPTS", "5")
	defer func() {
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("API_RETRY_ATTEMPTS")
	}()

	cfg, err = Load()
	if err != nil {
		t.Fatalf("Failed to load config with env vars: %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", cfg.Server.Port)
	}
	if cfg.API.RetryAttempts != 5 {
		t.Errorf("Expected 5 retry attempts, got %d", cfg.API.RetryAttempts)
	}
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name      string
		envVars   map[string]string
		expectErr bool
	}{
		{
			name: "valid config",
			envVars: map[string]string{
				"SERVER_PORT": "8080",
			},
			expectErr: false,
		},
		{
			name: "invalid port",
			envVars: map[string]string{
				"SERVER_PORT": "99999",
			},
			expectErr: true,
		},
		{
			name: "invalid retry attempts",
			envVars: map[string]string{
				"API_RETRY_ATTEMPTS": "-1",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer func() {
				// Clean up environment variables
				for k := range tt.envVars {
					os.Unsetenv(k)
				}
			}()

			_, err := Load()
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
