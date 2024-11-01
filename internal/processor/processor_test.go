package processor

import (
	"context"
	"errors"
	"testing"

	"github.com/bvedant/concurrent-fetch/internal/fetcher"
)

// MockFetcher implements fetcher.DataFetcher for testing
type MockFetcher struct {
	data []byte
	err  error
}

func (m *MockFetcher) FetchData(ctx context.Context) ([]byte, error) {
	return m.data, m.err
}

func TestProcessConcurrently(t *testing.T) {
	tests := []struct {
		name     string
		fetchers []MockFetcher
	}{
		{
			name: "successful fetches",
			fetchers: []MockFetcher{
				{data: []byte("data1"), err: nil},
				{data: []byte("data2"), err: nil},
			},
		},
		{
			name: "mixed success and failure",
			fetchers: []MockFetcher{
				{data: []byte("data1"), err: nil},
				{data: nil, err: errors.New("fetch error")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert MockFetchers to fetcher.DataFetcher interface
			fetchers := make([]fetcher.DataFetcher, len(tt.fetchers))
			for i := range tt.fetchers {
				fetchers[i] = &tt.fetchers[i]
			}

			processor := NewDataProcessor(fetchers...)
			results := processor.ProcessConcurrently(context.Background())

			if len(results) != len(tt.fetchers) {
				t.Errorf("Expected %d results, got %d", len(tt.fetchers), len(results))
			}

			for i, result := range results {
				if tt.fetchers[i].err != nil {
					if result.Error == nil {
						t.Errorf("Expected error for fetcher %d", i)
					}
				} else {
					if result.Error != nil {
						t.Errorf("Unexpected error for fetcher %d: %v", i, result.Error)
					}
					if string(result.Data) != string(tt.fetchers[i].data) {
						t.Errorf("Expected data %s, got %s", tt.fetchers[i].data, result.Data)
					}
				}
			}
		})
	}
}
