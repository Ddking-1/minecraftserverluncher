package api

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSkinLifecycle(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Handler()

	_, out := register(t, h, "skinuser", "supersecret")
	access := out["access_token"].(string)
	png := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

	// Upload a skin.
	req := newJSONRequest(t, http.MethodPut, "/api/v1/skins/default", nil)
	req.Header.Set("Authorization", "Bearer "+access)
	req.Header.Set("Content-Type", "image/png")
	req.Body = io.NopCloser(bytes.NewReader(png))
	rec := httptestRecorder(req, h)
	if rec.Code != http.StatusOK {
		t.Fatalf("upload status = %d, body = %v", rec.Code, rec.Body.String())
	}

	// List own skins contains the variant.
	code, body := doJSON(t, h, http.MethodGet, "/api/v1/skins", nil, access)
	if code != http.StatusOK {
		t.Fatalf("list status = %d", code)
	}
	skins, ok := body["skins"].([]any)
	if !ok || len(skins) != 1 {
		t.Fatalf("skins = %v, want 1 entry", body["skins"])
	}

	// Fetch the skin publicly.
	req = newJSONRequest(t, http.MethodGet, "/api/v1/skins/skinuser/default", nil)
	req.Header.Set("Authorization", "Bearer "+access)
	rec = httptestRecorder(req, h)
	if rec.Code != http.StatusOK {
		t.Fatalf("fetch status = %d", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("content-type = %q", rec.Header().Get("Content-Type"))
	}
	if !bytes.Equal(rec.Body.Bytes(), png) {
		t.Fatal("fetched skin data does not match upload")
	}

	// Delete and confirm gone.
	code, _ = doJSON(t, h, http.MethodDelete, "/api/v1/skins/default", nil, access)
	if code != http.StatusOK {
		t.Fatalf("delete status = %d", code)
	}
	req = newJSONRequest(t, http.MethodGet, "/api/v1/skins/skinuser/default", nil)
	req.Header.Set("Authorization", "Bearer "+access)
	rec = httptestRecorder(req, h)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("fetch after delete status = %d, want 404", rec.Code)
	}
}

func TestSkinUploadRequiresAuth(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Handler()

	req := newJSONRequest(t, http.MethodPut, "/api/v1/skins/default", nil)
	req.Header.Set("Content-Type", "image/png")
	req.Body = io.NopCloser(bytes.NewReader([]byte{1, 2, 3}))
	if code := httptestRecorder(req, h).Code; code != http.StatusUnauthorized {
		t.Fatalf("upload without token status = %d, want 401", code)
	}
}

func TestSkinUploadRejectsOversizeAndBadType(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Handler()
	_, out := register(t, h, "skinuser2", "supersecret")
	access := out["access_token"].(string)

	// Bad content type.
	req := newJSONRequest(t, http.MethodPut, "/api/v1/skins/default", nil)
	req.Header.Set("Authorization", "Bearer "+access)
	req.Header.Set("Content-Type", "text/plain")
	req.Body = io.NopCloser(bytes.NewReader([]byte("x")))
	if code := httptestRecorder(req, h).Code; code != http.StatusUnsupportedMediaType {
		t.Fatalf("bad content-type status = %d, want 415", code)
	}

	// Oversize body.
	big := bytes.Repeat([]byte{0}, 2<<20)
	req = newJSONRequest(t, http.MethodPut, "/api/v1/skins/default", nil)
	req.Header.Set("Authorization", "Bearer "+access)
	req.Header.Set("Content-Type", "image/png")
	req.Body = io.NopCloser(bytes.NewReader(big))
	if code := httptestRecorder(req, h).Code; code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversize status = %d, want 413", code)
	}
}

func TestSkinRejectsForeignDelete(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Handler()
	_, outA := register(t, h, "skinuserA", "supersecret")
	accessA := outA["access_token"].(string)
	_, outB := register(t, h, "skinuserB", "supersecret")
	accessB := outB["access_token"].(string)

	// A uploads a skin.
	req := newJSONRequest(t, http.MethodPut, "/api/v1/skins/default", nil)
	req.Header.Set("Authorization", "Bearer "+accessA)
	req.Header.Set("Content-Type", "image/png")
	req.Body = io.NopCloser(bytes.NewReader([]byte{0x89, 0x50, 0x4E, 0x47}))
	if code := httptestRecorder(req, h).Code; code != http.StatusOK {
		t.Fatalf("upload A status = %d", code)
	}

	// B cannot delete A's skin (B only deletes its own variants).
	code, _ := doJSON(t, h, http.MethodDelete, "/api/v1/skins/default", nil, accessB)
	if code != http.StatusOK {
		t.Fatalf("B delete status = %d", code)
	}
	// A's skin must still exist.
	req = newJSONRequest(t, http.MethodGet, "/api/v1/skins/skinuserA/default", nil)
	req.Header.Set("Authorization", "Bearer "+accessA)
	if rec := httptestRecorder(req, h); rec.Code != http.StatusOK {
		t.Fatalf("A skin after B delete status = %d, want 200", rec.Code)
	}
}

// helpers reused by skins tests
func httptestRecorder(req *http.Request, h http.Handler) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

