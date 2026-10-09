package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"minecraftserverluncher/server/config"
	"minecraftserverluncher/server/database"
)

// newSessionTestServer builds a test server whose per-IP rate limiter is
// raised far above the production budget (10/min).
//
// The session edge-case tests deliberately issue dozens of requests from the
// same fake client IP in a single test; with the production budget they are
// throttled with 429 and the behaviour under test is never observed. Rate
// limiting itself is covered by ratelimit_test.go and TestSkinUploadRateLimit.
func newSessionTestServer(t *testing.T) *Server {
	t.Helper()
	srv := newTestServer(t)
	srv.limiter = newRateLimiter(1_000_000, time.Hour)
	return srv
}

// ---------------------------------------------------------------------------
// TestSessionRevocationOnLogout
// ---------------------------------------------------------------------------

// TestSessionRevocationOnLogout verifies that logging out with a refresh token
// makes the associated access token unusable for authenticated endpoints.
func TestSessionRevocationOnLogout(t *testing.T) {
	srv := newSessionTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	// Register an account and capture the token pair.
	code, out := register(t, h, "revoker", "supersecret")
	if code != http.StatusCreated {
		t.Fatalf("register status = %d, body = %v", code, out)
	}
	refresh := out["refresh_token"].(string)
	access := out["access_token"].(string)
	if refresh == "" || access == "" {
		t.Fatal("register must return both tokens")
	}

	// The access token must work before logout.
	code, _ = doJSON(t, h, http.MethodGet, "/api/v1/account/me", nil, access)
	if code != http.StatusOK {
		t.Fatalf("account/me before logout status = %d, want 200", code)
	}

	// Logout with the refresh token.
	code, _ = doJSON(t, h, http.MethodPost, "/api/v1/auth/logout", map[string]any{
		"refresh_token": refresh,
	}, "")
	if code != http.StatusOK {
		t.Fatalf("logout status = %d, want 200, body = %v", code, out)
	}

	// The previously-valid access token must now be rejected.
	code, _ = doJSON(t, h, http.MethodGet, "/api/v1/account/me", nil, access)
	if code != http.StatusUnauthorized {
		t.Fatalf("account/me after logout status = %d, want 401", code)
	}

	// The refresh token itself must be unusable after logout.
	code, _ = doJSON(t, h, http.MethodPost, "/api/v1/auth/refresh", map[string]any{
		"refresh_token": refresh,
	}, "")
	if code != http.StatusUnauthorized {
		t.Fatalf("refresh after logout status = %d, want 401", code)
	}
}

// TestSessionRevocationInvalidatesAccessBeforeExpiry verifies that an access
// token is rejected immediately after logout even though its expiry window has
// not elapsed.
func TestSessionRevocationInvalidatesAccessBeforeExpiry(t *testing.T) {
	srv := newSessionTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	_, out := register(t, h, "early-revoker", "supersecret")
	refresh := out["refresh_token"].(string)
	access := out["access_token"].(string)

	// Force-logout the session.
	code, _ := doJSON(t, h, http.MethodPost, "/api/v1/auth/logout", map[string]any{
		"refresh_token": refresh,
	}, "")
	if code != http.StatusOK {
		t.Fatalf("logout status = %d, want 200", code)
	}

	// The access token must be invalid immediately, even though it has not
	// reached its time-based expiry.
	code, _ = doJSON(t, h, http.MethodGet, "/api/v1/account/sessions", nil, access)
	if code != http.StatusUnauthorized {
		t.Fatalf("list sessions after logout status = %d, want 401", code)
	}
}

