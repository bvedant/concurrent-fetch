package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bvedant/concurrent-fetch/internal/fetcher"
	"github.com/bvedant/concurrent-fetch/internal/processor"
)

func main() {
	// Create HTTP server
	http.HandleFunc("/process", func(w http.ResponseWriter, r *http.Request) {
		// Create context with timeout
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Initialize fetchers
		fetchers := []fetcher.DataFetcher{
			fetcher.NewAPIFetcher(
				"https://api.github.com/users/microsoft",
				map[string]string{"Accept": "application/json"},
			),
			fetcher.NewAPIFetcher(
				"https://api.github.com/users/google",
				map[string]string{"Accept": "application/json"},
			),
		}

		// Process data concurrently
		proc := processor.NewDataProcessor(fetchers...)
		results := proc.ProcessConcurrently(ctx)

		// Send response
		response := make([]map[string]interface{}, 0)
		for _, result := range results {
			if result.Error != nil {
				response = append(response, map[string]interface{}{
					"error": result.Error.Error(),
				})
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

		json.NewEncoder(w).Encode(response)
	})

	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
