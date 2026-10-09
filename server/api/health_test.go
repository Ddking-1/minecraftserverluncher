package api

import (
	"net/http"
	"testing"
)

func TestReadyzHealthy(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	code, body := doJSON(t, h, http.MethodGet, "/readyz", nil, "")
	if code != http.StatusOK {
		t.Fatalf("readyz status = %d, body = %v", code, body)
	}
	if body["status"] != "ready" {
		t.Fatalf("readyz status field = %v, want ready", body["status"])
	}
}

func TestReadyzAfterDBClose(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	if err := srv.db.Close(); err != nil {
		t.Fatalf("close db: %v", err)
	}

	code, _ := doJSON(t, h, http.MethodGet, "/readyz", nil, "")
	if code != http.StatusServiceUnavailable {
		t.Fatalf("readyz after close status = %d, want 503", code)
	}
}

func TestServerCloseIdempotent(t *testing.T) {
	srv := newTestServer(t)
	srv.Close()
	srv.Close()
}
