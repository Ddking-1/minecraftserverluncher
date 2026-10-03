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

func TestChangePassword(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Handler()

	_, out := register(t, h, "pwuser", "supersecret")
	access := out["access_token"].(string)

	// Wrong current password is rejected.
	code, _ := doJSON(t, h, http.MethodPost, "/api/v1/account/password", map[string]any{
		"current_password": "wrongpass", "new_password": "newsecret1",
	}, access)
	if code != http.StatusUnauthorized {
		t.Fatalf("wrong current password status = %d, want 401", code)
	}

	// Too-short new password is rejected.
	code, _ = doJSON(t, h, http.MethodPost, "/api/v1/account/password", map[string]any{
		"current_password": "supersecret", "new_password": "short",
	}, access)
	if code != http.StatusBadRequest {
		t.Fatalf("short new password status = %d, want 400", code)
	}

	// Successful change.
	code, _ = doJSON(t, h, http.MethodPost, "/api/v1/account/password", map[string]any{
		"current_password": "supersecret", "new_password": "newsecret1",
	}, access)
	if code != http.StatusOK {
		t.Fatalf("change password status = %d, want 200", code)
	}

	// Old password no longer works.
	if code, _ := doJSON(t, h, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "pwuser", "password": "supersecret",
	}, ""); code != http.StatusUnauthorized {
		t.Fatalf("old password login status = %d, want 401", code)
	}

	// New password works.
	if code, _ := doJSON(t, h, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "pwuser", "password": "newsecret1",
	}, ""); code != http.StatusOK {
		t.Fatalf("new password login status = %d, want 200", code)
	}
}

func TestChangePasswordRevokesSessions(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Handler()

	_, out := register(t, h, "pwuser2", "supersecret")
	access := out["access_token"].(string)

	code, _ := doJSON(t, h, http.MethodPost, "/api/v1/account/password", map[string]any{
		"current_password": "supersecret", "new_password": "newsecret2",
	}, access)
	if code != http.StatusOK {
		t.Fatalf("change password status = %d, want 200", code)
	}

	// The access token used for the change is now revoked (logout-all).
	code, _ = doJSON(t, h, http.MethodGet, "/api/v1/account/me", nil, access)
	if code != http.StatusUnauthorized {
		t.Fatalf("access token after password change status = %d, want 401", code)
	}
}
