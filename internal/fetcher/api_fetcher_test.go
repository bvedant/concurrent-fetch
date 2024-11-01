package fetcher

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bvedant/concurrent-fetch/internal/cache"
)

func TestAPIFetcher_FetchData(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Test-Header") == "test-value" {
			w.Write([]byte("test response"))
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid request"))
		}
	}))
	defer server.Close()

	// Create cache with short TTL for testing
	testCache := cache.NewCache(100 * time.Millisecond)

	tests := []struct {
		name           string
		headers        map[string]string
		expectedData   string
		expectedError  bool
		expectedStatus int
	}{
		{
			name: "successful request",
			headers: map[string]string{
				"Test-Header": "test-value",
			},
			expectedData:   "test response",
			expectedError:  false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "failed request",
			headers:        map[string]string{},
			expectedError:  true,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher := NewAPIFetcher(server.URL, tt.headers, testCache)
			data, err := fetcher.FetchData(context.Background())

			if tt.expectedError {
				if err == nil {
					t.Error("expected error but got none")
				}
				if apiErr, ok := err.(*APIError); ok {
					if apiErr.StatusCode != tt.expectedStatus {
						t.Errorf("expected status code %d, got %d", tt.expectedStatus, apiErr.StatusCode)
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if string(data) != tt.expectedData {
					t.Errorf("expected %s, got %s", tt.expectedData, string(data))
				}

				// Test cache hit
				cachedData, err := fetcher.FetchData(context.Background())
				if err != nil {
					t.Errorf("cache hit error: %v", err)
				}
				if string(cachedData) != tt.expectedData {
					t.Errorf("cache hit: expected %s, got %s", tt.expectedData, string(cachedData))
				}
			}
		})
	}
}
