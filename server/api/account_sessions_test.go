package api

import (
	"net/http"
	"testing"
)

func TestListSessions(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Handler()

	_, out := register(t, h, "sessuser", "supersecret")
	access := out["access_token"].(string)

	// A second login creates a second session.
	if code, _ := doJSON(t, h, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "sessuser", "password": "supersecret", "device_id": "second",
	}, ""); code != http.StatusOK {
		t.Fatalf("second login status = %d", code)
	}

	code, body := doJSON(t, h, http.MethodGet, "/api/v1/account/sessions", nil, access)
	if code != http.StatusOK {
		t.Fatalf("list sessions status = %d, body = %v", code, body)
	}
	sessions, ok := body["sessions"].([]any)
	if !ok {
		t.Fatalf("sessions is not an array: %T", body["sessions"])
	}
	if len(sessions) != 2 {
		t.Fatalf("session count = %d, want 2", len(sessions))
	}
	first := sessions[0].(map[string]any)
	if _, ok := first["id"]; !ok {
		t.Fatal("session item must include id")
	}
	if _, ok := first["revoked"]; !ok {
		t.Fatal("session item must include revoked")
	}
}

func TestListSessionsEmptyIsArray(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Handler()

	_, out := register(t, h, "emptyuser", "supersecret")
	access := out["access_token"].(string)

	code, body := doJSON(t, h, http.MethodGet, "/api/v1/account/sessions", nil, access)
	if code != http.StatusOK {
		t.Fatalf("list sessions status = %d", code)
	}
	sessions, ok := body["sessions"].([]any)
	if !ok {
		t.Fatalf("sessions must be a JSON array, got %T", body["sessions"])
	}
	if len(sessions) != 1 {
		t.Fatalf("session count = %d, want 1", len(sessions))
	}
}

func TestRevokeSession(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Handler()

	_, out := register(t, h, "revuser", "supersecret")
	access := out["access_token"].(string)

	// Create a second session and capture its refresh token.
	_, out2 := doJSON(t, h, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "revuser", "password": "supersecret", "device_id": "second",
	}, "")
	secondRefresh := out2["refresh_token"].(string)

	// The newest session (highest id) is the second login.
	_, body := doJSON(t, h, http.MethodGet, "/api/v1/account/sessions", nil, access)
	sessions := body["sessions"].([]any)
	if len(sessions) != 2 {
		t.Fatalf("session count = %d, want 2", len(sessions))
	}
	targetID := sessions[0].(map[string]any)["id"].(float64)
	if targetID == 0 {
		t.Fatal("could not find the second session id")
	}

	// Revoke it.
	code, _ := doJSON(t, h, http.MethodPost, "/api/v1/account/sessions/revoke", map[string]any{
		"session_id": int64(targetID),
	}, access)
	if code != http.StatusOK {
		t.Fatalf("revoke status = %d, want 200", code)
	}

	// Its refresh token no longer works.
	if code, _ := doJSON(t, h, http.MethodPost, "/api/v1/auth/refresh", map[string]any{
		"refresh_token": secondRefresh,
	}, ""); code != http.StatusUnauthorized {
		t.Fatalf("refresh after revoke status = %d, want 401", code)
	}
}

func TestRevokeSessionRejectsForeignSession(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Handler()

	_, outA := register(t, h, "userA", "supersecret")
	accessA := outA["access_token"].(string)

	_, outB := register(t, h, "userB", "supersecret")
	accessB := outB["access_token"].(string)

	// userB's session id.
	_, body := doJSON(t, h, http.MethodGet, "/api/v1/account/sessions", nil, accessB)
	sessionsB := body["sessions"].([]any)
	sessionBID := int64(sessionsB[0].(map[string]any)["id"].(float64))

	// userA tries to revoke userB's session.
	code, _ := doJSON(t, h, http.MethodPost, "/api/v1/account/sessions/revoke", map[string]any{
		"session_id": sessionBID,
	}, accessA)
	if code != http.StatusForbidden {
		t.Fatalf("foreign revoke status = %d, want 403", code)
	}
}

func TestRevokeSessionUnknownIsNotFound(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Handler()

	_, out := register(t, h, "revuser2", "supersecret")
	access := out["access_token"].(string)

	code, _ := doJSON(t, h, http.MethodPost, "/api/v1/account/sessions/revoke", map[string]any{
		"session_id": 999999,
	}, access)
	if code != http.StatusNotFound {
		t.Fatalf("unknown session revoke status = %d, want 404", code)
	}
}
