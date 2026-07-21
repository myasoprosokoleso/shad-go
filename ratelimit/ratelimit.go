//go:build !solution

package ratelimit

import (
	"context"
	"errors"
	"time"
)

// Limiter is precise rate limiter with context support.
type Limiter struct {
	tokens   chan struct{}
	stop     chan struct{}
	interval time.Duration
}

var ErrStopped = errors.New("limiter stopped")

// NewLimiter returns limiter that throttles rate of successful Acquire() calls
// to maxSize events at any given interval.
func NewLimiter(maxCount int, interval time.Duration) *Limiter {
	l := &Limiter{
		interval: interval,
		stop:     make(chan struct{}),
	}

	if interval > 0 {
		l.tokens = make(chan struct{}, maxCount)
		for range maxCount {
			l.tokens <- struct{}{}
		}
	}

	return l
}

func (l *Limiter) Acquire(ctx context.Context) error {
	select {
	case <-l.stop:
		return ErrStopped
	default:
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		if l.tokens != nil {
			select {
			case <-l.stop:
				return ErrStopped
			case <-ctx.Done():
				return ctx.Err()
			case <-l.tokens:
				go l.replenish()
			}
		}
		return nil
	}
}

func (l *Limiter) replenish() {
	timer := time.NewTimer(l.interval)
	defer timer.Stop()
	select {
	case <-timer.C:
		l.tokens <- struct{}{}
	case <-l.stop:
	}
}

func (l *Limiter) Stop() {
	close(l.stop)
}
