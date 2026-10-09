package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeadersPresent(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("healthz status = %d", rec.Code)
	}
	want := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"Cache-Control":          "no-store",
		"Referrer-Policy":        "no-referrer",
	}
	for k, v := range want {
		if got := rec.Header().Get(k); got != v {
			t.Errorf("header %s = %q, want %q", k, got, v)
		}
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	panicking := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})
	h := srv.withRecovery(panicking)

	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("panic status = %d, want 500", rec.Code)
	}
}

func TestRequestLogPreservesStatus(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	h := srv.withRequestLog(inner)

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Fatalf("status = %d, want 418", rec.Code)
	}
}

func TestRequestIDGenerated(t *testing.T) {
	var seen string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = requestID(r)
	})
	h := withRequestID(inner)

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if seen == "" {
		t.Fatal("request id not stored in context")
	}
	if rec.Header().Get("X-Request-Id") != seen {
		t.Fatalf("response header %q != context %q", rec.Header().Get("X-Request-Id"), seen)
	}
}

func TestRequestIDEchoed(t *testing.T) {
	const provided = "abc123"
	var seen string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = requestID(r)
	})
	h := withRequestID(inner)

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("X-Request-Id", provided)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if seen != provided {
		t.Fatalf("context id = %q, want %q", seen, provided)
	}
	if rec.Header().Get("X-Request-Id") != provided {
		t.Fatalf("response header = %q, want %q", rec.Header().Get("X-Request-Id"), provided)
	}
}

func TestRequestIDUnique(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	h := withRequestID(inner)

	first := httptest.NewRecorder()
	h.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/x", nil))
	second := httptest.NewRecorder()
	h.ServeHTTP(second, httptest.NewRequest(http.MethodGet, "/x", nil))

	if first.Header().Get("X-Request-Id") == second.Header().Get("X-Request-Id") {
		t.Fatal("generated request ids should differ")
	}
}
