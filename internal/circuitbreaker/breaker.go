package circuitbreaker

import (
	"time"

	"github.com/bvedant/concurrent-fetch/internal/logger"
	"github.com/sony/gobreaker"
	"go.uber.org/zap"
)

func NewCircuitBreaker(name string) *gobreaker.CircuitBreaker {
	return gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        name,
		MaxRequests: 3,
		Interval:    10 * time.Second,
		Timeout:     60 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			shouldTrip := counts.Requests >= 3 && failureRatio >= 0.6

			if shouldTrip {
				logger.Log.Warn("Circuit breaker tripping",
					zap.String("name", name),
					zap.Float64("failure_ratio", failureRatio),
					zap.Int64("total_requests", int64(counts.Requests)),
				)
			}
			return shouldTrip
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			logger.Log.Info("Circuit breaker state changed",
				zap.String("name", name),
				zap.String("from", from.String()),
				zap.String("to", to.String()),
			)
		},
	})
}
