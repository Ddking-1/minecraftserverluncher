package api

import (
	"net/http"
	"testing"
)

func TestAccountMe(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Handler()

	_, out := register(t, h, "player9", "supersecret")
	access := out["access_token"].(string)

	code, body := doJSON(t, h, http.MethodGet, "/api/v1/account/me", nil, access)
	if code != http.StatusOK {
		t.Fatalf("account/me status = %d, body = %v", code, body)
	}
	if body["username"] != "player9" {
		t.Fatalf("username = %v, want player9", body["username"])
	}
	if _, ok := body["id"]; !ok {
		t.Fatal("response must include id")
	}
	if _, ok := body["created_at"]; !ok {
		t.Fatal("response must include created_at")
	}
}

func TestAccountMeRequiresAuth(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Handler()

	code, _ := doJSON(t, h, http.MethodGet, "/api/v1/account/me", nil, "")
	if code != http.StatusUnauthorized {
		t.Fatalf("account/me without token status = %d, want 401", code)
	}
}
