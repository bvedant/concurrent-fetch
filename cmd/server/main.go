package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bvedant/concurrent-fetch/internal/cache"
	"github.com/bvedant/concurrent-fetch/internal/fetcher"
	"github.com/bvedant/concurrent-fetch/internal/health"
	"github.com/bvedant/concurrent-fetch/internal/middleware"
	"github.com/bvedant/concurrent-fetch/internal/models"
	"github.com/bvedant/concurrent-fetch/internal/processor"
	"github.com/bvedant/concurrent-fetch/internal/validation"
)

func buildResponse(results []models.Result) []map[string]interface{} {
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
	return response
}

func processHandler(apiCache *cache.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := validation.ValidateRequest(r); err != nil {
			http.Error(w, err.Error(), http.StatusMethodNotAllowed)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		headers := map[string]string{
			"Accept":     "application/json",
			"User-Agent": "concurrent-fetch-app",
		}

		fetchers := []fetcher.DataFetcher{
			fetcher.NewAPIFetcher(
				"https://api.agify.io/?name=bella",
				headers,
				apiCache,
			),
			fetcher.NewAPIFetcher(
				"https://api.nationalize.io/?name=nathan",
				headers,
				apiCache,
			),
		}

		proc := processor.NewDataProcessor(fetchers...)
		results := proc.ProcessConcurrently(ctx)

		w.Header().Set("Content-Type", "application/json")
		response := buildResponse(results)

		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "    ")
		if err := encoder.Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

func main() {
	log.Printf("Initializing server components...")

	apiCache := cache.NewCache(5 * time.Minute)
	log.Printf("Cache initialized with 5-minute TTL")

	healthChecker := health.NewHealthChecker()

	healthChecker.AddCheck("cache", func() error {
		testKey := "health-check"
		testData := []byte("test")

		apiCache.Set(testKey, testData)
		data, found := apiCache.Get(testKey)

		if !found {
			return fmt.Errorf("cache health check failed: data not found")
		}
		if string(data) != string(testData) {
			return fmt.Errorf("cache health check failed: data mismatch")
		}
		return nil
	})

	rateLimiter := middleware.NewRateLimiter(10, 10)
	log.Printf("Rate limiter initialized: 10 requests/second per IP")

	http.Handle("/health", healthChecker.Handler())
	http.Handle("/process", rateLimiter.Middleware(processHandler(apiCache)))

	log.Printf("Configuring HTTP server...")

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      http.DefaultServeMux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigChan
		log.Printf("Received shutdown signal: %v", sig)

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		log.Printf("Initiating graceful shutdown...")
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		} else {
			log.Printf("Server shutdown completed successfully")
		}
	}()

	log.Printf("Server starting on :8080")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %v", err)
	}

	log.Printf("Server stopped")
}
