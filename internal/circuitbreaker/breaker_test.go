package circuitbreaker

import (
	"errors"
	"testing"
	"time"

	"github.com/bvedant/concurrent-fetch/internal/logger"
	"github.com/sony/gobreaker"
	"go.uber.org/zap"
)

func TestNewCircuitBreaker(t *testing.T) {
	// Setup logger for testing
	testLogger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create test logger: %v", err)
	}
	logger.Log = testLogger
	defer logger.Log.Sync()

	breaker := NewCircuitBreaker("test-breaker")

	// Test initial state
	if breaker.State() != gobreaker.StateClosed {
		t.Errorf("Expected initial state to be Closed, got %v", breaker.State())
	}

	// Test successful execution
	result, err := breaker.Execute(func() (interface{}, error) {
		return "success", nil
	})

	if err != nil {
		t.Errorf("Expected successful execution, got error: %v", err)
	}
	if result != "success" {
		t.Errorf("Expected result 'success', got %v", result)
	}

	// Test circuit breaker tripping
	testErr := errors.New("test error")

	// Force multiple failures
	for i := 0; i < 5; i++ {
		_, _ = breaker.Execute(func() (interface{}, error) {
			return nil, testErr
		})
	}

	// Verify circuit breaker is open
	if breaker.State() != gobreaker.StateOpen {
		t.Errorf("Expected state to be Open after failures, got %v", breaker.State())
	}

	// Test that requests fail fast when circuit is open
	_, err = breaker.Execute(func() (interface{}, error) {
		return "success", nil
	})
	if err == nil {
		t.Error("Expected error when circuit is open")
	}

	// Test recovery (without long sleeps for test efficiency)
	time.Sleep(100 * time.Millisecond) // Short sleep for test purposes

	// Reset the breaker for testing recovery
	breaker = NewCircuitBreaker("test-breaker-recovery")

	result, err = breaker.Execute(func() (interface{}, error) {
		return "recovered", nil
	})

	if err != nil {
		t.Errorf("Expected successful recovery, got error: %v", err)
	}
	if result != "recovered" {
		t.Errorf("Expected result 'recovered', got %v", result)
	}
}
