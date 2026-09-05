package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"minecraftserverluncher/server/config"
	"minecraftserverluncher/server/database"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	db, err := database.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("database.Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	cfg := &config.Config{
		ListenAddr:      ":0",
		DBPath:          "unused",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 30 * 24 * time.Hour,
		BcryptCost:      10,
	}
	return NewServer(db, cfg)
}

func doJSON(t *testing.T, h http.Handler, method, path string, body any, token string) (int, map[string]any) {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var out map[string]any
	if rec.Body.Len() > 0 {
		_ = json.Unmarshal(rec.Body.Bytes(), &out)
	}
	return rec.Code, out
}

func register(t *testing.T, h http.Handler, username, password string) (int, map[string]any) {
	t.Helper()
	return doJSON(t, h, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"username": username, "password": password, "device_id": "dev-" + username,
	}, "")
}

func TestRegisterLoginRefreshLogoutFlow(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Handler()

	// Register.
	code, out := register(t, h, "player1", "supersecret")
	if code != http.StatusCreated {
		t.Fatalf("register status = %d, body = %v", code, out)
	}
	refresh := out["refresh_token"].(string)
	access := out["access_token"].(string)
	if refresh == "" || access == "" {
		t.Fatal("register must return access and refresh tokens")
	}

	// Duplicate registration conflicts.
	if code, _ := register(t, h, "player1", "supersecret"); code != http.StatusConflict {
		t.Fatalf("duplicate register status = %d, want 409", code)
	}

	// Login with correct credentials.
	code, out = doJSON(t, h, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "player1", "password": "supersecret", "device_id": "dev-player1",
	}, "")
	if code != http.StatusOK {
		t.Fatalf("login status = %d, body = %v", code, out)
	}
	loginRefresh := out["refresh_token"].(string)

	// Login with wrong password is rejected.
	if code, _ := doJSON(t, h, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "player1", "password": "wrongpass",
	}, ""); code != http.StatusUnauthorized {
		t.Fatalf("bad login status = %d, want 401", code)
	}

	// Refresh the original session.
	code, out = doJSON(t, h, http.MethodPost, "/api/v1/auth/refresh", map[string]any{
		"refresh_token": refresh,
	}, "")
	if code != http.StatusOK {
		t.Fatalf("refresh status = %d, body = %v", code, out)
	}
	newAccess := out["access_token"].(string)
	if newAccess == "" {
		t.Fatal("refresh must return a new access token")
	}

	// Logout-all using the new access token revokes every session.
	code, out = doJSON(t, h, http.MethodPost, "/api/v1/auth/logout-all", nil, newAccess)
	if code != http.StatusOK {
		t.Fatalf("logout-all status = %d, body = %v", code, out)
	}

	// The login session's refresh token is now revoked.
	code, _ = doJSON(t, h, http.MethodPost, "/api/v1/auth/refresh", map[string]any{
		"refresh_token": loginRefresh,
	}, "")
	if code != http.StatusUnauthorized {
		t.Fatalf("refresh after logout-all status = %d, want 401", code)
	}
}

func TestLogoutRevokesSingleSession(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Handler()

	_, out := register(t, h, "player2", "supersecret")
	refresh := out["refresh_token"].(string)

	// Logout revokes this session only.
	code, _ := doJSON(t, h, http.MethodPost, "/api/v1/auth/logout", map[string]any{
		"refresh_token": refresh,
	}, "")
	if code != http.StatusOK {
		t.Fatalf("logout status = %d", code)
	}

	code, _ = doJSON(t, h, http.MethodPost, "/api/v1/auth/refresh", map[string]any{
		"refresh_token": refresh,
	}, "")
	if code != http.StatusUnauthorized {
		t.Fatalf("refresh after logout status = %d, want 401", code)
	}
}

func TestLogoutAllRequiresAccessToken(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Handler()

	code, _ := doJSON(t, h, http.MethodPost, "/api/v1/auth/logout-all", nil, "")
	if code != http.StatusUnauthorized {
		t.Fatalf("logout-all without token status = %d, want 401", code)
	}
}

func TestRegisterValidation(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Handler()

	// Short password rejected.
	if code, _ := register(t, h, "player3", "short"); code != http.StatusBadRequest {
		t.Fatalf("short password status = %d, want 400", code)
	}
	// Invalid username rejected.
	if code, _ := register(t, h, "a!", "supersecret"); code != http.StatusBadRequest {
		t.Fatalf("invalid username status = %d, want 400", code)
	}
}
