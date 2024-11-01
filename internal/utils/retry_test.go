package utils

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetryWithBackoff(t *testing.T) {
	tests := []struct {
		name          string
		maxAttempts   int
		shouldSucceed bool
		operation     func() error
	}{
		{
			name:          "success on first attempt",
			maxAttempts:   3,
			shouldSucceed: true,
			operation: func() error {
				return nil
			},
		},
		{
			name:          "success after retries",
			maxAttempts:   3,
			shouldSucceed: true,
			operation: func() func() error {
				attempts := 0
				return func() error {
					attempts++
					if attempts < 2 {
						return errors.New("temporary error")
					}
					return nil
				}
			}(),
		},
		{
			name:          "failure after all attempts",
			maxAttempts:   2,
			shouldSucceed: false,
			operation: func() error {
				return errors.New("persistent error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			config := RetryConfig{
				MaxAttempts:       tt.maxAttempts,
				InitialBackoff:    10 * time.Millisecond,
				MaxBackoff:        100 * time.Millisecond,
				BackoffMultiplier: 2.0,
			}

			err := RetryWithBackoff(ctx, config, tt.operation)

			if tt.shouldSucceed && err != nil {
				t.Errorf("Expected success, got error: %v", err)
			}
			if !tt.shouldSucceed && err == nil {
				t.Error("Expected error, got success")
			}
		})
	}
}
