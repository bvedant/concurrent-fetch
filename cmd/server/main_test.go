package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bvedant/concurrent-fetch/internal/cache"
	"github.com/bvedant/concurrent-fetch/internal/fetcher"
	"github.com/bvedant/concurrent-fetch/internal/middleware"
	"github.com/bvedant/concurrent-fetch/internal/models"
)

func TestBuildResponse(t *testing.T) {
	tests := []struct {
		name     string
		results  []models.Result
		expected []map[string]interface{}
	}{
		{
			name: "successful results",
			results: []models.Result{
				{
					Data: []byte(`{"name": "test1", "value": 1}`),
				},
				{
					Data: []byte(`{"name": "test2", "value": 2}`),
				},
			},
			expected: []map[string]interface{}{
				{"name": "test1", "value": float64(1)},
				{"name": "test2", "value": float64(2)},
			},
		},
		{
			name: "mixed results with errors",
			results: []models.Result{
				{
					Data: []byte(`{"name": "test1"}`),
				},
				{
					Error: &fetcher.APIError{
						StatusCode: 404,
						Message:    "Not Found",
						URL:        "http://test.com",
					},
				},
				{
					Data: []byte(`invalid json`),
				},
			},
			expected: []map[string]interface{}{
				{"name": "test1"},
				{
					"error":       "Not Found",
					"status_code": 404,
					"url":         "http://test.com",
				},
				{
					"error": "Failed to parse data",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := buildResponse(tt.results)

			if len(response) != len(tt.expected) {
				t.Errorf("Expected %d results, got %d", len(tt.expected), len(response))
				return
			}

			for i, exp := range tt.expected {
				for k, v := range exp {
					if response[i][k] != v {
						t.Errorf("Expected %v for key %s, got %v", v, k, response[i][k])
					}
				}
			}
		})
	}
}

func TestProcessHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		setupCache     func(*cache.Cache)
	}{
		{
			name:           "invalid method",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "successful GET",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			setupCache: func(c *cache.Cache) {
				// Pre-populate cache with test data
				c.Set("https://api.github.com/repos/golang/go",
					[]byte(`{"name": "go", "stars": 100000}`))
				c.Set("https://api.github.com/repos/kubernetes/kubernetes",
					[]byte(`{"name": "kubernetes", "stars": 90000}`))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test cache
			testCache := cache.NewCache(5 * time.Minute)
			if tt.setupCache != nil {
				tt.setupCache(testCache)
			}

			// Create test request
			req := httptest.NewRequest(tt.method, "/process", nil)
			w := httptest.NewRecorder()

			// Create and execute handler
			handler := processHandler(testCache)
			handler.ServeHTTP(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, w.Code)
			}

			// For successful requests, verify response structure
			if tt.expectedStatus == http.StatusOK {
				var response []map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
					return
				}

				if len(response) != 2 {
					t.Errorf("Expected 2 results, got %d", len(response))
				}
			}
		})
	}
}

// TestRateLimiting tests the rate limiter middleware
func TestRateLimiting(t *testing.T) {
	rateLimiter := middleware.NewRateLimiter(2, 2)
	// Use a simple handler that always returns 200 OK
	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := rateLimiter.Middleware(baseHandler)

	makeRequest := func() int {
		req := httptest.NewRequest("GET", "/process", nil)
		req.RemoteAddr = "192.168.1.1:12345"
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

// TestGracefulShutdown tests the server's graceful shutdown behavior
func TestGracefulShutdown(t *testing.T) {
	// Create a test server
	srv := &http.Server{
		Addr: ":0", // Let the OS choose a port
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond) // Simulate work
			w.WriteHeader(http.StatusOK)
		}),
	}

	// Start server in goroutine
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			t.Errorf("Expected ErrServerClosed, got %v", err)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Initiate shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown error: %v", err)
	}
}
