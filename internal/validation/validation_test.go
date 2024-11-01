package validation

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateRequest(t *testing.T) {
	tests := []struct {
		name          string
		method        string
		expectedError bool
	}{
		{
			name:          "valid GET request",
			method:        http.MethodGet,
			expectedError: false,
		},
		{
			name:          "invalid POST request",
			method:        http.MethodPost,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/test", nil)
			err := ValidateRequest(req)

			if tt.expectedError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
