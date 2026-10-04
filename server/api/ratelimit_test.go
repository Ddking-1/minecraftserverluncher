package api

import (
	"net/http"
	"net/http/httptest"
	"net"
	"testing"
	"time"

	"minecraftserverluncher/server/config"
)

func TestRateLimiterAllowsUpToLimit(t *testing.T) {
	l := newRateLimiter(3, time.Minute)
	for i := 0; i < 3; i++ {
		if !l.allow("1.2.3.4") {
			t.Fatalf("request %d should be allowed", i)
		}
	}
	if l.allow("1.2.3.4") {
		t.Fatal("4th request should be denied")
	}
}

func TestRateLimiterIsPerKey(t *testing.T) {
	l := newRateLimiter(1, time.Minute)
	if !l.allow("a") {
		t.Fatal("a should be allowed")
	}
	if l.allow("a") {
		t.Fatal("a should be denied")
	}
	if !l.allow("b") {
		t.Fatal("b should be allowed independently")
	}
}

func TestRateLimiterWindowResets(t *testing.T) {
	l := newRateLimiter(1, 10*time.Millisecond)
	if !l.allow("k") {
		t.Fatal("first request should be allowed")
	}
	if l.allow("k") {
		t.Fatal("second request should be denied")
	}
	time.Sleep(20 * time.Millisecond)
	if !l.allow("k") {
		t.Fatal("request should be allowed after window reset")
	}
}

func TestClientIPDirectModeIgnoresForwardHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.5:1234"
	req.Header.Set("X-Forwarded-For", "198.51.100.9")
	got := clientIP(req, &config.Config{})
	if got != "10.0.0.5" {
		t.Fatalf("clientIP = %q, want remote ip", got)
	}
}

func TestClientIPTrustedProxyUsesForwardedFor(t *testing.T) {
	_, trusted, _ := net.ParseCIDR("127.0.0.0/8")
	cfg := &config.Config{
		TrustedProxyHeaders: true,
		TrustedProxyCIDRs:   []*net.IPNet{trusted},
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:4321"
	req.Header.Set("X-Forwarded-For", "198.51.100.9, 127.0.0.1")
	got := clientIP(req, cfg)
	if got != "198.51.100.9" {
		t.Fatalf("clientIP = %q, want forwarded client", got)
	}
}

func TestRateLimitIsSeparatedPerForwardedClientWhenTrusted(t *testing.T) {
	_, trusted, _ := net.ParseCIDR("127.0.0.0/8")
	s := &Server{
		cfg: &config.Config{
			TrustedProxyHeaders: true,
			TrustedProxyCIDRs:   []*net.IPNet{trusted},
		},
		limiter: newRateLimiter(1, time.Minute),
	}
	h := s.limit(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "127.0.0.1:1000"
	req1.Header.Set("X-Forwarded-For", "198.51.100.1")
	rec1 := httptest.NewRecorder()
	h(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("first client got %d", rec1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "127.0.0.1:1001"
	req2.Header.Set("X-Forwarded-For", "198.51.100.2")
	rec2 := httptest.NewRecorder()
	h(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("second client got %d", rec2.Code)
	}

	req3 := httptest.NewRequest(http.MethodGet, "/", nil)
	req3.RemoteAddr = "127.0.0.1:1002"
	req3.Header.Set("X-Forwarded-For", "198.51.100.1")
	rec3 := httptest.NewRecorder()
	h(rec3, req3)
	if rec3.Code != http.StatusTooManyRequests {
		t.Fatalf("repeat first client got %d, want 429", rec3.Code)
	}
}

func TestUntrustedProxyCannotSpoofForwardedIP(t *testing.T) {
	_, trusted, _ := net.ParseCIDR("127.0.0.0/8")
	s := &Server{
		cfg: &config.Config{
			TrustedProxyHeaders: true,
			TrustedProxyCIDRs:   []*net.IPNet{trusted},
		},
		limiter: newRateLimiter(1, time.Minute),
	}
	h := s.limit(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "10.0.0.9:2000"
	req1.Header.Set("X-Forwarded-For", "198.51.100.1")
	rec1 := httptest.NewRecorder()
	h(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("first request got %d", rec1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "10.0.0.9:2001"
	req2.Header.Set("X-Forwarded-For", "198.51.100.2")
	rec2 := httptest.NewRecorder()
	h(rec2, req2)
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("spoofed second request got %d, want 429", rec2.Code)
	}
}
