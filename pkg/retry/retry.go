package retry

import (
	"context"
	"time"
)

type Config struct {
	MaxAttempts       int
	InitialBackoff    time.Duration
	MaxBackoff        time.Duration
	BackoffMultiplier float64
}

func WithBackoff(ctx context.Context, config Config, operation func() error) error {
	var err error
	currentBackoff := config.InitialBackoff

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		if err = operation(); err == nil {
			return nil
		}

		if attempt == config.MaxAttempts {
			break
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(currentBackoff):
		}

		currentBackoff = time.Duration(float64(currentBackoff) * config.BackoffMultiplier)
		if currentBackoff > config.MaxBackoff {
			currentBackoff = config.MaxBackoff
		}
	}

	return err
}
