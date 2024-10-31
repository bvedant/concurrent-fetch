package utils

import (
	"context"
	"time"
)

// RetryConfig holds configuration for retry behavior
type RetryConfig struct {
	MaxAttempts       int
	InitialBackoff    time.Duration
	MaxBackoff        time.Duration
	BackoffMultiplier float64
}

// DefaultRetryConfig provides sensible default values
var DefaultRetryConfig = RetryConfig{
	MaxAttempts:       3,
	InitialBackoff:    100 * time.Millisecond,
	MaxBackoff:        2 * time.Second,
	BackoffMultiplier: 2.0,
}

// RetryWithBackoff executes the given operation with exponential backoff
func RetryWithBackoff(ctx context.Context, config RetryConfig, operation func() error) error {
	var err error
	currentBackoff := config.InitialBackoff

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		err = operation()
		if err == nil {
			return nil
		}

		// If this was our last attempt, return the error
		if attempt == config.MaxAttempts {
			return err
		}

		// Check if context is cancelled before sleeping
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(currentBackoff):
		}

		// Increase backoff for next attempt
		currentBackoff = time.Duration(float64(currentBackoff) * config.BackoffMultiplier)
		if currentBackoff > config.MaxBackoff {
			currentBackoff = config.MaxBackoff
		}
	}

	return err
}
