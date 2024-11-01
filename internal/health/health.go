package health

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type HealthChecker struct {
	checks map[string]CheckFunc
	mu     sync.RWMutex
}

type CheckFunc func() error

type HealthStatus struct {
	Status    string            `json:"status"`
	Checks    map[string]string `json:"checks"`
	Timestamp time.Time         `json:"timestamp"`
}

func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checks: make(map[string]CheckFunc),
	}
}

func (h *HealthChecker) AddCheck(name string, check CheckFunc) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checks[name] = check
}

func (h *HealthChecker) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := HealthStatus{
			Status:    "ok",
			Checks:    make(map[string]string),
			Timestamp: time.Now(),
		}

		h.mu.RLock()
		defer h.mu.RUnlock()

		for name, check := range h.checks {
			if err := check(); err != nil {
				status.Status = "error"
				status.Checks[name] = err.Error()
			} else {
				status.Checks[name] = "ok"
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if status.Status != "ok" {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		json.NewEncoder(w).Encode(status)
	}
}
