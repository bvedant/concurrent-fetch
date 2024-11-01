package api

import (
	"context"
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
		// Implement your fetch logic here
		w.WriteHeader(http.StatusOK)
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
