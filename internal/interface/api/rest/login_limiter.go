package rest

import (
	"sync"
	"time"
)

type bucket struct {
	count int
	start time.Time
}

type LoginLimiter struct {
	mu      sync.Mutex
	max     int
	entries map[string]bucket
}

func NewLoginLimiter(max int) *LoginLimiter {
	return &LoginLimiter{max: max, entries: make(map[string]bucket)}
}

func (l *LoginLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	b := l.entries[key]
	now := time.Now()
	if b.start.IsZero() || now.Sub(b.start) >= time.Minute {
		l.entries[key] = bucket{count: 1, start: now}
		return true
	}
	if b.count >= l.max {
		return false
	}
	b.count++
	l.entries[key] = b
	return true
}