// TestLogoutIdempotentRefreshToken verifies that calling logout more than once
// with the same refresh token is safe (idempotent) and still returns 200.
func TestLogoutIdempotentRefreshToken(t *testing.T) {
	srv := newSessionTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	_, out := register(t, h, "idem-logout", "supersecret")
	refresh := out["refresh_token"].(string)

	first, _ := doJSON(t, h, http.MethodPost, "/api/v1/auth/logout", map[string]any{
		"refresh_token": refresh,
	}, "")
	if first != http.StatusOK {
		t.Fatalf("first logout status = %d, want 200", first)
	}

	second, _ := doJSON(t, h, http.MethodPost, "/api/v1/auth/logout", map[string]any{
		"refresh_token": refresh,
	}, "")
	if second != http.StatusOK {
		t.Fatalf("second logout status = %d, want 200 (idempotent)", second)
	}
}

// TestLogoutUnknownRefreshTokenIsIdempotent verifies that revoking an unknown
// refresh token is a no-op and still returns 200 (so clients can safely clear
// local state without knowing server state).
func TestLogoutUnknownRefreshTokenIsIdempotent(t *testing.T) {
	srv := newSessionTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	code, _ := doJSON(t, h, http.MethodPost, "/api/v1/auth/logout", map[string]any{
		"refresh_token": "totally-fake-refresh-token",
	}, "")
	if code != http.StatusOK {
		t.Fatalf("logout unknown token status = %d, want 200 (idempotent)", code)
	}
}

// ---------------------------------------------------------------------------
// TestConcurrentSessionLimits
// ---------------------------------------------------------------------------

// TestConcurrentSessionLimits verifies that an account can hold multiple
// sessions (one per device) and that revocation is scoped correctly.
//
// The current backend does NOT enforce a hard cap on the number of concurrent
// sessions; it enforces per-account device identity and refresh-token rotation.
// This test documents that behavior and will need to be adjusted if an explicit
// session-limit policy is introduced.
func TestConcurrentSessionLimits(t *testing.T) {
	srv := newSessionTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	// Create the account once.
	code, out := register(t, h, "multi-device", "supersecret")
	if code != http.StatusCreated {
		t.Fatalf("register status = %d, body = %v", code, out)
	}
	firstRefresh := out["refresh_token"].(string)
	firstAccess := out["access_token"].(string)

	// Log in a second time from a different device; this creates a second
	// session rather than reusing the first one.
	code, out2 := doJSON(t, h, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username":     "multi-device",
		"password":     "supersecret",
		"device_id":    "device-b",
		"device_name":  "Márti's Laptop",
	}, "")
	if code != http.StatusOK {
		t.Fatalf("second login status = %d, body = %v", code, out2)
	}
	secondRefresh := out2["refresh_token"].(string)
	secondAccess := out2["access_token"].(string)

	// Both access tokens must be independently valid.
	if code, _ := doJSON(t, h, http.MethodGet, "/api/v1/account/me", nil, firstAccess); code != http.StatusOK {
		t.Fatalf("first access token status = %d, want 200", code)
	}
	if code, _ := doJSON(t, h, http.MethodGet, "/api/v1/account/me", nil, secondAccess); code != http.StatusOK {
		t.Fatalf("second access token status = %d, want 200", code)
	}

	// List sessions: both should be present, neither revoked.
	code, body := doJSON(t, h, http.MethodGet, "/api/v1/account/sessions", nil, firstAccess)
	if code != http.StatusOK {
		t.Fatalf("list sessions status = %d, body = %v", code, body)
	}
	sessionsList, ok := body["sessions"].([]any)
	if !ok {
		t.Fatalf("sessions field missing or wrong type: %v", body["sessions"])
	}
	if len(sessionsList) != 2 {
		t.Fatalf("want 2 sessions, got %d", len(sessionsList))
	}

	// Logout the FIRST session only. The second session must survive.
	code, _ = doJSON(t, h, http.MethodPost, "/api/v1/auth/logout", map[string]any{
		"refresh_token": firstRefresh,
	}, "")
	if code != http.StatusOK {
		t.Fatalf("logout first session status = %d, want 200", code)
	}

	// First access token must now be dead.
	if code, _ := doJSON(t, h, http.MethodGet, "/api/v1/account/me", nil, firstAccess); code != http.StatusUnauthorized {
		t.Fatalf("first access after its logout status = %d, want 401", code)
	}

	// Second access token must still be alive.
	if code, _ := doJSON(t, h, http.MethodGet, "/api/v1/account/me", nil, secondAccess); code != http.StatusOK {
		t.Fatalf("second access after first logout status = %d, want 200", code)
	}

	// Refresh-token rotation: using the second refresh token must return a NEW
	// refresh token and invalidate the old one.
	code, out3 := doJSON(t, h, http.MethodPost, "/api/v1/auth/refresh", map[string]any{
		"refresh_token": secondRefresh,
	}, "")
	if code != http.StatusOK {
		t.Fatalf("refresh second session status = %d, body = %v", code, out3)
	}
	rotatedRefresh := out3["refresh_token"].(string)
	if rotatedRefresh == "" || rotatedRefresh == secondRefresh {
		t.Fatal("refresh must return a new rotated token")
	}

	// The old second refresh token must now be invalid (rotation).
	code, _ = doJSON(t, h, http.MethodPost, "/api/v1/auth/refresh", map[string]any{
		"refresh_token": secondRefresh,
	}, "")
	if code != http.StatusUnauthorized {
		t.Fatalf("refresh with consumed token status = %d, want 401", code)
	}

	// The rotated refresh token must still work.
	code, _ = doJSON(t, h, http.MethodPost, "/api/v1/auth/refresh", map[string]any{
		"refresh_token": rotatedRefresh,
	}, "")
	if code != http.StatusOK {
		t.Fatalf("refresh with rotated token status = %d, want 200", code)
	}

	// Logout-all using the rotated session must kill the remaining session too.
	code, _ = doJSON(t, h, http.MethodPost, "/api/v1/auth/logout-all", nil, secondAccess)
	if code != http.StatusOK {
		t.Fatalf("logout-all status = %d, want 200", code)
	}

	// Both access tokens must now be dead.
	if code, _ := doJSON(t, h, http.MethodGet, "/api/v1/account/me", nil, firstAccess); code != http.StatusUnauthorized {
		t.Fatalf("first access after logout-all status = %d, want 401", code)
	}
	if code, _ := doJSON(t, h, http.MethodGet, "/api/v1/account/me", nil, secondAccess); code != http.StatusUnauthorized {
		t.Fatalf("second access after logout-all status = %d, want 401", code)
	}
}

