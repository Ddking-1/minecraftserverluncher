package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestHoneypotRejectsRegister(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	code, _ := doJSON(t, h, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"username": "honeypotuser", "password": "supersecret", "device_id": "dev-hp",
		"website": "http://spam.example",
	}, "")
	if code != http.StatusBadRequest {
		t.Fatalf("honeypot register status = %d, want 400", code)
	}
}

func TestHoneypotRejectsLogin(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	register(t, h, "hploginuser", "supersecret")
	code, _ := doJSON(t, h, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "hploginuser", "password": "supersecret", "device_id": "dev-hp",
		"website": "x",
	}, "")
	if code != http.StatusBadRequest {
		t.Fatalf("honeypot login status = %d, want 400", code)
	}
}

func TestAdminSkinApprovalEndpoint(t *testing.T) {
	srv := newAdminTestServer(t, "secret-admin", "")
	defer srv.Close()
	h := srv.Handler()

	code := doAdminJSON(t, h, http.MethodPost, "/api/admin/skins/skin-123/approve", map[string]any{"approved": false}, "secret-admin")
	if code != http.StatusOK {
		t.Fatalf("skin approve status = %d, want 200", code)
	}
	code = doAdminJSON(t, h, http.MethodPost, "/api/admin/skins/skin-123/approve", map[string]any{"approved": true}, "secret-admin")
	if code != http.StatusOK {
		t.Fatalf("skin approve status = %d, want 200", code)
	}
	// Wrong admin token rejected.
	code = doAdminJSON(t, h, http.MethodPost, "/api/admin/skins/skin-123/approve", map[string]any{"approved": true}, "nope")
	if code != http.StatusUnauthorized {
		t.Fatalf("skin approve wrong token status = %d, want 401", code)
	}
}

func TestAccountApprovalGate(t *testing.T) {
	srv := newAdminTestServer(t, "secret-admin", "")
	defer srv.Close()
	h := srv.Handler()

	if code, _ := register(t, h, "pendinguser", "supersecret"); code != http.StatusCreated {
		t.Fatalf("register status = %d", code)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	req.Header.Set("X-Admin-Token", "secret-admin")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var response struct {
		Users []map[string]any `json:"users"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	userID := int64(0)
	for _, u := range response.Users {
		if name, _ := u["username"].(string); name == "pendinguser" {
			if idf, ok := u["id"].(float64); ok {
				userID = int64(idf)
			}
		}
	}
	if userID == 0 {
		t.Fatal("could not find pendinguser")
	}

	code := doAdminJSON(t, h, http.MethodPost, "/api/admin/users/"+strconv.FormatInt(userID, 10)+"/approve", map[string]any{"approved": false}, "secret-admin")
	if code != http.StatusOK {
		t.Fatalf("approve revoke status = %d, want 200", code)
	}

	code, _ = doJSON(t, h, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "pendinguser", "password": "supersecret", "device_id": "dev-pendinguser",
	}, "")
	if code != http.StatusForbidden {
		t.Fatalf("login while unapproved status = %d, want 403", code)
	}

	code = doAdminJSON(t, h, http.MethodPost, "/api/admin/users/"+strconv.FormatInt(userID, 10)+"/approve", map[string]any{"approved": true}, "secret-admin")
	if code != http.StatusOK {
		t.Fatalf("approve status = %d, want 200", code)
	}
	code, _ = doJSON(t, h, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "pendinguser", "password": "supersecret", "device_id": "dev-pendinguser",
	}, "")
	if code != http.StatusOK {
		t.Fatalf("login after approval status = %d, want 200", code)
	}
}
