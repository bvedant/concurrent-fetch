package fetcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// APIFetcher implements DataFetcher for REST APIs
type APIFetcher struct {
	URL     string
	Headers map[string]string
}

func NewAPIFetcher(url string, headers map[string]string) *APIFetcher {
	return &APIFetcher{
		URL:     url,
		Headers: headers,
	}
}

func (a *APIFetcher) FetchData(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", a.URL, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range a.Headers {
		req.Header.Add(key, value)
	}

	req.Header.Set("User-Agent", "concurrent-fetch-app")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
