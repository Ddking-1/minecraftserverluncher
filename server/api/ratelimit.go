package api

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Config configures a RateLimiter.
type Config struct {
	// RequestsPerMinute is the sustained per-key request budget.
	RequestsPerMinute int

	// CleanupInterval is how often idle keys are swept from memory.
	CleanupInterval time.Duration

	// CleanupTolerance is how long a key may stay idle before it is removed.
	CleanupTolerance time.Duration

	// TrustProxy makes the limiter read the client address from
	// X-Forwarded-For / X-Real-IP. Only enable this when the backend is
	// reachable exclusively through a reverse proxy that overwrites those
	// headers; otherwise a client can spoof them and bypass rate limiting.
	TrustProxy bool
}

const defaultRequestsPerMinute = 30

// defaultConfig returns sensible, non-secret defaults.
func defaultConfig() Config {
	return Config{
		RequestsPerMinute: defaultRequestsPerMinute,
		CleanupInterval:   3 * time.Minute,
		CleanupTolerance:  5 * time.Minute,
	}
}

// visitor holds the token bucket and last-seen time for one key.
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter is a concurrency-safe per-client token-bucket limiter.
type RateLimiter struct {
	visitors    map[string]*visitor
	mu          sync.RWMutex
	cfg         Config
	cleanupDone chan struct{}
	stopOnce    sync.Once
}

// NewRateLimiter creates a RateLimiter with the default configuration or the
// supplied one. A non-positive RequestsPerMinute falls back to the default.
func NewRateLimiter(cfg ...Config) *RateLimiter {
	c := defaultConfig()
	if len(cfg) > 0 {
		c = cfg[0]
	}
	if c.RequestsPerMinute <= 0 {
		c.RequestsPerMinute = defaultRequestsPerMinute
	}
	if c.CleanupInterval <= 0 {
		c.CleanupInterval = 3 * time.Minute
	}
	if c.CleanupTolerance <= 0 {
		c.CleanupTolerance = 5 * time.Minute
	}

	rl := &RateLimiter{
		visitors:    make(map[string]*visitor),
		cfg:         c,
		cleanupDone: make(chan struct{}),
	}

	go rl.cleanupLoop()

	return rl
}

// getVisitor returns the visitor for key, creating it when missing.
func (rl *RateLimiter) getVisitor(key string) *visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, ok := rl.visitors[key]
	if !ok {
		// Convert the per-minute budget to a per-second refill rate. The burst
		// equals the per-minute budget so short legitimate bursts are allowed.
		limit := rate.Limit(float64(rl.cfg.RequestsPerMinute) / 60.0)
		v = &visitor{
			limiter:  rate.NewLimiter(limit, rl.cfg.RequestsPerMinute),
			lastSeen: time.Now(),
		}
		rl.visitors[key] = v
	}
	v.lastSeen = time.Now()
	return v
}

// Allow reports whether a request from the given key may proceed.
func (rl *RateLimiter) Allow(key string) bool {
	return rl.getVisitor(key).limiter.Allow()
}

// cleanupCycle removes visitors that have been idle for longer than the
// tolerance window.
func (rl *RateLimiter) cleanupCycle() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	threshold := time.Now().Add(-rl.cfg.CleanupTolerance)
	for key, v := range rl.visitors {
		if v.lastSeen.Before(threshold) {
			delete(rl.visitors, key)
		}
	}
}

// cleanupLoop sweeps idle visitors until Stop is called.
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cfg.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanupCycle()
		case <-rl.cleanupDone:
			return
		}
	}
}

// Stop terminates the cleanup goroutine. It is safe to call multiple times.
func (rl *RateLimiter) Stop() {
	rl.stopOnce.Do(func() {
		close(rl.cleanupDone)
	})
}

// clientIP resolves the rate-limit key for a request.
//
// Forwarding headers are attacker-controlled unless a trusted proxy sets them,
// so they are only honoured when TrustProxy is enabled. The right-most value of
// X-Forwarded-For is the address appended by the closest proxy; the left-most
// value is client-controlled and must not be used for security decisions.
func (rl *RateLimiter) clientIP(r *http.Request) string {
	if rl.cfg.TrustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			for i := len(parts) - 1; i >= 0; i-- {
				if ip := strings.TrimSpace(parts[i]); ip != "" {
					return ip
				}
			}
		}
		if ip := strings.TrimSpace(r.Header.Get("X-Real-IP")); ip != "" {
			return ip
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil || host == "" {
		host = r.RemoteAddr
	}
	if host == "" {
		host = "unknown"
	}
	return host
}

// HandlerMiddleware applies the limiter to the next handler and answers 429 with
// Retry-After / X-RateLimit-Reset headers when the budget is exhausted.
func (rl *RateLimiter) HandlerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.Allow(rl.clientIP(r)) {
			w.Header().Set("Retry-After", "60")
			w.Header().Set("X-RateLimit-Reset", time.Now().Add(time.Minute).UTC().Format(time.RFC1123))
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte("429 Too Many Requests"))
			return
		}

		next.ServeHTTP(w, r)
	})
}
