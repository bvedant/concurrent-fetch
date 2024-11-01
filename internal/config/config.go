package config

import (
	"time"
)

type Config struct {
	Server struct {
		Port         int           `env:"SERVER_PORT" envDefault:"8080"`
		ReadTimeout  time.Duration `env:"SERVER_READ_TIMEOUT" envDefault:"15s"`
		WriteTimeout time.Duration `env:"SERVER_WRITE_TIMEOUT" envDefault:"15s"`
	}

	API struct {
		RequestTimeout time.Duration `env:"API_REQUEST_TIMEOUT" envDefault:"5s"`
		RetryAttempts  int           `env:"API_RETRY_ATTEMPTS" envDefault:"3"`
		RetryBackoff   time.Duration `env:"API_RETRY_BACKOFF" envDefault:"100ms"`
	}

	Cache struct {
		TTL time.Duration `env:"CACHE_TTL" envDefault:"5m"`
	}

	CircuitBreaker struct {
		MaxRequests  uint32        `env:"CB_MAX_REQUESTS" envDefault:"3"`
		Interval     time.Duration `env:"CB_INTERVAL" envDefault:"10s"`
		Timeout      time.Duration `env:"CB_TIMEOUT" envDefault:"60s"`
		FailureRatio float64       `env:"CB_FAILURE_RATIO" envDefault:"0.6"`
	}

	RateLimit struct {
		RequestsPerSecond float64 `env:"RATE_LIMIT_RPS" envDefault:"10"`
		Burst             int     `env:"RATE_LIMIT_BURST" envDefault:"20"`
	}
}

func Load() (*Config, error) {
	// TODO: Implement configuration loading from environment variables
	// For now, return default config
	return &Config{}, nil
}
