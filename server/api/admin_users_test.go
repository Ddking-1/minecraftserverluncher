package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdminUserListAndDelete(t *testing.T) {
	srv := newAdminTestServer(t, "secret-admin", "")
	defer srv.Close()
	h := srv.Handler()

	if code, _ := register(t, h, "adminlistuser1", "supersecret"); code != http.StatusCreated {
		t.Fatalf("register user1 status = %d", code)
	}
	if code, _ := register(t, h, "adminlistuser2", "supersecret"); code != http.StatusCreated {
		t.Fatalf("register user2 status = %d", code)
	}

	// Listing requires the admin token.
	if code := doAdminJSON(t, h, http.MethodGet, "/api/admin/users", nil, "secret-admin"); code != http.StatusOK {
		t.Fatalf("list with admin token status = %d, want 200", code)
	}
	if code := doAdminJSON(t, h, http.MethodGet, "/api/admin/users", nil, "nope"); code != http.StatusUnauthorized {
		t.Fatalf("list with wrong token status = %d, want 401", code)
	}

	// Read the body to find the target user's id.
	req := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	req.Header.Set("X-Admin-Token", "secret-admin")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var response struct {
		Users []map[string]any `json:"users"`
		Total int              `json:"total"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode users: %v", err)
	}
	if response.Total < 2 {
		t.Fatalf("total = %d, want >= 2", response.Total)
	}

	userID := int64(0)
	for _, u := range response.Users {
		if name, ok := u["username"].(string); ok && name == "adminlistuser1" {
			if idf, ok := u["id"].(float64); ok {
				userID = int64(idf)
			}
			break
		}
	}
	if userID == 0 {
		t.Fatal("could not find adminlistuser1 in the list")
	}

	// Delete the user, then deleting again is 404.
	if code := doAdminJSON(t, h, http.MethodDelete, fmt.Sprintf("/api/admin/users/%d", userID), nil, "secret-admin"); code != http.StatusOK {
		t.Fatalf("delete status = %d, want 200", code)
	}
	if code := doAdminJSON(t, h, http.MethodDelete, fmt.Sprintf("/api/admin/users/%d", userID), nil, "secret-admin"); code != http.StatusNotFound {
		t.Fatalf("second delete status = %d, want 404", code)
	}
}

func TestAdminUserListSearch(t *testing.T) {
	srv := newAdminTestServer(t, "secret-admin", "")
	defer srv.Close()
	h := srv.Handler()

	register(t, h, "searchalpha", "supersecret")
	register(t, h, "searchbeta", "supersecret")

	req := httptest.NewRequest(http.MethodGet, "/api/admin/users?q=alpha", nil)
	req.Header.Set("X-Admin-Token", "secret-admin")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("search status = %d, want 200", rec.Code)
	}
	var response struct {
		Users []map[string]any `json:"users"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(response.Users) != 1 {
		t.Fatalf("search returned %d users, want 1", len(response.Users))
	}
	if response.Users[0]["username"] != "searchalpha" {
		t.Fatalf("search returned %v", response.Users[0]["username"])
	}
}
