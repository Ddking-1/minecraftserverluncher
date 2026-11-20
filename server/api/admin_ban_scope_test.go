package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"minecraftserverluncher/server/config"
	"minecraftserverluncher/server/database"
)

// This test proves the ban-scope isolation contract against the real handler
// stack (routing, admin auth, SQLite, sessions and ban enforcement), not a
// mock. Three distinct people are created: alice and carol share an IP but
// have different hardware IDs, bob is on a different IP, and dave shares one
// of alice's hardware IDs.

const isolationAdminToken = "isolation-admin-token"

func newIsolationServer(t *testing.T) *Server {
	t.Helper()
	db, err := database.Open(filepath.Join(t.TempDir(), "isolation.db"))
	if err != nil {
		t.Fatalf("database.Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	cfg := &config.Config{
		ListenAddr:      ":0",
		DBPath:          "unused",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 30 * 24 * time.Hour,
		BcryptCost:      4,
		AdminToken:      isolationAdminToken,
		TrustProxy:      true,
	}
	srv := NewServer(db, cfg, nil)
	srv.limiter = NewRateLimiter(Config{RequestsPerMinute: 100000, TrustProxy: true})
	return srv
}

func isolationRequest(t *testing.T, h http.Handler, method, path string, body any, headers map[string]string) (int, map[string]any) {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var out map[string]any
	if rec.Body.Len() > 0 {
		_ = json.Unmarshal(rec.Body.Bytes(), &out)
	}
	return rec.Code, out
}

func registerAt(t *testing.T, h http.Handler, username, device, ip string) {
	t.Helper()
	code, out := isolationRequest(t, h, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"username": username, "password": "password123", "device_id": device, "device_name": "PC",
	}, map[string]string{"X-Forwarded-For": ip})
	if code != http.StatusCreated {
		t.Fatalf("register %s = %d, body = %v", username, code, out)
	}
}

func loginAt(t *testing.T, h http.Handler, username, device, ip string) (int, string) {
	t.Helper()
	code, out := isolationRequest(t, h, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": username, "password": "password123", "device_id": device, "device_name": "PC",
	}, map[string]string{"X-Forwarded-For": ip})
	token, _ := out["access_token"].(string)
	return code, token
}

func adminBan(t *testing.T, h http.Handler, accountID int64, body map[string]any) (int, map[string]any) {
	t.Helper()
	return isolationRequest(t, h, http.MethodPost, fmt.Sprintf("/api/admin/users/%d/ban", accountID), body,
		map[string]string{"X-Admin-Token": isolationAdminToken})
}

func adminUnbanAll(t *testing.T, h http.Handler, accountID int64) {
	t.Helper()
	code, out := isolationRequest(t, h, http.MethodPost, fmt.Sprintf("/api/admin/users/%d/unban", accountID),
		map[string]any{"targets": []string{"account", "hwid", "ip"}},
		map[string]string{"X-Admin-Token": isolationAdminToken})
	if code != http.StatusOK {
		t.Fatalf("unban %d = %d, body = %v", accountID, code, out)
	}
}

func adminGet(t *testing.T, h http.Handler, path string) (int, map[string]any) {
	t.Helper()
	return isolationRequest(t, h, http.MethodGet, path, nil, map[string]string{"X-Admin-Token": isolationAdminToken})
}

func accountID(t *testing.T, srv *Server, username string) int64 {
	t.Helper()
	acc, err := srv.db.AccountByUsername(username)
	if err != nil {
		t.Fatalf("AccountByUsername(%s): %v", username, err)
	}
	return acc.ID
}

func countOf(body map[string]any, key string) int {
	counts, _ := body["counts"].(map[string]any)
	v, _ := counts[key].(float64)
	return int(v)
}

