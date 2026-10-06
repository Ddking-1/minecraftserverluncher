package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testSigningKey = "unit-test-signing-key"

func newUpdateTestServer(t *testing.T) *Server {
	t.Helper()
	srv := newTestServer(t)
	srv.cfg.LauncherVersion = "1.2.3"
	srv.cfg.UpdateBaseURL = "https://cdn.example.com"
	srv.cfg.UpdateSHA256 = "aaaabbbbccccddddeeeeffff0000111122223333444455556666777788889999"
	srv.cfg.UpdateMandatory = true
	srv.cfg.UpdateSigningKey = testSigningKey
	return srv
}

func TestUpdatesManifest(t *testing.T) {
	srv := newUpdateTestServer(t)
	h := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/updates/manifest", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("manifest status = %d, want 200", rec.Code)
	}
	var body struct {
		Update struct {
			Version   string `json:"version"`
			URL       string `json:"url"`
			SHA256    string `json:"sha256"`
			Mandatory bool   `json:"mandatory"`
			Signature string `json:"signature"`
		} `json:"update"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Update.Version != "1.2.3" {
		t.Fatalf("version = %q, want 1.2.3", body.Update.Version)
	}
	if body.Update.URL != "https://cdn.example.com/MinecraftServerLauncher-1.2.3-setup.exe" {
		t.Fatalf("url = %q", body.Update.URL)
	}
	if body.Update.SHA256 != "aaaabbbbccccddddeeeeffff0000111122223333444455556666777788889999" {
		t.Fatalf("sha256 = %q", body.Update.SHA256)
	}
	if !body.Update.Mandatory {
		t.Fatal("mandatory should be true")
	}
	// The signature must verify against the returned fields.
	ok := VerifyManifestSignature(testSigningKey, body.Update.Version, body.Update.URL,
		body.Update.SHA256, body.Update.Mandatory, body.Update.Signature)
	if !ok {
		t.Fatal("manifest signature did not verify")
	}
}

func TestUpdatesManifestSignatureTampered(t *testing.T) {
	srv := newUpdateTestServer(t)
	h := srv.Handler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/updates/manifest", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var body struct {
		Update struct {
			Version   string `json:"version"`
			URL       string `json:"url"`
			SHA256    string `json:"sha256"`
			Mandatory bool   `json:"mandatory"`
			Signature string `json:"signature"`
		} `json:"update"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)

	// A tampered URL must not verify against the unchanged signature.
	ok := VerifyManifestSignature(testSigningKey, body.Update.Version,
		"https://evil.example.com/x.exe", body.Update.SHA256, body.Update.Mandatory, body.Update.Signature)
	if ok {
		t.Fatal("tampered URL verified against original signature")
	}
	// Wrong key must not verify.
	ok = VerifyManifestSignature("wrong-key", body.Update.Version, body.Update.URL,
		body.Update.SHA256, body.Update.Mandatory, body.Update.Signature)
	if ok {
		t.Fatal("wrong key verified")
	}
}

func TestUpdatesManifestNoAuth(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Handler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/updates/manifest", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("manifest without auth status = %d, want 200", rec.Code)
	}
}
