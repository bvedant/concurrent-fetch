package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bvedant/concurrent-fetch/internal/cache"
	"github.com/bvedant/concurrent-fetch/internal/fetcher"
	"github.com/bvedant/concurrent-fetch/internal/processor"
)

func main() {
	// Create a cache with 5-minute TTL
	apiCache := cache.NewCache(5 * time.Minute)

	// Create HTTP server
	http.HandleFunc("/process", func(w http.ResponseWriter, r *http.Request) {
		// Create context with timeout
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Initialize fetchers with cache
		fetchers := []fetcher.DataFetcher{
			fetcher.NewAPIFetcher(
				"https://api.github.com/repos/golang/go",
				map[string]string{
					"Accept":     "application/json",
					"User-Agent": "concurrent-fetch-app",
				},
				apiCache,
			),
			fetcher.NewAPIFetcher(
				"https://api.github.com/repos/kubernetes/kubernetes",
				map[string]string{
					"Accept":     "application/json",
					"User-Agent": "concurrent-fetch-app",
				},
				apiCache,
			),
		}

		// Process data concurrently
		proc := processor.NewDataProcessor(fetchers...)
		results := proc.ProcessConcurrently(ctx)

		// Set response headers
		w.Header().Set("Content-Type", "application/json")

		// Send response
		response := make([]map[string]interface{}, 0)
		for _, result := range results {
			if result.Error != nil {
				apiErr, ok := result.Error.(*fetcher.APIError)
				if ok {
					response = append(response, map[string]interface{}{
						"error":       apiErr.Message,
						"status_code": apiErr.StatusCode,
						"url":         apiErr.URL,
					})
				} else {
					response = append(response, map[string]interface{}{
						"error": result.Error.Error(),
					})
				}
				continue
			}

			var data map[string]interface{}
			if err := json.Unmarshal(result.Data, &data); err != nil {
				response = append(response, map[string]interface{}{
					"error": "Failed to parse data",
				})
				continue
			}
			response = append(response, data)
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
	})

	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
