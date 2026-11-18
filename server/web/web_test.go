package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIndexServesSPA(t *testing.T) {
	rec := request(t, http.MethodGet, "/")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{"Alpha Network", `id="app"`, "/assets/js/main.js", "/assets/css/tokens.css"} {
		if !strings.Contains(body, want) {
			t.Errorf("index page must contain %q", want)
		}
	}
	if strings.Contains(body, "Skin Manager") {
		t.Error("index page must not contain the legacy Skin Manager title")
	}
	if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", got)
	}
}

func TestAppRoutesServeIndex(t *testing.T) {
	for _, route := range []string{"/login", "/register", "/dashboard", "/players", "/admin"} {
		t.Run(route, func(t *testing.T) {
			rec := request(t, http.MethodGet, route)
			if rec.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want 200", route, rec.Code)
			}
			if !strings.Contains(rec.Body.String(), `id="app"`) {
				t.Fatalf("GET %s must return the SPA shell", route)
			}
		})
	}
}

func TestUnknownExtensionlessRouteFallsBackToSPA(t *testing.T) {
	rec := request(t, http.MethodGet, "/this/route/does/not/exist")
	if rec.Code != http.StatusOK {
		t.Fatalf("unknown SPA route status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `id="app"`) {
		t.Fatal("unknown SPA route must return index.html")
	}
}

func TestEmbeddedAssetsAreServed(t *testing.T) {
	cases := map[string]string{
		"/assets/js/main.js":       "import",
		"/assets/css/tokens.css":   "--accent",
		"/assets/js/views/admin.js": "permission",
	}
	for route, want := range cases {
		rec := request(t, http.MethodGet, route)
		if rec.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want 200", route, rec.Code)
			continue
		}
		if !strings.Contains(rec.Body.String(), want) {
			t.Errorf("GET %s does not contain %q", route, want)
		}
	}
}

func TestMissingAssetNotFound(t *testing.T) {
	for _, route := range []string{"/nope.html", "/assets/css/missing.css", "/assets/"} {
		rec := request(t, http.MethodGet, route)
		if rec.Code != http.StatusNotFound {
			t.Errorf("GET %s status = %d, want 404", route, rec.Code)
		}
	}
}

func TestLegacyPagesRedirect(t *testing.T) {
	cases := map[string]string{
		"/skin-manager.html": "/dashboard",
		"/dashboard.html":    "/dashboard",
		"/admin.html":        "/admin",
	}
	for route, target := range cases {
		rec := request(t, http.MethodGet, route)
		if rec.Code != http.StatusMovedPermanently {
			t.Errorf("GET %s status = %d, want 301", route, rec.Code)
		}
		if got := rec.Header().Get("Location"); got != target {
			t.Errorf("GET %s Location = %q, want %q", route, got, target)
		}
	}
}

func TestIndexRejectsNonReadMethods(t *testing.T) {
	rec := request(t, http.MethodPost, "/login")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /login status = %d, want 405", rec.Code)
	}
	if got := rec.Header().Get("Allow"); got != "GET, HEAD" {
		t.Errorf("Allow = %q, want GET, HEAD", got)
	}
}

func TestHeadReturnsHeadersWithoutBody(t *testing.T) {
	rec := request(t, http.MethodHead, "/dashboard")
	if rec.Code != http.StatusOK {
		t.Fatalf("HEAD /dashboard status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", got)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("HEAD body length = %d, want 0", rec.Body.Len())
	}
}

func request(t *testing.T, method, target string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	Handler().ServeHTTP(rec, httptest.NewRequest(method, target, nil))
	return rec
}
