package api

import (
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"minecraftserverluncher/server/config"
	"minecraftserverluncher/server/database"
)

func TestCapabilitiesEndpoint(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	code, body := doJSON(t, h, http.MethodGet, "/api/v1/capabilities", nil, "")
	if code != http.StatusOK {
		t.Fatalf("capabilities status = %d, body = %v", code, body)
	}
	feats, ok := body["features"].(map[string]any)
	if !ok {
		t.Fatalf("capabilities missing features map: %v", body)
	}
	for _, k := range []string{"invites", "captcha", "crash_reports", "modpacks", "offline_mode", "require_device_id"} {
		if _, ok := feats[k]; !ok {
			t.Fatalf("capabilities missing feature key %q", k)
		}
	}
}

func TestNewsEndpoint(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	code, body := doJSON(t, h, http.MethodGet, "/api/v1/news", nil, "")
	if code != http.StatusOK {
		t.Fatalf("news status = %d, body = %v", code, body)
	}
	items, ok := body["items"].([]any)
	if !ok || len(items) == 0 {
		t.Fatalf("news missing items: %v", body)
	}
}

func TestCrashReportRequiresAuth(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	code, _ := doJSON(t, h, http.MethodPost, "/api/v1/crash-reports", map[string]any{"message": "boom"}, "")
	if code != http.StatusUnauthorized {
		t.Fatalf("crash-report without auth status = %d, want 401", code)
	}
}

func TestRequireDeviceIDWhenEnabled(t *testing.T) {
	db, err := database.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("database.Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	cfg := &config.Config{
		ListenAddr:      ":0",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 30 * 24 * time.Hour,
		BcryptCost:      10,
		RequireDeviceID: true,
	}
	srv := NewServer(db, cfg, nil)
	srv.limiter = NewRateLimiter(Config{RequestsPerMinute: 10000})
	defer srv.Close()
	h := srv.Handler()

	register(t, h, "devuser", "supersecret")
	// login without device_id should be rejected
	code, _ := doJSON(t, h, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "devuser", "password": "supersecret",
	}, "")
	if code != http.StatusBadRequest {
		t.Fatalf("login without device_id status = %d, want 400", code)
	}
	// login with device_id should succeed
	code, _ = doJSON(t, h, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "devuser", "password": "supersecret", "device_id": "dev-1",
	}, "")
	if code != http.StatusOK {
		t.Fatalf("login with device_id status = %d, want 200", code)
	}
}

func TestCrashReportAccepted(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	// register a user to get a token
	code, out := register(t, h, "crashuser", "supersecret")
	if code != http.StatusCreated {
		t.Fatalf("register status = %d", code)
	}
	token := out["access_token"].(string)

	code, body := doJSON(t, h, http.MethodPost, "/api/v1/crash-reports", map[string]any{"message": "boom"}, token)
	if code != http.StatusAccepted {
		t.Fatalf("crash-report status = %d, body = %v, want 202", code, body)
	}
}