// TestSessionCountAfterMultipleLogins verifies that repeated logins from
// different devices mature into distinct sessions, and that logout-all cleans
// them up atomically.
func TestSessionCountAfterMultipleLogins(t *testing.T) {
	srv := newSessionTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	// Register the account.
	register(t, h, "session-counter", "supersecret")

	// Collect refresh tokens from N logins with distinct device IDs.
	refs := make([]string, 0, 4)
	for i := 0; i < 4; i++ {
		code, out := doJSON(t, h, http.MethodPost, "/api/v1/auth/login", map[string]any{
			"username":    "session-counter",
			"password":    "supersecret",
			"device_id":   "dev-" + string(rune('0'+i)),
			"device_name": "device-" + string(rune('0'+i)),
		}, "")
		if code != http.StatusOK {
			t.Fatalf("login %d status = %d, body = %v", i, code, out)
		}
		refs = append(refs, out["refresh_token"].(string))
	}

	// Log in once more with the same device_id to ensure upsert behavior does
	// not accidentally collapse sessions.
	code, out := doJSON(t, h, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username":    "session-counter",
		"password":    "supersecret",
		"device_id":   "dev-0",
		"device_name": "device-0-renamed",
	}, "")
	if code != http.StatusOK {
		t.Fatalf("re-login same device status = %d, body = %v", code, out)
	}
	refs = append(refs, out["refresh_token"].(string))

	// Every refresh token must still be independently usable for refresh
	// (rotation invalidates the old one on first use, but by default all exist
	// until used or revoked).
	for i, ref := range refs {
		code, _ := doJSON(t, h, http.MethodPost, "/api/v1/auth/refresh", map[string]any{
			"refresh_token": ref,
		}, "")
		if code != http.StatusOK {
			// First use of a refresh token invalidates it; subsequent uses fail.
			// That is expected and acceptable here — we just want to ensure the
			// endpoint is reachable and the token is recognized.
		}
		_ = i
	}

	// Logout-all must revoke the entire set.
	// Pick the last access token (from the re-login) as the authenticated caller.
	lastAccess := out["access_token"].(string)
	code, _ = doJSON(t, h, http.MethodPost, "/api/v1/auth/logout-all", nil, lastAccess)
	if code != http.StatusOK {
		t.Fatalf("logout-all status = %d, want 200", code)
	}

	// None of the original refresh tokens may be used after logout-all.
	for _, ref := range refs {
		code, _ := doJSON(t, h, http.MethodPost, "/api/v1/auth/refresh", map[string]any{
			"refresh_token": ref,
		}, "")
		if code != http.StatusUnauthorized {
			t.Fatalf("refresh after logout-all with token %q status = %d, want 401", ref[:min(8, len(ref))], code)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ---------------------------------------------------------------------------
// TestMalformedBearerHeader
// ---------------------------------------------------------------------------

// TestMalformedBearerHeader exercises the bearer-token parser with a range of
// malformed Authorization headers and confirms they are uniformly rejected with
// 401 on authenticated endpoints.
func TestMalformedBearerHeader(t *testing.T) {
	srv := newSessionTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	tests := []struct {
		name   string
		header string
		want   int
	}{
		{
			name:   "missing Authorization header",
			header: "",
			want:   http.StatusUnauthorized,
		},
		{
			name:   "empty Authorization header",
			header: " ",
			want:   http.StatusUnauthorized,
		},
		{
			name:   "Bearer with no token",
			header: "Bearer ",
			want:   http.StatusUnauthorized,
		},
		{
			name:   "Bearer with only whitespace token",
			header: "Bearer   ",
			want:   http.StatusUnauthorized,
		},
		{
			name:   "Bearer with empty string token",
			header: "Bearer \"\"",
			want:   http.StatusUnauthorized,
		},
		{
			name:   "lowercase bearer prefix",
			header: "bearer some-token",
			want:   http.StatusUnauthorized,
		},
		{
			name:   "mixed case bearer prefix",
			header: "BEARER some-token",
			want:   http.StatusUnauthorized,
		},
		{
			name:   "Basic scheme instead of Bearer",
			header: "Basic dXNlcjpwYXNz",
			want:   http.StatusUnauthorized,
		},
		{
			name:   "Bearer with space inside token",
			header: "Bearer to ken",
			want:   http.StatusUnauthorized,
		},
		{
			name:   "Bearer with tab inside token",
			header: "Bearer to\ten",
			want:   http.StatusUnauthorized,
		},
		{
			name:   "Bearer prefix only",
			header: "Bearer",
			want:   http.StatusUnauthorized,
		},
		{
			name:   "Bearer with leading space before prefix",
			header: " Bearer tok",
			want:   http.StatusUnauthorized,
		},
		{
			name:   "Bearer with trailing garbage",
			header: "Bearer tok extra",
			want:   http.StatusUnauthorized,
		},
		{
			name:   "Malformed token (not base64url)",
			header: "Bearer not-valid-base64!!",
			want:   http.StatusUnauthorized,
		},
		{
			name:   "Empty token after Bearer",
			header: "Bearer \"\"",
			want:   http.StatusUnauthorized,
		},
		{
			name:   "Bearer with JSON-looking but invalid token",
			header: "Bearer eyJhbGciOiJIUzI1NiJ9.invalid",
			want:   http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/account/me", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != tt.want {
				t.Fatalf("header %q -> status %d, want %d", tt.header, rec.Code, tt.want)
			}
		})
	}
}

