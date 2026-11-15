package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"minecraftserverluncher/server/skins"
)

func newSkinFileServer(t *testing.T) (*Server, http.Handler) {
	t.Helper()
	srv := newTestServer(t)
	srv.SkinHandler = &SkinHandler{
		Storage: skins.NewMemoryStorage(),
		History: skins.NewThreadSafeSkinHistory(),
	}
	return srv, srv.Handler()
}

func multipartSkinRequest(t *testing.T, png []byte, token string) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("skin", "skin.png")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write(png); err != nil {
		t.Fatalf("part.Write: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/skin-files", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}

// The endpoints used by the launcher must be reachable through the real router
// and must not be served without authentication.
func TestSkinFileRoutesRequireAuthentication(t *testing.T) {
	_, h := newSkinFileServer(t)
	png, err := base64.StdEncoding.DecodeString(png64x64B64)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, multipartSkinRequest(t, png, ""))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated upload status = %d, want 401", rec.Code)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/skin-files/some-id/avatar", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated avatar status = %d, want 401", rec.Code)
	}
}

func TestSkinFileUploadRoundTrip(t *testing.T) {
	_, h := newSkinFileServer(t)
	_, out := register(t, h, "skinfiles", "supersecret")
	access := out["access_token"].(string)

	png, err := base64.StdEncoding.DecodeString(png64x64B64)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, multipartSkinRequest(t, png, access))
	if rec.Code != http.StatusOK {
		t.Fatalf("upload status = %d, want 200; body=%q", rec.Code, rec.Body.String())
	}
	var uploaded struct {
		Success   bool   `json:"success"`
		SkinID    string `json:"skin_id"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &uploaded); err != nil {
		t.Fatalf("decode upload response: %v", err)
	}
	if !uploaded.Success || uploaded.SkinID == "" {
		t.Fatalf("unexpected upload response: %+v", uploaded)
	}
	if want := "/api/v1/skin-files/" + uploaded.SkinID + "/avatar"; uploaded.AvatarURL != want {
		t.Fatalf("avatar_url = %q, want %q", uploaded.AvatarURL, want)
	}

	avatarReq := httptest.NewRequest(http.MethodGet, uploaded.AvatarURL, nil)
	avatarReq.Header.Set("Authorization", "Bearer "+access)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, avatarReq)
	if rec.Code != http.StatusOK {
		t.Fatalf("avatar status = %d, want 200; body=%q", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "image/png" {
		t.Fatalf("avatar content-type = %q, want image/png", ct)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/skin-files/"+uploaded.SkinID, nil)
	deleteReq.Header.Set("Authorization", "Bearer "+access)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, deleteReq)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete status = %d, want 200; body=%q", rec.Code, rec.Body.String())
	}

	goneReq := httptest.NewRequest(http.MethodGet, "/api/v1/skin-files/"+uploaded.SkinID, nil)
	goneReq.Header.Set("Authorization", "Bearer "+access)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, goneReq)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("get after delete status = %d, want 404", rec.Code)
	}
}

func TestSkinFileDeleteRejectsOtherAccount(t *testing.T) {
	_, h := newSkinFileServer(t)
	_, ownerOut := register(t, h, "skinowner", "supersecret")
	owner := ownerOut["access_token"].(string)
	_, otherOut := register(t, h, "skinother", "supersecret")
	other := otherOut["access_token"].(string)

	png, err := base64.StdEncoding.DecodeString(png64x64B64)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, multipartSkinRequest(t, png, owner))
	if rec.Code != http.StatusOK {
		t.Fatalf("upload status = %d, want 200", rec.Code)
	}
	var uploaded struct {
		SkinID string `json:"skin_id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &uploaded); err != nil {
		t.Fatal(err)
	}

	del := httptest.NewRequest(http.MethodDelete, "/api/v1/skin-files/"+uploaded.SkinID, nil)
	del.Header.Set("Authorization", "Bearer "+other)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, del)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("other account delete = %d, want 404", rec.Code)
	}

	get := httptest.NewRequest(http.MethodGet, "/api/v1/skin-files/"+uploaded.SkinID, nil)
	get.Header.Set("Authorization", "Bearer "+owner)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, get)
	if rec.Code != http.StatusOK {
		t.Fatalf("owner get after rejected delete = %d, want 200", rec.Code)
	}
}

func TestPterodactylWebhookRouteIsMounted(t *testing.T) {
	srv := newTestServer(t)
	srv.PterodactylWebhook = &PterodactylWebhook{Handler: func([]byte) {}}
	h := srv.Handler()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/pterodactyl", bytes.NewReader([]byte(`{}`)))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	// Without a configured secret the handler must fail closed with 401, not 404.
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("webhook route status = %d, want 401 (route must be mounted)", rec.Code)
	}
}

func TestServerStartRouteRequiresAdminToken(t *testing.T) {
	srv := newTestServer(t)
	srv.cfg.AdminToken = "admin-secret"
	h := srv.Handler()

	req := httptest.NewRequest(http.MethodPost, "/api/server/start", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("server start without admin token = %d, want 401 (route must be mounted)", rec.Code)
	}
}
