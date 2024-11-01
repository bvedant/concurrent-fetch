package circuit

import (
	"errors"
	"sync"
	"time"
)

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

type Breaker struct {
	state     State
	failures  int
	threshold int
	timeout   time.Duration
	lastErr   error
	mu        sync.RWMutex
}

func NewBreaker(threshold int, timeout time.Duration) *Breaker {
	return &Breaker{
		state:     StateClosed,
		threshold: threshold,
		timeout:   timeout,
	}
}

func (b *Breaker) Execute(fn func() error) error {
	b.mu.RLock()
	if b.state == StateOpen {
		b.mu.RUnlock()
		return errors.New("circuit breaker is open")
	}
	b.mu.RUnlock()

	err := fn()
	if err != nil {
		b.recordFailure(err)
		return err
	}

	b.reset()
	return nil
}

func (b *Breaker) recordFailure(err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.failures++
	b.lastErr = err

	if b.failures >= b.threshold {
		b.state = StateOpen
		go b.attemptReset()
	}
}

func (b *Breaker) attemptReset() {
	time.Sleep(b.timeout)
	b.mu.Lock()
	defer b.mu.Unlock()
	b.state = StateHalfOpen
}

func (b *Breaker) reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures = 0
	b.state = StateClosed
}