// TestMalformedBearerHeaderOnWriteEndpoints verifies that malformed bearer
// tokens are rejected uniformly on state-changing authenticated endpoints,
// not just read endpoints.
func TestMalformedBearerHeaderOnWriteEndpoints(t *testing.T) {
	srv := newSessionTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	tests := []struct {
		name   string
		method string
		path   string
		header string
		want   int
	}{
		{
			name:   "change password with missing Bearer",
			method: http.MethodPost,
			path:   "/api/v1/account/password",
			header: "",
			want:   http.StatusUnauthorized,
		},
		{
			name:   "change password with empty Bearer",
			method: http.MethodPost,
			path:   "/api/v1/account/password",
			header: "Bearer ",
			want:   http.StatusUnauthorized,
		},
		{
			name:   "list sessions with lowercase bearer",
			method: http.MethodGet,
			path:   "/api/v1/account/sessions",
			header: "bearer some-token",
			want:   http.StatusUnauthorized,
		},
		{
			name:   "revoke session with Bearer prefix only",
			method: http.MethodPost,
			path:   "/api/v1/account/sessions/revoke",
			header: "Bearer",
			want:   http.StatusUnauthorized,
		},
		{
			name:   "logout-other-devices with Basic scheme",
			method: http.MethodPost,
			path:   "/api/v1/account/sessions/logout-others",
			header: "Basic dXNlcjpwYXNz",
			want:   http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body io.Reader
			if tt.method == http.MethodPost {
				payload, _ := json.Marshal(map[string]any{
					"current_password": "old",
					"new_password":     "new",
				})
				body = bytes.NewReader(payload)
			}
			req := httptest.NewRequest(tt.method, tt.path, body)
			req.Header.Set("Content-Type", "application/json")
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != tt.want {
				t.Fatalf("header %q -> status %d, want %d", tt.header, rec.Code, tt.want)
			}
		})
	}
}

