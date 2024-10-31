package fetcher

import (
	"context"
)

// DataFetcher interface defines contract for data sources
type DataFetcher interface {
	FetchData(ctx context.Context) ([]byte, error)
}
