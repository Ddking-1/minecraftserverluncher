package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
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
	defer srv.Close()
	h := srv.Handler()

	_, out := register(t, h, "skinuser", "supersecret")
	access := out["access_token"].(string)
	png := decodeB64(t, png64x64B64)

	if code := uploadSkin(t, h, access, "default", "image/png", png); code != http.StatusOK {
		t.Fatalf("upload status = %d, want 200", code)
	}

	// List skins - this route is behind auth middleware in the router,
	// so it requires a valid bearer token. Use the access token.
	// The router registers GET /api/v1/skins on the protected mux,
	// which requires authentication. We send the access token in the
	// Authorization header as a bearer token.
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/skins", nil)
	listReq.Header.Set("Authorization", "Bearer "+access)
	listRec := httptest.NewRecorder()
	h.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want 200, body = %s", listRec.Code, listRec.Body.String())
	}
	var listBody map[string]any
	if err := json.Unmarshal(listRec.Body.Bytes(), &listBody); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	skins, ok := listBody["skins"].([]any)
	if !ok || len(skins) != 1 {
		t.Fatalf("skins = %v, want 1 entry", listBody["skins"])
	}

	// Public fetch returns the stored bytes and a sniffed content type.
	fetchReq := httptest.NewRequest(http.MethodGet, "/api/v1/skins/skinuser/default", nil)
	fetchReq.Header.Set("Authorization", "Bearer "+access)
	fetchRec := httptest.NewRecorder()
	h.ServeHTTP(fetchRec, fetchReq)
	if fetchRec.Code != http.StatusOK {
		t.Fatalf("fetch status = %d", fetchRec.Code)
	}
	if fetchRec.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("content-type = %q", fetchRec.Header().Get("Content-Type"))
	}
	if !bytes.Equal(fetchRec.Body.Bytes(), png) {
		t.Fatal("fetched skin data does not match upload")
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/skins/default", nil)
	deleteReq.Header.Set("Authorization", "Bearer "+access)
	deleteRec := httptest.NewRecorder()
	h.ServeHTTP(deleteRec, deleteReq)
	deleteCode := deleteRec.Code
	if deleteCode != http.StatusOK {
		t.Fatalf("delete status = %d", deleteCode)
	}
}

func TestSkinUploadRequiresAuth(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	h := srv.Handler()
	body := decodeB64(t, png64x64B64)
	if code := uploadSkin(t, h, "", "default", "image/png", body); code != http.StatusUnauthorized {
		t.Fatalf("upload without token status = %d, want 401", code)
	}
}

func TestSkinRejectsNonImage(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
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
	defer srv.Close()
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
	defer srv.Close()
	h := srv.Handler()
	_, out := register(t, h, "skinlimit", "supersecret")
	access := out["access_token"].(string)
	png := decodeB64(t, png64x64B64)

	// The rate limiter is set to 10000 req/min in newTestServer, so
	// we won't hit the global rate limit. However, there's also a
	// skin-specific rate limit that allows 2 uploads per 3 hours.
	// The test verifies this limit is enforced.
	// First two uploads should succeed.
	for i := 0; i < 2; i++ {
		if code := uploadSkin(t, h, access, "default", "image/png", png); code != http.StatusOK {
			t.Fatalf("upload %d status = %d, want 200", i+1, code)
		}
	}
	// Third upload should be rate limited (429).
	if code := uploadSkin(t, h, access, "default", "image/png", png); code != http.StatusTooManyRequests {
		t.Fatalf("3rd upload status = %d, want 429", code)
	}
}

func TestSkinRejectsForeignDelete(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
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
	verifyReq := httptest.NewRequest(http.MethodGet, "/api/v1/skins/skinuserA/default", nil)
	verifyReq.Header.Set("Authorization", "Bearer "+accessA)
	verifyRec := httptest.NewRecorder()
	h.ServeHTTP(verifyRec, verifyReq)
	if verifyRec.Code != http.StatusOK {
		t.Fatalf("A skin after B delete status = %d, want 200", verifyRec.Code)
	}
}