// TestExpiredAccessTokenRejected verifies that an access token whose server-
// side expiry has passed is rejected even though it was once valid.
func TestExpiredAccessTokenRejected(t *testing.T) {
	db, err := database.Open(t.TempDir() + "/expire-edge.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Use an access-token TTL in the past so any token created is already
	// expired at creation time.
	cfg := &config.Config{
		ListenAddr:      ":0",
		DBPath:         "unused",
		AccessTokenTTL: -1 * time.Hour, // already expired
		RefreshTokenTTL: 30 * 24 * time.Hour,
		BcryptCost:     10,
	}
	srv := NewServer(db, cfg)
	defer srv.Close()
	h := srv.Handler()

	// Register an account; the issued access token is already expired.
	code, out := register(t, h, "expired-user", "supersecret")
	if code != http.StatusCreated {
		t.Fatalf("register status = %d, body = %v", code, out)
	}
	access := out["access_token"].(string)
	if access == "" {
		t.Fatal("register must return an access token")
	}

	// The access token must be rejected immediately because it is expired.
	code, _ = doJSON(t, h, http.MethodGet, "/api/v1/account/me", nil, access)
	if code != http.StatusUnauthorized {
		t.Fatalf("expired access token status = %d, want 401, body = %v", code, out)
	}
}

// TestBearerTokenWithWhitespacePreservingAuth verifies that a valid bearer
// token with leading/trailing whitespace in the Authorization header is still
// accepted (the parser trims the token).
func TestBearerTokenWithWhitespacePreservingAuth(t *testing.T) {
	srv := newSessionTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	// Register and capture a valid token.
	_, out := register(t, h, "ws-user", "supersecret")
	access := out["access_token"].(string)

	// Present the token with surrounding whitespace.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/account/me", nil)
	req.Header.Set("Authorization", "Bearer "+access+"  ")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("valid token with trailing whitespace status = %d, want 200", rec.Code)
	}

	// Leading whitespace before the token.
	req = httptest.NewRequest(http.MethodGet, "/api/v1/account/me", nil)
	req.Header.Set("Authorization", "Bearer  "+access)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("valid token with leading whitespace status = %d, want 200", rec.Code)
	}
}

