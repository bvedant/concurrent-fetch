package health

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthChecker(t *testing.T) {
	h := NewHealthChecker()

	// Add test checks
	h.AddCheck("passing", func() error {
		return nil
	})
	h.AddCheck("failing", func() error {
		return errors.New("test error")
	})

	// Create test request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Execute handler
	h.Handler().ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}

	var response HealthStatus
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != "error" {
		t.Errorf("Expected status 'error', got '%s'", response.Status)
	}

	if response.Checks["passing"] != "ok" {
		t.Errorf("Expected passing check to be 'ok', got '%s'", response.Checks["passing"])
	}

	if response.Checks["failing"] != "test error" {
		t.Errorf("Expected failing check to be 'test error', got '%s'", response.Checks["failing"])
	}
}
