package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"minecraftserverluncher/server/config"
	"minecraftserverluncher/server/database"
	"minecraftserverluncher/server/sessions"
)

func TestAccessTokenRejectsMalformedBearer(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/account/me", nil)
	req.Header.Set("Authorization", "Bearer not-base64!!")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestRefreshRejectsMalformedRefreshToken(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	code, body := doJSON(t, h, http.MethodPost, "/api/v1/auth/refresh", map[string]any{
		"refresh_token": "not-valid-base64!!",
	}, "")

	if code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401, body = %v", code, body)
	}
}

func TestRefreshRejectsMissingRefreshToken(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	code, body := doJSON(t, h, http.MethodPost, "/api/v1/auth/refresh", map[string]any{
		"refresh_token": "",
	}, "")

	if code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400, body = %v", code, body)
	}
}

func TestRefreshRejectsExpiredRefreshToken(t *testing.T) {
	db, err := database.Open(t.TempDir() + "/edge.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	cfg := &config.Config{
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: -100 * time.Hour, // already expired
		BcryptCost:      10,
	}
	srv := NewServer(db, cfg)
	defer srv.Close()
	h := srv.Handler()
	var buf bytes.Buffer
	_ = json.NewEncoder(&buf).Encode(map[string]any{"refresh_token": ""})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", &buf)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 for expired refresh token, body = %s", rec.Code, rec.Body.String())
	}
}

func TestInvalidJSONBodyReturnsBadRequest(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	// An unreadable body provokes json.Decoder.ErrUnexpectedEOF, which we treat
	// as a bad request rather than a 500.
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Body = &invalidJSONReader{}
	req.ContentLength = -1
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

type invalidJSONReader struct{}

func (r *invalidJSONReader) Read(dst []byte) (int, error) { return 0, nil }
func (r *invalidJSONReader) Close() error                 { return nil }

func TestOversizedBodyReturnsBadRequest(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	payload := make([]byte, 2<<20)
	payload[0] = '{'
	payload[len(payload)-1] = '}'

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Body = &oversizedReader{}
	req.ContentLength = int64(len(payload))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

type oversizedReader struct{}

func (r *oversizedReader) Read(dst []byte) (int, error) {
	if len(dst) == 0 {
		return 0, nil
	}
	n := copy(dst, make([]byte, len(dst)))
	return n, nil
}

func (r *oversizedReader) Close() error { return nil }

func TestSessionTokenParsingRejectsMalformedInput(t *testing.T) {
	ts := sessions.NewService(nil, 0, 0)
	for _, tok := range []string{"", "not-base64!!", "a===", "AAAA", "AA=="} {
		if _, err := ts.ParsePair(tok); err == nil {
			t.Fatalf("ParsePair(%q) should error", tok)
		}
	}
}