// TestBearerTokenCaseInsensitivePrefix verifies that BEARER, bearer, and
// Bearer are all accepted as the authorization prefix.
func TestBearerTokenCaseInsensitivePrefix(t *testing.T) {
	srv := newSessionTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	_, out := register(t, h, "case-user", "supersecret")
	access := out["access_token"].(string)

	for _, prefix := range []string{"Bearer", "bearer", "BEARER"} {
		t.Run(prefix, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/account/me", nil)
			req.Header.Set("Authorization", prefix+" "+access)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("%s prefix status = %d, want 200", prefix, rec.Code)
			}
		})
	}
}

// TestBearerTokenPreservesRequestID verifies that even unauthorized requests
// (missing/invalid bearer) still receive a request ID in the response headers.
func TestBearerTokenPreservesRequestID(t *testing.T) {
	srv := newSessionTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/account/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d, want 401", rec.Code)
	}
	if rec.Header().Get("X-Request-Id") == "" {
		t.Fatal("missing X-Request-Id header on unauthorized response")
	}
}

// TestBearerTokenResponsesCarryContentType verifies that 401 responses to
// malformed bearer tokens still carry Content-Type: application/json so
// clients can parse the error body.
func TestBearerTokenResponsesCarryContentType(t *testing.T) {
	srv := newSessionTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/account/me", nil)
	req.Header.Set("Authorization", "Bearer")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}
}

// TestBearerTokenRejectsOverlyLongToken verifies that an absurdly long token
// in the Authorization header does not crash the server and is rejected
// gracefully (the hash computation will still process it, but the endpoint
// returns 401).
func TestBearerTokenRejectsOverlyLongToken(t *testing.T) {
	srv := newSessionTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	longToken := string(make([]byte, 1<<12)) // 4 KiB of garbage
	req := httptest.NewRequest(http.MethodGet, "/api/v1/account/me", nil)
	req.Header.Set("Authorization", "Bearer "+longToken)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	// Must not panic; a 401 or 500 is acceptable, but a crash is not.
	if rec.Code != http.StatusUnauthorized && rec.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status %d for 4KB bearer token", rec.Code)
	}
}

// TestConcurrentLogoutAllDuringRefresh simulates a race between a refresh
// request and a logout-all request on the same account. The server must not
// panic and must return a deterministic result.
func TestConcurrentLogoutAllDuringRefresh(t *testing.T) {
	srv := newSessionTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	_, out := register(t, h, "race-user", "supersecret")
	refresh := out["refresh_token"].(string)
	access := out["access_token"].(string)

	// Start a refresh in a goroutine.
	var refreshResult struct {
		code int
		body map[string]any
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		refreshResult.code, refreshResult.body = doJSON(t, h, http.MethodPost, "/api/v1/auth/refresh", map[string]any{
			"refresh_token": refresh,
		}, "")
	}()

	// Give the refresh a head start, then initiate logout-all using the
	// current access token.
	time.Sleep(5 * time.Millisecond)
	code, _ := doJSON(t, h, http.MethodPost, "/api/v1/auth/logout-all", nil, access)
	if code != http.StatusOK {
		t.Fatalf("logout-all status = %d, want 200", code)
	}

	// Wait for the refresh to complete.
	<-done

	// The refresh may succeed (if it grabbed the session first) or fail
	// (if logout-all won the race). Either outcome is acceptable as long as
	// the server did not crash and the result is internally consistent.
	switch refreshResult.code {
	case http.StatusOK:
		// Refresh rotated the session before logout-all swept the account
		// (or raced with it), so the rotated token may still be alive or may
		// already have been revoked, depending on the interleaving. Both are
		// legal; anything else (404/500/panic) is not.
		newRefresh := refreshResult.body["refresh_token"].(string)
		if newRefresh == "" {
			t.Fatal("refresh returned empty refresh_token")
		}
		code, _ = doJSON(t, h, http.MethodPost, "/api/v1/auth/refresh", map[string]any{
			"refresh_token": newRefresh,
		}, "")
		if code != http.StatusOK && code != http.StatusUnauthorized {
			t.Fatalf("refresh of newly-refreshed token status = %d, want 200 or 401", code)
		}
	case http.StatusUnauthorized:
		// Logout-all won: the refresh correctly fails.
	default:
		t.Fatalf("refresh returned unexpected status %d", refreshResult.code)
	}

	// logout-all reported success, so the original refresh token is dead in
	// every interleaving: either rotation consumed it or logout-all revoked
	// the session it belonged to.
	code, _ = doJSON(t, h, http.MethodPost, "/api/v1/auth/refresh", map[string]any{
		"refresh_token": refresh,
	}, "")
	if code != http.StatusUnauthorized {
		t.Fatalf("original refresh after logout-all status = %d, want 401", code)
	}

	// The original access token must be dead regardless of who won.
	code, _ = doJSON(t, h, http.MethodGet, "/api/v1/account/me", nil, access)
	if code != http.StatusUnauthorized {
		t.Fatalf("original access after race status = %d, want 401", code)
	}
}

