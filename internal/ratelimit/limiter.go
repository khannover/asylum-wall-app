package ratelimit

import (
	"sync"
	"time"
)

type Limiter struct {
	mu     sync.Mutex
	events map[string][]time.Time
}

func New() *Limiter {
	return &Limiter{events: make(map[string][]time.Time)}
}

func (l *Limiter) Allow(key string, max int, window time.Duration) bool {
	if max <= 0 {
		return true
	}

	now := time.Now()
	cutoff := now.Add(-window)

	l.mu.Lock()
	defer l.mu.Unlock()

	times := l.events[key]
	kept := make([]time.Time, 0, len(times))
	for _, t := range times {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= max {
		l.events[key] = kept
		return false
	}

	kept = append(kept, now)
	l.events[key] = kept
	return true
}