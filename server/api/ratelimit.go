package api

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"minecraftserverluncher/server/config"
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
func (l *rateLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	b, ok := l.buckets[key]
	if !ok || now.After(b.resetAt) {
		l.buckets[key] = &bucket{count: 1, resetAt: now.Add(l.window)}
		return true
	}
	if b.count >= l.limit {
		return false
	}
	b.count++
	return true
}

// limit wraps a handler with per-IP rate limiting.
func (s *Server) limit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r, s.cfg)
		if !s.limiter.allow(ip) {
			writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}
		next(w, r)
	}
}

func clientIP(r *http.Request, cfg *config.Config) string {
	remote := parseRemoteIP(r.RemoteAddr)
	if !isTrustedProxy(remote, cfg) {
		return remote.String()
	}

	if headerIP := forwardedClientIP(r); headerIP != nil {
		return headerIP.String()
	}
	return remote.String()
}

func parseRemoteIP(remoteAddr string) net.IP {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	ip := net.ParseIP(strings.TrimSpace(host))
	if ip == nil {
		// keep deterministic fallback for malformed test and local transports
		return net.ParseIP("0.0.0.0")
	}
	return ip
}

func isTrustedProxy(remote net.IP, cfg *config.Config) bool {
	if cfg == nil || !cfg.TrustedProxyHeaders || remote == nil {
		return false
	}
	for _, cidr := range cfg.TrustedProxyCIDRs {
		if cidr != nil && cidr.Contains(remote) {
			return true
		}
	}
	return false
}

func forwardedClientIP(r *http.Request) net.IP {
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		first := strings.TrimSpace(strings.Split(xff, ",")[0])
		if ip := net.ParseIP(first); ip != nil {
			return ip
		}
	}
	if xrip := strings.TrimSpace(r.Header.Get("X-Real-IP")); xrip != "" {
		if ip := net.ParseIP(xrip); ip != nil {
			return ip
		}
	}
	return nil
}
