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

func newJSONRequest(t *testing.T, method, path string, body any) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func serveStatus(h http.Handler, req *http.Request) int {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Code
}

func newAdminTestServer(t *testing.T, adminToken, script string) *Server {
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
		AdminToken:      adminToken,
		PteroScript:     script,
	}
	return NewServer(db, cfg, nil)
}

func doAdminJSON(t *testing.T, h http.Handler, method, path string, body any, adminToken string) int {
	t.Helper()
	req := newJSONRequest(t, method, path, body)
	if adminToken != "" {
		req.Header.Set("X-Admin-Token", adminToken)
	}
	return serveStatus(h, req)
}

func TestServerStartDisabledWithoutAdminToken(t *testing.T) {
	srv := newAdminTestServer(t, "", "does-not-exist.py")
	defer srv.Close()
	h := srv.Handler()

	// No admin token configured -> endpoint must be indistinguishable from
	// a non-existent route.
	if code := doAdminJSON(t, h, http.MethodPost, "/api/server/start", nil, "anything"); code != http.StatusNotFound {
		t.Fatalf("start with no admin token configured status = %d, want 404", code)
	}
}

func TestServerStartRequiresCorrectAdminToken(t *testing.T) {
	srv := newAdminTestServer(t, "secret-admin", "does-not-exist.py")
	defer srv.Close()
	h := srv.Handler()

	if code := doAdminJSON(t, h, http.MethodPost, "/api/server/start", nil, "wrong"); code != http.StatusUnauthorized {
		t.Fatalf("start with wrong admin token status = %d, want 401", code)
	}
	if code := doAdminJSON(t, h, http.MethodPost, "/api/server/start", nil, ""); code != http.StatusUnauthorized {
		t.Fatalf("start with missing admin token status = %d, want 401", code)
	}
	// Correct token but missing script: internal error, never raw output.
	if code := doAdminJSON(t, h, http.MethodPost, "/api/server/start", nil, "secret-admin"); code != http.StatusInternalServerError {
		t.Fatalf("start with correct token and bad script status = %d, want 500", code)
	}
}

func TestAdminHWIDBanLifecycle(t *testing.T) {
	srv := newAdminTestServer(t, "secret-admin", "")
	defer srv.Close()
	h := srv.Handler()

	// Ban requires the admin token.
	if code := doAdminJSON(t, h, http.MethodPost, "/api/admin/hwid/ban", map[string]any{"device_id": "hw-1"}, "wrong"); code != http.StatusUnauthorized {
		t.Fatalf("ban with wrong token status = %d, want 401", code)
	}

	// Ban a device.
	if code := doAdminJSON(t, h, http.MethodPost, "/api/admin/hwid/ban", map[string]any{"device_id": "hw-1"}, "secret-admin"); code != http.StatusOK {
		t.Fatalf("ban status = %d, want 200", code)
	}

	// The login path must now reject that device.
	_, out := register(t, h, "bannedplayer", "supersecret")
	_ = out
	code, _ := doJSON(t, h, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "bannedplayer", "password": "supersecret", "device_id": "hw-1",
	}, "")
	if code != http.StatusForbidden {
		t.Fatalf("login with banned device status = %d, want 403", code)
	}

	// List contains the ban (requires the admin token).
	req := newJSONRequest(t, http.MethodGet, "/api/admin/hwid/bans", nil)
	req.Header.Set("X-Admin-Token", "secret-admin")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list bans status = %d, want 200", rec.Code)
	}
	var listed struct {
		Bans []string `json:"bans"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode bans: %v", err)
	}
	found := false
	for _, b := range listed.Bans {
		if b == "hw-1" {
			found = true
		}
	}
	if !found {
		t.Fatalf("bans list %v does not contain hw-1", listed.Bans)
	}

	// Unban and verify login succeeds again.
	if code := doAdminJSON(t, h, http.MethodPost, "/api/admin/hwid/unban", map[string]any{"device_id": "hw-1"}, "secret-admin"); code != http.StatusOK {
		t.Fatalf("unban status = %d, want 200", code)
	}
	code, _ = doJSON(t, h, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "bannedplayer", "password": "supersecret", "device_id": "hw-1",
	}, "")
	if code != http.StatusOK {
		t.Fatalf("login after unban status = %d, want 200", code)
	}
}
