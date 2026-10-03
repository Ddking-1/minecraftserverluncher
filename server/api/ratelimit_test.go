package api

import (
	"testing"
	"time"
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
