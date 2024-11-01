package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(2, 2) // 2 requests per second, burst of 2
	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	makeRequest := func() int {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "127.0.0.1:12345" // Set consistent IP
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		return w.Code
	}

	// First two requests should succeed (within burst limit)
	if code := makeRequest(); code != http.StatusOK {
		t.Errorf("First request: expected 200, got %d", code)
	}
	if code := makeRequest(); code != http.StatusOK {
		t.Errorf("Second request: expected 200, got %d", code)
	}

	// Third request should be rate limited
	if code := makeRequest(); code != http.StatusTooManyRequests {
		t.Errorf("Third request: expected 429, got %d", code)
	}

	// Wait for rate limit to reset
	time.Sleep(time.Second)

	// Should succeed again
	if code := makeRequest(); code != http.StatusOK {
		t.Errorf("After wait: expected 200, got %d", code)
	}
}
