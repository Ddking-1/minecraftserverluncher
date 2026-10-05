package api

import (
	"net/http"
	"testing"
)

func TestLogoutOtherDevices(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Handler()

	_, out := register(t, h, "multi", "supersecret")
	firstAccess := out["access_token"].(string)
	firstRefresh := out["refresh_token"].(string)

	// Second login creates a second session.
	_, out2 := doJSON(t, h, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "multi", "password": "supersecret", "device_id": "second",
	}, "")
	if out2["access_token"] == "" {
		t.Fatal("second login must return a token")
	}
	secondRefresh := out2["refresh_token"].(string)

	// Log out from all OTHER devices using the FIRST session's token.
	code, _ := doJSON(t, h, http.MethodPost, "/api/v1/account/sessions/logout-others", nil, firstAccess)
	if code != http.StatusOK {
		t.Fatalf("logout-others status = %d, want 200", code)
	}

	// The second session's refresh token is now revoked.
	if code, _ := doJSON(t, h, http.MethodPost, "/api/v1/auth/refresh", map[string]any{
		"refresh_token": secondRefresh,
	}, ""); code != http.StatusUnauthorized {
		t.Fatalf("refresh after logout-others status = %d, want 401", code)
	}

	// The first (current) session's refresh token still works.
	if code, _ := doJSON(t, h, http.MethodPost, "/api/v1/auth/refresh", map[string]any{
		"refresh_token": firstRefresh,
	}, ""); code != http.StatusOK {
		t.Fatalf("refresh of kept session status = %d, want 200", code)
	}
}

func TestLogoutOtherDevicesRequiresAuth(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Handler()

	code, _ := doJSON(t, h, http.MethodPost, "/api/v1/account/sessions/logout-others", nil, "")
	if code != http.StatusUnauthorized {
		t.Fatalf("logout-others without token status = %d, want 401", code)
	}
}
