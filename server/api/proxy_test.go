package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func serveStatusRecorder(h http.Handler, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestProxyAuthValidToken(t *testing.T) {
	srv := newAdminTestServer(t, "proxy-admin", "")
	h := srv.Handler()

	_, out := register(t, h, "proxyuser", "supersecret")
	access := out["access_token"].(string)

	// A valid access token is accepted.
	req := newJSONRequest(t, http.MethodPost, "/api/v1/proxy/auth", map[string]any{"access_token": access})
	req.Header.Set("X-Admin-Token", "proxy-admin")
	rec := serveStatusRecorder(h, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("proxy auth valid status = %d, want 200", rec.Code)
	}
	if rec.Body.String() != "{\"status\":\"ok\"}\n" {
		t.Fatalf("proxy auth valid body = %q, want ok", rec.Body.String())
	}
}

func TestProxyAuthDeniesInvalidToken(t *testing.T) {
	srv := newAdminTestServer(t, "proxy-admin", "")
	h := srv.Handler()

	req := newJSONRequest(t, http.MethodPost, "/api/v1/proxy/auth", map[string]any{"access_token": "bogus"})
	req.Header.Set("X-Admin-Token", "proxy-admin")
	rec := serveStatusRecorder(h, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("proxy auth invalid status = %d, want 200", rec.Code)
	}
	if rec.Body.String() != "{\"status\":\"denied\"}\n" {
		t.Fatalf("proxy auth invalid body = %q, want denied", rec.Body.String())
	}
}

func TestProxyAuthRequiresAdminToken(t *testing.T) {
	srv := newAdminTestServer(t, "proxy-admin", "")
	h := srv.Handler()

	req := newJSONRequest(t, http.MethodPost, "/api/v1/proxy/auth", map[string]any{"access_token": "x"})
	req.Header.Set("X-Admin-Token", "wrong")
	rec := serveStatusRecorder(h, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("proxy auth wrong token status = %d, want 401", rec.Code)
	}
}
