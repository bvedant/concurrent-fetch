package metrics

import (
	"testing"
)

func TestMetrics(t *testing.T) {
	// Test histogram
	t.Run("request duration", func(t *testing.T) {
		metric := RequestDuration.WithLabelValues("test_endpoint")
		metric.Observe(1.5)
	})

	// Test counter
	t.Run("cache hits", func(t *testing.T) {
		metric := CacheHits.WithLabelValues("test_cache")
		metric.Inc()
	})

	// Test gauge
	t.Run("circuit breaker state", func(t *testing.T) {
		metric := CircuitBreakerState.WithLabelValues("test_breaker")
		metric.Set(1)
	})
}
