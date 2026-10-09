package api

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"minecraftserverluncher/server/skins"
)

func newSkinUploadRequest(t testing.TB, png []byte) *http.Request {
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
	req := httptest.NewRequest(http.MethodPost, "/api/v1/skins/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func validPNG64x64() []byte {
	b, err := base64.StdEncoding.DecodeString(png64x64B64)
	if err != nil {
		panic(err)
	}
	return b
}

func TestSkinHandlerLifecycle(t *testing.T) {
	h := &SkinHandler{Storage: skins.NewMemoryStorage()}
	png := validPNG64x64()

	req := newSkinUploadRequest(t, png)
	rec := httptest.NewRecorder()
	h.handleUpload(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("upload status = %d, want 200; body=%q", rec.Code, rec.Body.String())
	}

	var resp struct {
		Success   bool   `json:"success"`
		SkinID    string `json:"skin_id"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal upload response: %v", err)
	}
	if !resp.Success || resp.SkinID == "" || resp.AvatarURL == "" {
		t.Fatalf("unexpected upload response: %+v", resp)
	}
	skinID := resp.SkinID
	t.Logf("uploaded skinID=%s avatarURL=%s", skinID, resp.AvatarURL)

	func() {
		var err error
		raw, err := h.Storage.GetSkin(skinID)
		if err != nil {
			t.Fatalf("storage has no raw skin after upload: %v", err)
		}
		if !bytes.Equal(raw, png) {
			t.Fatalf("stored raw skin bytes mismatch: got %d bytes, want %d bytes", len(raw), len(png))
		}
		avatar, err := h.Storage.GetSkin(skinID + "-avatar")
		if err != nil {
			t.Fatalf("storage has no avatar after upload: %v", err)
		}
		if len(avatar) == 0 {
			t.Fatal("stored avatar is empty")
		}
		t.Log("storage ok")
	}()

	h.checkGetSkin(t, skinID, png)
	h.checkGetAvatar(t, skinID)
	h.checkDelete(t, skinID)
	// After delete, GET /api/v1/skins/{id} must return 404.
	h.checkDeletedSkin(t, skinID)

	// After delete, GET /api/v1/skins/{id}/avatar must return 404.
	h.checkDeletedAvatar(t, skinID)
}

// The handlers read the resource id from the router path value, so requests
// built by hand must carry it explicitly (production sets it through
// http.ServeMux's "GET /api/v1/skins/{id}" pattern).
func (h *SkinHandler) checkGetSkin(t *testing.T, skinID string, want []byte) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/skins/"+skinID, nil)
	req.SetPathValue("id", skinID)
	rec := httptest.NewRecorder()
	h.handleGetSkin(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get raw skin status = %d, want 200; body=%q", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("raw skin content-type = %q", rec.Header().Get("Content-Type"))
	}
	if !bytes.Equal(rec.Body.Bytes(), want) {
		t.Fatal("raw skin bytes mismatch")
	}
}

func (h *SkinHandler) checkGetAvatar(t *testing.T, skinID string) {
	t.Helper()
	req, rec := httptest.NewRequest(http.MethodGet, "/api/v1/skins/"+skinID+"/avatar", nil), httptest.NewRecorder()
	req.SetPathValue("id", skinID)
	h.handleGetAvatar(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get avatar status = %d, want 200; body=%q", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("avatar content-type = %q", rec.Header().Get("Content-Type"))
	}
	if len(rec.Body.Bytes()) == 0 {
		t.Fatal("avatar body is empty")
	}
}

func (h *SkinHandler) checkDelete(t *testing.T, skinID string) {
	t.Helper()
	req, rec := httptest.NewRequest(http.MethodDelete, "/api/v1/skins/"+skinID, nil), httptest.NewRecorder()
	req.SetPathValue("id", skinID)
	h.handleDeleteSkin(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete status = %d, want 200; body=%q", rec.Code, rec.Body.String())
	}
}

func (h *SkinHandler) checkDeletedSkin(t *testing.T, skinID string) {
	t.Helper()
	req, rec := httptest.NewRequest(http.MethodGet, "/api/v1/skins/"+skinID, nil), httptest.NewRecorder()
	req.SetPathValue("id", skinID)
	h.handleGetSkin(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("get after delete status = %d, want 404; body=%q", rec.Code, rec.Body.String())
	}
}

func (h *SkinHandler) checkDeletedAvatar(t *testing.T, skinID string) {
	t.Helper()
	req, rec := httptest.NewRequest(http.MethodGet, "/api/v1/skins/"+skinID+"/avatar", nil), httptest.NewRecorder()
	req.SetPathValue("id", skinID)
	h.handleGetAvatar(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("get avatar after delete status = %d, want 404; body=%q", rec.Code, rec.Body.String())
	}
}

func TestSkinHandlerDeleteNotFound(t *testing.T) {
	h := &SkinHandler{Storage: skins.NewMemoryStorage()}
	png := validPNG64x64()

	// Upload the skin, then delete it and verify subsequent GETs return 404.
	req := newSkinUploadRequest(t, png)
	rec := httptest.NewRecorder()
	h.handleUpload(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("upload status = %d, want 200; body=%q", rec.Code, rec.Body.String())
	}

	var resp struct {
		Success  bool   `json:"success"`
		SkinID   string `json:"skin_id"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal upload response: %v", err)
	}
	if !resp.Success || resp.SkinID == "" || resp.AvatarURL == "" {
		t.Fatalf("unexpected upload response: %+v", resp)
	}
	skinID := resp.SkinID
	t.Logf("uploaded skinID=%s avatarURL=%s", skinID, resp.AvatarURL)

	// No independent delete helper; run the delete inline.
	{
		req, rec := httptest.NewRequest(http.MethodDelete, "/api/v1/skins/"+skinID, nil), httptest.NewRecorder()
		req.SetPathValue("id", skinID)
		h.handleDeleteSkin(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("delete status = %d, want 200; body=%q", rec.Code, rec.Body.String())
		}
		t.Log("delete ok")
	}

	// After delete, GET /api/v1/skins/{id} must return 404.
	h.checkDeletedSkin(t, skinID)

	// After delete, GET /api/v1/skins/{id}/avatar must return 404.
	h.checkDeletedAvatar(t, skinID)
}


func TestSkinHandlerRejectsBadUploads(t *testing.T) {
	h := &SkinHandler{Storage: skins.NewMemoryStorage()}

	// Non-PNG upload should be rejected (stream still needs valid multipart).
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("skin", "notskin.png")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write([]byte("this is not a PNG")); err != nil {
		t.Fatalf("part.Write: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/skins/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	h.handleUpload(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("non-PNG upload status = %d, want 400; body=%q", rec.Code, rec.Body.String())
	}

	// Oversized upload should be rejected.
	big := make([]byte, skins.MaxSkinBytes+1)
	body = &bytes.Buffer{}
	writer = multipart.NewWriter(body)
	part, err = writer.CreateFormFile("skin", "big.png")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write(big); err != nil {
		t.Fatalf("part.Write: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close: %v", err)
	}
	req = httptest.NewRequest(http.MethodPost, "/api/v1/skins/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec = httptest.NewRecorder()
	h.handleUpload(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("oversized upload status = %d, want 400; body=%q", rec.Code, rec.Body.String())
	}

	// Too-small skin should be rejected.
	small := decodeB64(t, png32x32B64)
	body = &bytes.Buffer{}
	writer = multipart.NewWriter(body)
	part, err = writer.CreateFormFile("skin", "small.png")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write(small); err != nil {
		t.Fatalf("part.Write: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close: %v", err)
	}
	req = httptest.NewRequest(http.MethodPost, "/api/v1/skins/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec = httptest.NewRecorder()
	h.handleUpload(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad dimension upload status = %d, want 400; body=%q", rec.Code, rec.Body.String())
	}
}


func TestPterodactylWebhookHandler(t *testing.T) {
	t.Setenv("PTERODACTYL_WEBHOOK_SECRET", "test-secret")

	var received []byte
	wh := &PterodactylWebhook{
		Handler: func(body []byte) { received = append([]byte{}, body...) },
	}
	h := wh.Handle()

	payload := []byte(`{"event":"server.started","data":{"id":1}}`)
	mac := hmac.New(sha256.New, []byte("test-secret"))
	mac.Write(payload)
	signature := fmt.Sprintf("%x", mac.Sum(nil))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/pterodactyl", bytes.NewReader(payload))
	req.Header.Set("X-Pterodactyl-Signature", signature)
	req.Header.Set("X-Pterodactyl-Event", "server.started")
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("valid signature status = %d, want 200; body=%q", rec.Code, rec.Body.String())
	}
	if !bytes.Equal(received, payload) {
		t.Fatal("payload mismatch")
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/pterodactyl", bytes.NewReader(payload))
	req.Header.Set("X-Pterodactyl-Signature", "deadbeef")
	rec = httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("invalid signature status = %d, want 401", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/pterodactyl", bytes.NewReader([]byte(`{}`)))
	rec = httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("missing signature status = %d, want 401", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/webhooks/pterodactyl", nil)
	rec = httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("get webhook status = %d, want 405", rec.Code)
	}
}
