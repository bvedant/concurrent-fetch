package processor

import (
	"context"
	"sync"

	"github.com/bvedant/concurrent-fetch/internal/fetcher"
	"github.com/bvedant/concurrent-fetch/internal/models"
)

// DataProcessor handles concurrent data processing
type DataProcessor struct {
	fetchers []fetcher.DataFetcher
}

func NewDataProcessor(fetchers ...fetcher.DataFetcher) *DataProcessor {
	return &DataProcessor{fetchers: fetchers}
}

func (dp *DataProcessor) ProcessConcurrently(ctx context.Context) []models.Result {
	results := make([]models.Result, len(dp.fetchers))
	var wg sync.WaitGroup

	// Create buffered channel for results
	resultChan := make(chan models.Result, len(dp.fetchers))

	// Launch goroutine for each fetcher
	for i := range dp.fetchers {
		wg.Add(1)
		go func(index int, fetcher fetcher.DataFetcher) {
			defer wg.Done()
			data, err := fetcher.FetchData(ctx)
			resultChan <- models.Result{Data: data, Error: err}
		}(i, dp.fetchers[i])
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for i := 0; i < len(dp.fetchers); i++ {
		results[i] = <-resultChan
	}

	return results
}
