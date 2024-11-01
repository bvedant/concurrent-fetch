package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bvedant/concurrent-fetch/internal/config"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"
)

type App struct {
	config *config.Config
	server *http.Server
}

func NewApp(cfg *config.Config) *App {
	app := &App{
		config: cfg,
	}

	// Create router and add handlers
	mux := http.NewServeMux()

	// Add rate limiting middleware
	limiter := rate.NewLimiter(rate.Limit(cfg.RateLimit.RequestsPerSecond),
		cfg.RateLimit.Burst)

	// Add your handlers here
	mux.Handle("/fetch", rateLimitMiddleware(limiter, app.handleFetch()))
	mux.Handle("/health", app.handleHealth())

	app.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	return app
}

func (a *App) Start() error {
	// Start server in a goroutine
	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Error starting server: %v\n", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return a.server.Shutdown(ctx)
}

func (a *App) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

func (a *App) handleFetch() http.HandlerFunc {
	const maxURLs = 10

	return func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New().String()
		start := time.Now()

		// Method validation
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// URL validation
		urls := r.URL.Query()["url"]
		if len(urls) == 0 {
			http.Error(w, "No URLs provided", http.StatusBadRequest)
			return
		}
		if len(urls) > maxURLs {
			http.Error(w, fmt.Sprintf("Too many URLs. Maximum allowed: %d", maxURLs),
				http.StatusBadRequest)
			return
		}

		for _, urlStr := range urls {
			if _, err := url.Parse(urlStr); err != nil {
				http.Error(w, fmt.Sprintf("Invalid URL provided: %s", urlStr),
					http.StatusBadRequest)
				return
			}
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(r.Context(), a.config.API.RequestTimeout)
		defer cancel()

		type fetchResult struct {
			URL     string        `json:"url"`
			Data    interface{}   `json:"data,omitempty"`
			Error   string        `json:"error,omitempty"`
			Status  int           `json:"status"`
			Latency time.Duration `json:"latency"`
		}

		// Create buffered channel for results
		resultsChan := make(chan fetchResult, len(urls))
		defer close(resultsChan)

		// Create error group for goroutine management
		g, ctx := errgroup.WithContext(ctx)

		// Launch goroutine for each URL
		for _, urlStr := range urls {
			urlStr := urlStr // Create new variable for goroutine
			g.Go(func() error {
				fetchStart := time.Now()
				client := NewClient(urlStr, a.config)
				data, err := client.Get(ctx, "")

				result := fetchResult{
					URL:     urlStr,
					Status:  http.StatusOK,
					Latency: time.Since(fetchStart),
				}

				if err != nil {
					switch {
					case errors.Is(err, context.DeadlineExceeded):
						result.Status = http.StatusGatewayTimeout
						result.Error = "request timed out"
					case errors.Is(err, context.Canceled):
						result.Status = http.StatusRequestTimeout
						result.Error = "request cancelled"
					default:
						result.Status = http.StatusInternalServerError
						result.Error = err.Error()
					}
				} else {
					var jsonData interface{}
					if err := json.Unmarshal(data, &jsonData); err != nil {
						result.Status = http.StatusUnprocessableEntity
						result.Error = "invalid JSON response"
						result.Data = string(data) // Include raw data for debugging
					} else {
						result.Data = jsonData
					}
				}

				select {
				case resultsChan <- result:
				case <-ctx.Done():
					return ctx.Err()
				}
				return nil
			})
		}

		// Wait for all goroutines
		if err := g.Wait(); err != nil {
			log.Printf("Error in fetch operations: %v", err)
		}

		// Collect results
		results := make([]fetchResult, 0, len(urls))
		errors := false
		for i := 0; i < len(urls); i++ {
			result := <-resultsChan
			results = append(results, result)
			if result.Error != "" {
				errors = true
			}
		}

		// Prepare response
		response := struct {
			RequestID string        `json:"request_id"`
			Success   bool          `json:"success"`
			Results   []fetchResult `json:"results"`
			Duration  string        `json:"duration"`
		}{
			RequestID: requestID,
			Success:   !errors,
			Results:   results,
			Duration:  time.Since(start).String(),
		}

		// Set response headers
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("X-Request-ID", requestID)

		if errors {
			w.WriteHeader(http.StatusMultiStatus)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		// Write response
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			return
		}
	}
}

func rateLimitMiddleware(limiter *rate.Limiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}
