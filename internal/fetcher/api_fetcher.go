package fetcher

import (
	"context"
	"encoding/json"
	"time"
)

// APIFetcher implements DataFetcher for REST APIs
type APIFetcher struct {
	URL string
}

func NewAPIFetcher(url string) *APIFetcher {
	return &APIFetcher{URL: url}
}

func (a *APIFetcher) FetchData(ctx context.Context) ([]byte, error) {
	// Simulate API call with timeout
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(1 * time.Second):
		// Simulated data
		return json.Marshal(map[string]string{"source": "API", "data": "sample"})
	}
}
