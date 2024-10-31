package fetcher

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	"github.com/bvedant/concurrent-fetch/internal/cache"
	"github.com/bvedant/concurrent-fetch/internal/utils"
)

// APIError represents a detailed API error
type APIError struct {
	StatusCode int
	Message    string
	URL        string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error: %s (status: %d, url: %s)",
		e.Message, e.StatusCode, e.URL)
}

// APIFetcher implements DataFetcher for REST APIs
type APIFetcher struct {
	URL     string
	Headers map[string]string
	Retry   utils.RetryConfig
	cache   *cache.Cache
}

func NewAPIFetcher(url string, headers map[string]string, cache *cache.Cache) *APIFetcher {
	return &APIFetcher{
		URL:     url,
		Headers: headers,
		Retry:   utils.DefaultRetryConfig,
		cache:   cache,
	}
}

// generateCacheKey creates a unique key based on URL and headers
func (a *APIFetcher) generateCacheKey() string {
	// Create a hash of URL and headers for cache key
	h := sha256.New()
	h.Write([]byte(a.URL))
	for k, v := range a.Headers {
		h.Write([]byte(k + v))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (a *APIFetcher) FetchData(ctx context.Context) ([]byte, error) {
	// Check cache first
	if a.cache != nil {
		if data, found := a.cache.Get(a.generateCacheKey()); found {
			return data, nil
		}
	}

	var responseData []byte

	operation := func() error {
		req, err := http.NewRequestWithContext(ctx, "GET", a.URL, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		for key, value := range a.Headers {
			req.Header.Add(key, value)
		}
		req.Header.Set("User-Agent", "concurrent-fetch-app")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return &APIError{
				StatusCode: resp.StatusCode,
				Message:    fmt.Sprintf("unexpected status code"),
				URL:        a.URL,
			}
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		responseData = data
		return nil
	}

	err := utils.RetryWithBackoff(ctx, a.Retry, operation)
	if err != nil {
		return nil, err
	}

	// Store in cache if successful
	if a.cache != nil {
		a.cache.Set(a.generateCacheKey(), responseData)
	}

	return responseData, nil
}
