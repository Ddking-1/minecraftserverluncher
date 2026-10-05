package api

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Real, minimal PNG fixtures (generated once, base64-encoded).
const (
	png64x64B64  = "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAYAAACqaXHeAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsMAAA7DAcdvqGQAAAAnSURBVHhe7cEBDQAAAMKg909tDjegAAAAAAAAAAAAAAAAAAAAgHcDQEAAAVkKewwAAAAASUVORK5CYII="
	png64x32B64  = "iVBORw0KGgoAAAANSUhEUgAAAEAAAAAgCAYAAACinX6EAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsMAAA7DAcdvqGQAAAAeSURBVGhD7cEBDQAAAMKg909tDwcUAAAAAAAAAPBuICAAAWbYvZgAAAAASUVORK5CYII="
	png32x32B64  = "iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAYAAABzenr0AAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsMAAA7DAcdvqGQAAAAaSURBVFhH7cEBAQAAAIIg/69uSEABAAAA7wYQIAABwWAsnwAAAABJRU5ErkJggg=="
)

func decodeB64(t *testing.T, s string) []byte {
	t.Helper()
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	return b
}

func uploadSkin(t *testing.T, h http.Handler, access, variant, contentType string, body []byte) int {
	t.Helper()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/skins/"+variant, io.NopCloser(bytes.NewReader(body)))
	req.Header.Set("Authorization", "Bearer "+access)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Code
}

func TestSkinLifecycle(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Handler()

	_, out := register(t, h, "skinuser", "supersecret")
	access := out["access_token"].(string)
	png := decodeB64(t, png64x64B64)

	if code := uploadSkin(t, h, access, "default", "image/png", png); code != http.StatusOK {
		t.Fatalf("upload status = %d, want 200", code)
	}

	code, body := doJSON(t, h, http.MethodGet, "/api/v1/skins", nil, access)
	if code != http.StatusOK {
		t.Fatalf("list status = %d", code)
	}
	skins, ok := body["skins"].([]any)
	if !ok || len(skins) != 1 {
		t.Fatalf("skins = %v, want 1 entry", body["skins"])
	}

	// Public fetch returns the stored bytes and a sniffed content type.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/skins/skinuser/default", nil)
	req.Header.Set("Authorization", "Bearer "+access)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("fetch status = %d", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("content-type = %q", rec.Header().Get("Content-Type"))
	}
	if !bytes.Equal(rec.Body.Bytes(), png) {
		t.Fatal("fetched skin data does not match upload")
	}

	deleteCode, _ := doJSON(t, h, http.MethodDelete, "/api/v1/skins/default", nil, access)
	if deleteCode != http.StatusOK {
		t.Fatalf("delete status = %d", deleteCode)
	}
}

func TestSkinUploadRequiresAuth(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Handler()
	body := decodeB64(t, png64x64B64)
	if code := uploadSkin(t, h, "", "default", "image/png", body); code != http.StatusUnauthorized {
		t.Fatalf("upload without token status = %d, want 401", code)
	}
}

func TestSkinRejectsNonImage(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Handler()
	_, out := register(t, h, "skinuser2", "supersecret")
	access := out["access_token"].(string)

	// A PHP payload declared as image/png must be rejected by magic-byte sniffing (T-C).
	php := []byte("<?php system($_GET['c']); ?>")
	if code := uploadSkin(t, h, access, "default", "image/png", php); code != http.StatusUnsupportedMediaType {
		t.Fatalf("php-as-png status = %d, want 415", code)
	}
}

func TestSkinRejectsBadDimensions(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Handler()
	_, out := register(t, h, "skinuser3", "supersecret")
	access := out["access_token"].(string)

	// 32x32 is a real PNG but not a valid Minecraft skin (T-B).
	if code := uploadSkin(t, h, access, "default", "image/png", decodeB64(t, png32x32B64)); code != http.StatusBadRequest {
		t.Fatalf("32x32 skin status = %d, want 400", code)
	}
	// 64x32 legacy is valid.
	if code := uploadSkin(t, h, access, "legacy", "image/png", decodeB64(t, png64x32B64)); code != http.StatusOK {
		t.Fatalf("64x32 skin status = %d, want 200", code)
	}
}

func TestSkinUploadRateLimit(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Handler()
	_, out := register(t, h, "skinlimit", "supersecret")
	access := out["access_token"].(string)
	png := decodeB64(t, png64x64B64)

	// The first two uploads succeed (the limit allows 2 per 3 hours).
	for i := 0; i < 2; i++ {
		if code := uploadSkin(t, h, access, "default", "image/png", png); code != http.StatusOK {
			t.Fatalf("upload %d status = %d, want 200", i+1, code)
		}
	}
	// The third is rejected with 429 (T-A).
	if code := uploadSkin(t, h, access, "default", "image/png", png); code != http.StatusTooManyRequests {
		t.Fatalf("3rd upload status = %d, want 429", code)
	}
}

func TestSkinRejectsForeignDelete(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Handler()
	_, outA := register(t, h, "skinuserA", "supersecret")
	accessA := outA["access_token"].(string)
	_, outB := register(t, h, "skinuserB", "supersecret")
	accessB := outB["access_token"].(string)

	if code := uploadSkin(t, h, accessA, "default", "image/png", decodeB64(t, png64x64B64)); code != http.StatusOK {
		t.Fatalf("upload A status = %d", code)
	}
	// B deleting "default" only affects B's own (nonexistent) variant.
	if code, _ := doJSON(t, h, http.MethodDelete, "/api/v1/skins/default", nil, accessB); code != http.StatusOK {
		t.Fatalf("B delete status = %d", code)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/skins/skinuserA/default", nil)
	req.Header.Set("Authorization", "Bearer "+accessA)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("A skin after B delete status = %d, want 200", rec.Code)
	}
}
