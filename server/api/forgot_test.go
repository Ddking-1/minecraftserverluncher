package api

import (
	"net/http"
	"testing"

	"minecraftserverluncher/server/auth"
)

// hashTokenForTest mirrors the server's storage hashing for a test token.
func hashTokenForTest(token string) string { return auth.HashToken(token) }

func TestForgotPasswordDoesNotEnumerate(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	// Known account.
	_, _ = register(t, h, "forgotuser", "supersecret")

	// Known identifier and unknown identifier must both return 200 with the
	// same generic message (no enumeration).
	for _, id := range []string{"forgotuser", "doesnotexist"} {
		code, body := doJSON(t, h, http.MethodPost, "/api/v1/auth/forgot", map[string]any{
			"identifier": id,
		}, "")
		if code != http.StatusOK {
			t.Fatalf("forgot(%q) status = %d, want 200", id, code)
		}
		if body["status"] == "" {
			t.Fatalf("forgot(%q) must return a status message", id)
		}
	}
}

func TestForgotPasswordRequiresIdentifier(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	code, _ := doJSON(t, h, http.MethodPost, "/api/v1/auth/forgot", map[string]any{
		"identifier": "",
	}, "")
	if code != http.StatusBadRequest {
		t.Fatalf("empty identifier status = %d, want 400", code)
	}
}

func TestResetPasswordLifecycle(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	_, _ = register(t, h, "resetuser", "supersecret")

	// Request a reset for the known user. (Server mints the token; for tests
	// we create one directly through the DB so we can consume it.)
	acc, err := srv.db.AccountByUsername("resetuser")
	if err != nil {
		t.Fatalf("account lookup: %v", err)
	}
	// Mint a reset token directly (bypassing the delivery channel).
	resetToken := "test-reset-token-123"
	if err := srv.db.CreatePasswordReset(acc.ID, hashTokenForTest(resetToken)); err != nil {
		t.Fatalf("CreatePasswordReset: %v", err)
	}

	// Invalid token rejected.
	code, _ := doJSON(t, h, http.MethodPost, "/api/v1/auth/reset", map[string]any{
		"token": "bogus", "new_password": "newsecret1",
	}, "")
	if code != http.StatusBadRequest {
		t.Fatalf("bogus reset token status = %d, want 400", code)
	}

	// Valid token resets the password.
	code, _ = doJSON(t, h, http.MethodPost, "/api/v1/auth/reset", map[string]any{
		"token": resetToken, "new_password": "newsecret1",
	}, "")
	if code != http.StatusOK {
		t.Fatalf("reset status = %d, want 200", code)
	}

	// New password works, old does not.
	if code, _ := doJSON(t, h, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "resetuser", "password": "newsecret1",
	}, ""); code != http.StatusOK {
		t.Fatalf("new password login status = %d, want 200", code)
	}
	if code, _ := doJSON(t, h, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "resetuser", "password": "supersecret",
	}, ""); code != http.StatusUnauthorized {
		t.Fatalf("old password login status = %d, want 401", code)
	}

	// Token is single-use: second use rejected.
	code, _ = doJSON(t, h, http.MethodPost, "/api/v1/auth/reset", map[string]any{
		"token": resetToken, "new_password": "newsecret2",
	}, "")
	if code != http.StatusBadRequest {
		t.Fatalf("reuse of reset token status = %d, want 400", code)
	}
}
