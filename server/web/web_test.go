package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIndexServesSkinManager(t *testing.T) {
	h := Handler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Skin Manager") {
		t.Fatal("index page must contain the Skin Manager title")
	}
	if !strings.Contains(body, "/api/v1/skins") {
		t.Fatal("index page must reference the skins API")
	}
}

func TestSkinManagerDirectPath(t *testing.T) {
	h := Handler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/skin-manager.html", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /skin-manager.html status = %d, want 200", rec.Code)
	}
}

func TestMissingAssetNotFound(t *testing.T) {
	h := Handler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/nope.html", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET /nope.html status = %d, want 404", rec.Code)
	}
}

func TestDashboardServed(t *testing.T) {
	h := Handler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/dashboard", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /dashboard status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Vezérlőpult") && !strings.Contains(body, "Dashboard") {
		t.Fatal("dashboard must contain its title")
	}
	// A dashboard.html nem tartalmazza közvetlenül az API URL-t, hanem csak az app.js hívja meg
	if !strings.Contains(body, "dashboardContent") {
		t.Fatal("dashboard must have dashboard content div")
	}
}

func TestAdminPageServed(t *testing.T) {
	h := Handler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/admin", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /admin status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Admin Panell") && !strings.Contains(body, "Admin Panel") {
		t.Fatal("admin page must contain its title")
	}
	// Az admin.html nem tartalmazza közvetlenül az API URL-t, hanem csak az app.js hívja meg
	if !strings.Contains(body, "userRows") {
		t.Fatal("admin page must have user rows table")
	}
}
