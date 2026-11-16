package api

import (
	"sync"
	"time"
)

// honeypotField is the JSON key bots are expected to leave empty. A non-empty
// value indicates an automated submission and the request is rejected.
const honeypotField = "website"

// antiBot tracks per-key submission counts within a sliding window to slow
// down credential-stuffing / registration spam. It is safe for concurrent use.
type antiBot struct {
	mu       sync.Mutex
	window   time.Duration
	limit    int
	counts   map[string][]time.Time
}

func newAntiBot(window time.Duration, limit int) *antiBot {
	if window <= 0 {
		window = time.Minute
	}
	if limit <= 0 {
		limit = 20
	}
	return &antiBot{window: window, limit: limit, counts: make(map[string][]time.Time)}
}

// allow reports whether key may proceed and records the attempt when allowed.
func (a *antiBot) allow(key string) bool {
	if a == nil {
		return true
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-a.window)
	var kept []time.Time
	for _, t := range a.counts[key] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= a.limit {
		a.counts[key] = kept
		return false
	}
	kept = append(kept, now)
	a.counts[key] = kept
	return true
}
