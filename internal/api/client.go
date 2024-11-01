package api

import (
	"context"
	"net/http"

	"github.com/bvedant/concurrent-fetch/internal/config"
	"github.com/bvedant/concurrent-fetch/internal/platform/cache"
	"github.com/bvedant/concurrent-fetch/internal/platform/circuit"
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
		config: cfg,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func (c *Client) Get(ctx context.Context, path string) ([]byte, error) {
	// TODO: Implement GET request with circuit breaker, caching, and retry logic
	return nil, nil
}
