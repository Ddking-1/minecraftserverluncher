package api

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// rateLimiter is a minimal per-client-IP token bucket used to protect
// registration, login and refresh endpoints (docs/AUTH.md "Rate limiting").
type rateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	buckets map[string]*bucket
}

type bucket struct {
	count    int
	resetAt  time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		limit:   limit,
		window:  window,
		buckets: make(map[string]*bucket),
	}
}

// allow reports whether a request from the given key is within the limit.
// Expired buckets are pruned so the map does not grow without bound.
func (l *rateLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	l.pruneLocked(now)
	if b, ok := l.buckets[key]; ok {
		if b.count >= l.limit {
			return false
		}
		b.count++
		return true
	}
	l.buckets[key] = &bucket{count: 1, resetAt: now.Add(l.window)}
	return true
}

// pruneLocked removes expired buckets. Callers must hold l.mu.
func (l *rateLimiter) pruneLocked(now time.Time) {
	for k, b := range l.buckets {
		if now.After(b.resetAt) {
			delete(l.buckets, k)
		}
	}
}

// limit wraps a handler with per-IP rate limiting.
func (s *Server) limit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if !s.limiter.allow(ip) {
			writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}
		next(w, r)
	}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