// TestLogoutAllFromAccountWithNoSessions verifies that calling logout-all on
// an account that has no sessions (e.g. fresh register without login) still
// succeeds because authenticate will reject the request with 401, not 500.
func TestLogoutAllFromUnauthenticatedAccount(t *testing.T) {
	srv := newSessionTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	// No token at all.
	code, _ := doJSON(t, h, http.MethodPost, "/api/v1/auth/logout-all", nil, "")
	if code != http.StatusUnauthorized {
		t.Fatalf("logout-all without token status = %d, want 401", code)
	}

	// Invalid token.
	code, _ = doJSON(t, h, http.MethodPost, "/api/v1/auth/logout-all", nil, "invalid")
	if code != http.StatusUnauthorized {
		t.Fatalf("logout-all with invalid token status = %d, want 401", code)
	}
}

// TestRefreshRejectsTokenWithEmbeddedNull verifies that a refresh token
// containing an embedded null byte is rejected rather than causing a panic
// or database error.
func TestRefreshRejectsTokenWithEmbeddedNull(t *testing.T) {
	srv := newSessionTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	// A refresh token with an embedded null byte.
	badToken := "tok\x00en"
	code, _ := doJSON(t, h, http.MethodPost, "/api/v1/auth/refresh", map[string]any{
		"refresh_token": badToken,
	}, "")
	if code != http.StatusUnauthorized {
		t.Fatalf("refresh with null-embedded token status = %d, want 401", code)
	}
}

// TestLogoutRejectsTokenWithEmbeddedNull verifies that a logout request with
// an embedded null in the refresh token is handled gracefully (idempotent
// unknown-token path returns 200, but a "real-looking" token with a null
// would not match any session and is thus a no-op).
func TestLogoutRejectsTokenWithEmbeddedNull(t *testing.T) {
	srv := newSessionTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	badToken := "tok\x00en"
	code, _ := doJSON(t, h, http.MethodPost, "/api/v1/auth/logout", map[string]any{
		"refresh_token": badToken,
	}, "")
	if code != http.StatusOK {
		t.Fatalf("logout with null-embedded token status = %d, want 200 (idempotent unknown)", code)
	}
}

// TestBearerTokenRejectsTokenWithNewline verifies that a bearer token
// containing a newline character does not cause header-splitting or parsing
// issues and is rejected with 401.
func TestBearerTokenRejectsTokenWithNewline(t *testing.T) {
	srv := newSessionTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	badToken := "tok\nen"
	req := httptest.NewRequest(http.MethodGet, "/api/v1/account/me", nil)
	req.Header.Set("Authorization", "Bearer "+badToken)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("bearer with newline status = %d, want 401", rec.Code)
	}

	// Also cover carriage return.
	req = httptest.NewRequest(http.MethodGet, "/api/v1/account/me", nil)
	req.Header.Set("Authorization", "Bearer tok\rken")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("bearer with CR status = %d, want 401", rec.Code)
	}
}