func TestAdminBanScopeIsolation(t *testing.T) {
	srv := newIsolationServer(t)
	defer srv.Close()
	h := srv.Handler()

	// alice: two hardware IDs, IP 10.0.0.1. bob: own IP/device. carol: same IP
	// as alice, own device. dave: shares alice's first hardware ID.
	registerAt(t, h, "alice", "alice-dev-1", "10.0.0.1")
	if code, _ := loginAt(t, h, "alice", "alice-dev-2", "10.0.0.1"); code != http.StatusOK {
		t.Fatalf("alice second device login = %d", code)
	}
	registerAt(t, h, "bob", "bob-dev", "10.0.0.2")
	registerAt(t, h, "carol", "carol-dev", "10.0.0.1")
	registerAt(t, h, "dave", "alice-dev-1", "10.0.0.3")

	aliceID := accountID(t, srv, "alice")

	expectLogin := func(user, device, ip string, want int) {
		t.Helper()
		if code, _ := loginAt(t, h, user, device, ip); code != want {
			t.Fatalf("%s login from %s (device %s) = %d, want %d", user, ip, device, code, want)
		}
	}

	// Baseline: everyone can log in before any ban.
	expectLogin("alice", "alice-dev-1", "10.0.0.1", http.StatusOK)
	expectLogin("bob", "bob-dev", "10.0.0.2", http.StatusOK)
	expectLogin("carol", "carol-dev", "10.0.0.1", http.StatusOK)
	expectLogin("dave", "alice-dev-1", "10.0.0.3", http.StatusOK)

	t.Run("account ban affects only the target account", func(t *testing.T) {
		code, body := adminBan(t, h, aliceID, map[string]any{
			"duration_minutes": 60,
			"reason":           "account-only test",
			"scope":            "both",
			"targets":          []string{"account"},
			"lookback_days":    30,
		})
		if code != http.StatusOK {
			t.Fatalf("account ban = %d, body = %v", code, body)
		}
		if got := countOf(body, "account"); got != 1 {
			t.Fatalf("account ban count = %d, want 1", got)
		}
		expectLogin("alice", "alice-dev-1", "10.0.0.1", http.StatusForbidden)
		expectLogin("bob", "bob-dev", "10.0.0.2", http.StatusOK)
		expectLogin("carol", "carol-dev", "10.0.0.1", http.StatusOK)
		expectLogin("dave", "alice-dev-1", "10.0.0.3", http.StatusOK)
		adminUnbanAll(t, h, aliceID)
	})

	t.Run("hwid ban targets only hardware the account used", func(t *testing.T) {
		code, body := adminBan(t, h, aliceID, map[string]any{
			"duration_minutes": 60,
			"reason":           "hwid-only test",
			"scope":            "both",
			"targets":          []string{"hwid"},
			"lookback_days":    30,
		})
		if code != http.StatusOK {
			t.Fatalf("hwid ban = %d, body = %v", code, body)
		}
		if got := countOf(body, "hwid"); got != 2 {
			t.Fatalf("hwid ban count = %d, want 2 (alice's two devices)", got)
		}
		expectLogin("alice", "alice-dev-1", "10.0.0.1", http.StatusForbidden)
		expectLogin("alice", "alice-dev-2", "10.0.0.1", http.StatusForbidden)
		// Same IP, different hardware: must stay usable.
		expectLogin("carol", "carol-dev", "10.0.0.1", http.StatusOK)
		expectLogin("bob", "bob-dev", "10.0.0.2", http.StatusOK)
		// dave actually used one of alice's banned hardware IDs.
		expectLogin("dave", "alice-dev-1", "10.0.0.3", http.StatusForbidden)
		// A device alice never used is not blanket-banned.
		expectLogin("alice", "alice-dev-new", "10.0.0.1", http.StatusOK)
		adminUnbanAll(t, h, aliceID)
	})

	t.Run("ip ban affects only accounts on that ip", func(t *testing.T) {
		code, body := adminBan(t, h, aliceID, map[string]any{
			"duration_minutes": 60,
			"reason":           "ip-only test",
			"scope":            "both",
			"targets":          []string{"ip"},
			"lookback_days":    30,
		})
		if code != http.StatusOK {
			t.Fatalf("ip ban = %d, body = %v", code, body)
		}
		if got := countOf(body, "ip"); got != 1 {
			t.Fatalf("ip ban count = %d, want 1 (alice used one ip)", got)
		}
		if got, _ := body["affected_ip_count"].(float64); int(got) != 2 {
			t.Fatalf("affected_ip_count = %v, want 2 (alice + carol)", body["affected_ip_count"])
		}
		expectLogin("alice", "alice-dev-1", "10.0.0.1", http.StatusForbidden)
		expectLogin("carol", "carol-dev", "10.0.0.1", http.StatusForbidden)
		expectLogin("bob", "bob-dev", "10.0.0.2", http.StatusOK)
		expectLogin("dave", "alice-dev-1", "10.0.0.3", http.StatusOK)

		// The count the UI shows must match the real sibling-account count.
		code, details := adminGet(t, h, fmt.Sprintf("/api/admin/users/%d/details", aliceID))
		if code != http.StatusOK {
			t.Fatalf("details = %d, body = %v", code, details)
		}
		ips, _ := details["ips_last_30d"].([]any)
		if len(ips) != 1 {
			t.Fatalf("ips_last_30d length = %d, want 1 (%v)", len(ips), details["ips_last_30d"])
		}
		first, _ := ips[0].(map[string]any)
		if got, _ := first["accounts_last_30d"].(float64); int(got) != 2 {
			t.Fatalf("UI ip account count = %v, want 2", first["accounts_last_30d"])
		}
		if first["ip"] != "10.0.0.1" {
			t.Fatalf("UI ip = %v, want 10.0.0.1", first["ip"])
		}

		code, ipAccounts := adminGet(t, h, "/api/admin/ip-accounts?ip=10.0.0.1")
		if code != http.StatusOK {
			t.Fatalf("ip-accounts = %d, body = %v", code, ipAccounts)
		}
		if got, _ := ipAccounts["count"].(float64); int(got) != 2 {
			t.Fatalf("ip-accounts count = %v, want 2", ipAccounts["count"])
		}
		adminUnbanAll(t, h, aliceID)
	})

	t.Run("missing hwid data does not widen the ban", func(t *testing.T) {
		registerAt(t, h, "erin", "erin-dev", "10.0.0.4")
		erinID := accountID(t, srv, "erin")
		// Age the only device past the 30-day window.
		if _, err := srv.db.SQL().Exec(`UPDATE devices SET last_seen = ? WHERE account_id = ?`,
			time.Now().UTC().AddDate(0, 0, -45).Format(time.RFC3339Nano), erinID); err != nil {
			t.Fatalf("age device: %v", err)
		}
		code, body := adminBan(t, h, erinID, map[string]any{
			"duration_minutes": 60,
			"reason":           "hwid-missing test",
			"scope":            "both",
			"targets":          []string{"hwid"},
			"lookback_days":    30,
		})
		if code != http.StatusOK {
			t.Fatalf("hwid ban with no recent devices = %d, body = %v", code, body)
		}
		if got := countOf(body, "hwid"); got != 0 {
			t.Fatalf("hwid ban count = %d, want 0", got)
		}
		code, details := adminGet(t, h, fmt.Sprintf("/api/admin/users/%d/details", erinID))
		if code != http.StatusOK {
			t.Fatalf("details = %d, body = %v", code, details)
		}
		if devs, _ := details["devices_last_30d"].([]any); len(devs) != 0 {
			t.Fatalf("UI hwid preview = %v, want empty", details["devices_last_30d"])
		}
		expectLogin("erin", "erin-dev", "10.0.0.4", http.StatusOK)
		adminUnbanAll(t, h, erinID)
	})

	t.Run("mc scope keeps the reason and duration for the proxy", func(t *testing.T) {
		code, token := loginAt(t, h, "alice", "alice-dev-1", "10.0.0.1")
		if code != http.StatusOK || token == "" {
			t.Fatalf("alice login before mc ban = %d, token empty = %v", code, token == "")
		}
		code, body := adminBan(t, h, aliceID, map[string]any{
			"duration_minutes": 90,
			"reason":           "Griefing spawn",
			"scope":            "mc",
			"targets":          []string{"account"},
			"lookback_days":    30,
		})
		if code != http.StatusOK {
			t.Fatalf("mc ban = %d, body = %v", code, body)
		}
		// mc scope must not block the launcher login.
		expectLogin("alice", "alice-dev-1", "10.0.0.1", http.StatusOK)

		// The proxy is what the Minecraft server reads; it must carry the reason.
		code, proxy := isolationRequest(t, h, http.MethodPost, "/api/v1/proxy/auth",
			map[string]any{"access_token": token},
			map[string]string{"X-Admin-Token": isolationAdminToken})
		if code != http.StatusOK {
			t.Fatalf("proxy auth = %d, body = %v", code, proxy)
		}
		if proxy["status"] != "denied" {
			t.Fatalf("proxy status = %v, want denied", proxy["status"])
		}
		if proxy["reason"] != "Griefing spawn" {
			t.Fatalf("proxy reason = %v, want Griefing spawn", proxy["reason"])
		}
		if got, _ := proxy["duration_minutes"].(float64); int(got) != 90 {
			t.Fatalf("proxy duration = %v, want 90", proxy["duration_minutes"])
		}
		if expires, _ := proxy["expires_at"].(string); expires == "" {
			t.Fatalf("proxy expires_at missing: %v", proxy)
		}

		// The stored record carries scope, reason, duration and expiry.
		code, bans := adminGet(t, h, "/api/admin/bans?limit=50")
		if code != http.StatusOK {
			t.Fatalf("bans = %d, body = %v", code, bans)
		}
		found := false
		for _, raw := range bans["bans"].([]any) {
			rec, _ := raw.(map[string]any)
			if rec["type"] != "account" || rec["account_id"] == nil {
				continue
			}
			if int64(rec["account_id"].(float64)) != aliceID {
				continue
			}
			if rec["scope"] != "mc" || rec["reason"] != "Griefing Spawn" && rec["reason"] != "Griefing spawn" {
				continue
			}
			if got, _ := rec["duration_minutes"].(float64); int(got) != 90 {
				t.Fatalf("stored duration = %v, want 90", rec["duration_minutes"])
			}
			if rec["expires_at"] == nil {
				t.Fatalf("stored expires_at missing: %v", rec)
			}
			found = true
		}
		if !found {
			t.Fatalf("mc ban record with reason/duration not found in %v", bans["bans"])
		}

		// A later platform ban must still block login even though an mc-scope
		// record for the same target was inserted first.
		code, body = adminBan(t, h, aliceID, map[string]any{
			"duration_minutes": 0,
			"reason":           "platform after mc",
			"scope":            "platform",
			"targets":          []string{"account"},
			"lookback_days":    30,
		})
		if code != http.StatusOK {
			t.Fatalf("platform ban after mc = %d, body = %v", code, body)
		}
		expectLogin("alice", "alice-dev-1", "10.0.0.1", http.StatusForbidden)
		adminUnbanAll(t, h, aliceID)
		expectLogin("alice", "alice-dev-1", "10.0.0.1", http.StatusOK)
	})
}
