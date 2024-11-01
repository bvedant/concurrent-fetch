package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path"

	"github.com/bvedant/concurrent-fetch/internal/config"
	"github.com/bvedant/concurrent-fetch/internal/platform/cache"
	"github.com/bvedant/concurrent-fetch/internal/platform/circuit"
	"github.com/bvedant/concurrent-fetch/pkg/retry"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	cache      *cache.Cache
	breaker    *circuit.Breaker
	config     *config.Config
}

type Option func(*Client)

func NewClient(baseURL string, cfg *config.Config, opts ...Option) *Client {
	client := &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: cfg.API.RequestTimeout,
		},
		config:  cfg,
		cache:   cache.New(cfg.Cache.TTL),
		breaker: circuit.NewBreaker(int(cfg.CircuitBreaker.MaxRequests), cfg.CircuitBreaker.Timeout),
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func (c *Client) Get(ctx context.Context, endpoint string) ([]byte, error) {
	cacheKey := path.Join(c.baseURL, endpoint)

	// Check cache first
	if data, found := c.cache.Get(cacheKey); found {
		return data, nil
	}

	var responseData []byte
	operation := func() error {
		return c.breaker.Execute(func() error {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet,
				path.Join(c.baseURL, endpoint), nil)
			if err != nil {
				return fmt.Errorf("creating request: %w", err)
			}

			resp, err := c.httpClient.Do(req)
			if err != nil {
				return fmt.Errorf("executing request: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			}

			data, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("reading response: %w", err)
			}

			responseData = data
			c.cache.Set(cacheKey, data)
			return nil
		})
	}

	retryConfig := retry.Config{
		MaxAttempts:       c.config.API.RetryAttempts,
		InitialBackoff:    c.config.API.RetryBackoff,
		MaxBackoff:        c.config.API.RetryBackoff * 10,
		BackoffMultiplier: 2.0,
	}

	if err := retry.WithBackoff(ctx, retryConfig, operation); err != nil {
		return nil, fmt.Errorf("failed after %d attempts: %w",
			c.config.API.RetryAttempts, err)
	}

	return responseData, nil
}
