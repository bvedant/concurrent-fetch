package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bvedant/concurrent-fetch/internal/config"
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
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		urls := r.URL.Query()["url"]
		if len(urls) == 0 {
			http.Error(w, "No URLs provided", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		results := make([]map[string]interface{}, 0, len(urls))
		errors := make([]error, 0)

		// Create a channel for results
		type result struct {
			url  string
			data []byte
			err  error
		}
		resultChan := make(chan result, len(urls))

		// Fetch concurrently
		for _, url := range urls {
			go func(url string) {
				client := NewClient(url, a.config)
				data, err := client.Get(ctx, "")
				resultChan <- result{url: url, data: data, err: err}
			}(url)
		}

		// Collect results
		for i := 0; i < len(urls); i++ {
			res := <-resultChan
			if res.err != nil {
				errors = append(errors, fmt.Errorf("failed to fetch %s: %w", res.url, res.err))
				continue
			}

			var data map[string]interface{}
			if err := json.Unmarshal(res.data, &data); err != nil {
				errors = append(errors, fmt.Errorf("failed to parse %s: %w", res.url, err))
				continue
			}
			results = append(results, data)
		}

		response := map[string]interface{}{
			"results": results,
		}
		if len(errors) > 0 {
			errMsgs := make([]string, len(errors))
			for i, err := range errors {
				errMsgs[i] = err.Error()
			}
			response["errors"] = errMsgs
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
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