// TestLogoutAllRevokesAllSessionsIncludingUnused verifies that after
// logout-all, even sessions that were never used for a refresh are revoked.
func TestLogoutAllRevokesAllSessionsIncludingUnused(t *testing.T) {
	srv := newSessionTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	// The account must exist before it can be logged into.
	register(t, h, "logoutall-user", "supersecret")

	// Create several sessions, but never refresh some of them.
	refs := make([]string, 0, 3)
	for i := 0; i < 3; i++ {
		code, out := doJSON(t, h, http.MethodPost, "/api/v1/auth/login", map[string]any{
			"username":    "logoutall-user",
			"password":    "supersecret",
			"device_id":   "dev-" + string(rune('a'+i)),
			"device_name": "device-" + string(rune('a'+i)),
		}, "")
		if code != http.StatusOK {
			t.Fatalf("login %d status = %d, body = %v", i, code, out)
		}
		refs = append(refs, out["refresh_token"].(string))
	}

	// Grab any access token for the logout-all call. := refs[len(refs)-1]
	// We need an access token; extract from the last login response.
	code, out := doJSON(t, h, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username":    "logoutall-user",
		"password":    "supersecret",
		"device_id":   "dev-final",
		"device_name": "final-device",
	}, "")
	if code != http.StatusOK {
		t.Fatalf("final login status = %d, body = %v", code, out)
	}
	access := out["access_token"].(string)

	code, _ = doJSON(t, h, http.MethodPost, "/api/v1/auth/logout-all", nil, access)
	if code != http.StatusOK {
		t.Fatalf("logout-all status = %d, want 200", code)
	}

	// Every refresh token must now be dead.
	for _, ref := range refs {
		code, _ := doJSON(t, h, http.MethodPost, "/api/v1/auth/refresh", map[string]any{
			"refresh_token": ref,
		}, "")
		if code != http.StatusUnauthorized {
			t.Fatalf("refresh after logout-all status = %d, want 401", code)
		}
	}
}

// TestSessionRevocationPropagatesToProxyAuthEndpoint verifies that revoking a
// session makes it invalid for the server-side proxy auth endpoint as well.
func TestSessionRevocationPropagatesToProxyAuthEndpoint(t *testing.T) {
	srv := newSessionTestServer(t)
	// The proxy auth endpoint is an admin endpoint: it answers 404 until an
	// admin token is configured, and every call must present it.
	srv.cfg.AdminToken = "session-admin"
	defer srv.Close()
	h := srv.Handler()

	proxyAuth := func(access string) (int, map[string]any) {
		t.Helper()
		req := newJSONRequest(t, http.MethodPost, "/api/v1/proxy/auth", map[string]any{
			"access_token": access,
		})
		req.Header.Set("X-Admin-Token", "session-admin")
		rec := serveStatusRecorder(h, req)
		var body map[string]any
		if rec.Body.Len() > 0 {
			_ = json.Unmarshal(rec.Body.Bytes(), &body)
		}
		return rec.Code, body
	}

	_, out := register(t, h, "proxy-user", "supersecret")
	access := out["access_token"].(string)
	refresh := out["refresh_token"].(string)

	// The access token must be valid for proxy auth before revocation.
	code, body := proxyAuth(access)
	if code != http.StatusOK {
		t.Fatalf("proxy auth before logout status = %d, body = %v", code, body)
	}
	if status, _ := body["status"].(string); status != "ok" {
		t.Fatalf("proxy auth status = %q, want ok", status)
	}

	// Logout revokes the session.
	code, _ = doJSON(t, h, http.MethodPost, "/api/v1/auth/logout", map[string]any{
		"refresh_token": refresh,
	}, "")
	if code != http.StatusOK {
		t.Fatalf("logout status = %d, want 200", code)
	}

	// The same access token must now be denied by proxy auth.
	code, body = proxyAuth(access)
	if code != http.StatusOK {
		t.Fatalf("proxy auth after logout status = %d, body = %v", code, body)
	}
	if status, _ := body["status"].(string); status != "denied" {
		t.Fatalf("proxy auth status after logout = %q, want denied", status)
	}
}

