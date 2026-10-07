package api

import (
	"testing"
	"time"
)

func TestRateLimiterAllowsUpToLimit(t *testing.T) {
	l := newRateLimiter(3, time.Minute)
	defer l.Close()

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
	defer l.Close()
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
	defer l.Close()
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

func TestRateLimiterGCLeak(t *testing.T) {
	l := newRateLimiter(1, 1*time.Millisecond)
	defer l.Close()
	l.allow("x")
	l.allow("y")
	// Use enough keys and time to trigger potential cleanup backlog.
	for i := 0; i < 50; i++ {
		l.allow(string(rune('a' + i)))
	}
}
