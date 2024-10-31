package fetcher

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bvedant/concurrent-fetch/internal/cache"
	"github.com/bvedant/concurrent-fetch/internal/logger"
	"github.com/bvedant/concurrent-fetch/internal/metrics"
	"github.com/bvedant/concurrent-fetch/internal/utils"
	"github.com/sony/gobreaker"
	"go.uber.org/zap"
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
	breaker *gobreaker.CircuitBreaker
}

func NewAPIFetcher(url string, headers map[string]string, cache *cache.Cache) *APIFetcher {
	return &APIFetcher{
		URL:     url,
		Headers: headers,
		Retry:   utils.DefaultRetryConfig,
		cache:   cache,
		breaker: gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name: url,
		}),
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
	if data, found := a.cache.Get(a.generateCacheKey()); found {
		metrics.CacheHits.WithLabelValues("hit").Inc()
		return data, nil
	}
	metrics.CacheHits.WithLabelValues("miss").Inc()

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

		result, err := a.breaker.Execute(func() (interface{}, error) {
			// Track request duration
			start := time.Now()
			resp, err := http.DefaultClient.Do(req)
			metrics.RequestDuration.WithLabelValues(a.URL).Observe(time.Since(start).Seconds())

			if err != nil {
				logger.Log.Error("API request failed",
					zap.String("url", a.URL),
					zap.Error(err),
				)
				return nil, err
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}

			// Check for non-200 status codes
			if resp.StatusCode != http.StatusOK {
				return nil, &APIError{
					StatusCode: resp.StatusCode,
					Message:    string(body),
					URL:        a.URL,
				}
			}

			// Cache successful response
			a.cache.Set(a.generateCacheKey(), body)

			return body, nil
		})

		if err != nil {
			return err
		}

		responseData = result.([]byte)
		return nil
	}

	err := utils.RetryWithBackoff(ctx, a.Retry, operation)
	if err != nil {
		return nil, err
	}

	return responseData, nil
}
