package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server struct {
		Port         int
		ReadTimeout  time.Duration
		WriteTimeout time.Duration
	}

	API struct {
		RequestTimeout time.Duration
		RetryAttempts  int
		RetryBackoff   time.Duration
	}

	Cache struct {
		TTL time.Duration
	}

	CircuitBreaker struct {
		MaxRequests  uint32
		Interval     time.Duration
		Timeout      time.Duration
		FailureRatio float64
	}

	RateLimit struct {
		RequestsPerSecond float64
		Burst             int
	}
}

// Load creates a new Config instance with values from environment variables,
// falling back to sensible defaults when env vars are not set
func Load() (*Config, error) {
	cfg := &Config{}

	// Server configuration
	cfg.Server.Port = getEnvInt("SERVER_PORT", 8080)
	cfg.Server.ReadTimeout = getEnvDuration("SERVER_READ_TIMEOUT", 15*time.Second)
	cfg.Server.WriteTimeout = getEnvDuration("SERVER_WRITE_TIMEOUT", 15*time.Second)

	// API configuration
	cfg.API.RequestTimeout = getEnvDuration("API_REQUEST_TIMEOUT", 5*time.Second)
	cfg.API.RetryAttempts = getEnvInt("API_RETRY_ATTEMPTS", 3)
	cfg.API.RetryBackoff = getEnvDuration("API_RETRY_BACKOFF", 100*time.Millisecond)

	// Cache configuration
	cfg.Cache.TTL = getEnvDuration("CACHE_TTL", 5*time.Minute)

	// Circuit breaker configuration
	cfg.CircuitBreaker.MaxRequests = uint32(getEnvInt("CB_MAX_REQUESTS", 3))
	cfg.CircuitBreaker.Interval = getEnvDuration("CB_INTERVAL", 10*time.Second)
	cfg.CircuitBreaker.Timeout = getEnvDuration("CB_TIMEOUT", 60*time.Second)
	cfg.CircuitBreaker.FailureRatio = getEnvFloat("CB_FAILURE_RATIO", 0.6)

	// Rate limit configuration
	cfg.RateLimit.RequestsPerSecond = getEnvFloat("RATE_LIMIT_RPS", 10.0)
	cfg.RateLimit.Burst = getEnvInt("RATE_LIMIT_BURST", 20)

	// Validate the configuration
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	// Server validation
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535")
	}
	if c.Server.ReadTimeout <= 0 {
		return fmt.Errorf("server read timeout must be positive")
	}
	if c.Server.WriteTimeout <= 0 {
		return fmt.Errorf("server write timeout must be positive")
	}

	// API validation
	if c.API.RequestTimeout <= 0 {
		return fmt.Errorf("API request timeout must be positive")
	}
	if c.API.RetryAttempts < 0 {
		return fmt.Errorf("API retry attempts must be non-negative")
	}
	if c.API.RetryBackoff <= 0 {
		return fmt.Errorf("API retry backoff must be positive")
	}

	// Cache validation
	if c.Cache.TTL <= 0 {
		return fmt.Errorf("cache TTL must be positive")
	}

	// Circuit breaker validation
	if c.CircuitBreaker.MaxRequests == 0 {
		return fmt.Errorf("circuit breaker max requests must be positive")
	}
	if c.CircuitBreaker.Interval <= 0 {
		return fmt.Errorf("circuit breaker interval must be positive")
	}
	if c.CircuitBreaker.Timeout <= 0 {
		return fmt.Errorf("circuit breaker timeout must be positive")
	}
	if c.CircuitBreaker.FailureRatio <= 0 || c.CircuitBreaker.FailureRatio > 1 {
		return fmt.Errorf("circuit breaker failure ratio must be between 0 and 1")
	}

	// Rate limit validation
	if c.RateLimit.RequestsPerSecond <= 0 {
		return fmt.Errorf("rate limit requests per second must be positive")
	}
	if c.RateLimit.Burst <= 0 {
		return fmt.Errorf("rate limit burst must be positive")
	}

	return nil
}

// Helper functions to get environment variables with type conversion
func getEnvStr(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	strValue := getEnvStr(key, "")
	if strValue == "" {
		return fallback
	}

	value, err := strconv.Atoi(strValue)
	if err != nil {
		return fallback
	}
	return value
}

func getEnvFloat(key string, fallback float64) float64 {
	strValue := getEnvStr(key, "")
	if strValue == "" {
		return fallback
	}

	value, err := strconv.ParseFloat(strValue, 64)
	if err != nil {
		return fallback
	}
	return value
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	strValue := getEnvStr(key, "")
	if strValue == "" {
		return fallback
	}

	value, err := time.ParseDuration(strValue)
	if err != nil {
		return fallback
	}
	return value
}
