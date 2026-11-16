package pterodactyl

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetServerStatus_Success(t *testing.T) {
	expected := ServerStatusResponse{
		ID:         1,
		Identifier: "test-node",
		Name:       "test-server",
		Status:     "running",
		Port:       25565,
		Suspended:  false,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, r.Header.Get("Authorization"), "Bearer app-key-123")
		require.Equal(t, r.Header.Get("Accept"), "application/json")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "app-key-123")
	require.NoError(t, err)

	status, err := client.GetServerStatus(context.Background(), "srv-1")
	require.NoError(t, err)
	assert.Equal(t, expected.ID, status.ID)
	assert.Equal(t, expected.Identifier, status.Identifier)
	assert.Equal(t, expected.Name, status.Name)
	assert.Equal(t, expected.Status, status.Status)
}

func TestClient_GetServerStatus_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"errors":[{"message":"Not found"}]}`)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "app-key-123")
	require.NoError(t, err)

	_, err = client.GetServerStatus(context.Background(), "missing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestClient_StartServer_Success(t *testing.T) {
	var receivedCmd string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, r.Method, http.MethodPost)
		require.Equal(t, r.Header.Get("Content-Type"), "application/json")

		var body struct {
			Command string `json:"command"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		receivedCmd = body.Command

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "app-key-123")
	require.NoError(t, err)

	err = client.StartServer(context.Background(), "srv-1")
	require.NoError(t, err)
	assert.Equal(t, "START", receivedCmd)
}

func TestClient_StopServer_Success(t *testing.T) {
	var receivedCmd string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Command string `json:"command"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		receivedCmd = body.Command

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "app-key-123")
	require.NoError(t, err)

	err = client.StopServer(context.Background(), "srv-1")
	require.NoError(t, err)
	assert.Equal(t, "STOP", receivedCmd)
}

func TestClient_AuthenticationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"errors":[{"message":"Unauthorized"}]}`)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "app-key-123")
	require.NoError(t, err)

	_, err = client.GetServerStatus(context.Background(), "srv-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

func TestClient_ForbiddenError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"errors":[{"message":"Forbidden"}]}`)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "app-key-123")
	require.NoError(t, err)

	_, err = client.GetServerStatus(context.Background(), "srv-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}

func TestClient_RateLimit_RetrySuccess(t *testing.T) {
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount < 3 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprint(w, `{"errors":[{"message":"Rate limit exceeded"}]}`)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ServerStatusResponse{
			ID:     1,
			Status: "running",
		})
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "app-key-123", WithTimeout(2*time.Second))
	require.NoError(t, err)

	status, err := client.GetServerStatus(context.Background(), "srv-1")
	require.NoError(t, err)
	assert.Equal(t, 1, status.ID)
	assert.GreaterOrEqual(t, callCount, 3)
}

func TestClient_InternalServerError_GivesUp(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"errors":[{"message":"Server error"}]}`)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "app-key-123", WithTimeout(500*time.Millisecond))
	require.NoError(t, err)

	_, err = client.GetServerStatus(context.Background(), "srv-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gave up")
}

func TestWebhookHandler_ValidSignature(t *testing.T) {
	secret := "webhook-secret"
	payload := []byte(`{"event":"server.started","data":{"id":1}}`)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	signature := fmt.Sprintf("%x", mac.Sum(nil))

	var handledPayload []byte

	server := httptest.NewServer(http.HandlerFunc(HandleWebhook(
		DefaultWebhookConfig(secret, func(body []byte) {
			handledPayload = append([]byte{}, body...)
		}),
	)))
	defer server.Close()

	req, err := http.NewRequest(http.MethodPost, server.URL, bytes.NewReader(payload))
	require.NoError(t, err)
	req.Header.Set("X-Pterodactyl-Signature", signature)
	req.Header.Set("X-Pterodactyl-Event", "server.started")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, payload, handledPayload)
}

func TestWebhookHandler_InvalidSignature(t *testing.T) {
	secret := "webhook-secret"
	payload := []byte(`{"event":"server.started","data":{"id":1}}`)

	server := httptest.NewServer(http.HandlerFunc(HandleWebhook(
		DefaultWebhookConfig(secret, func([]byte) {}),
	)))
	defer server.Close()

	req, err := http.NewRequest(http.MethodPost, server.URL, bytes.NewReader(payload))
	require.NoError(t, err)
	req.Header.Set("X-Pterodactyl-Signature", "deadbeef")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestWebhookHandler_MissingSignature(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(HandleWebhook(
		DefaultWebhookConfig("secret", func([]byte) {}),
	)))
	defer server.Close()

	req, err := http.NewRequest(http.MethodPost, server.URL, bytes.NewReader([]byte(`{}`)))
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestWebhookHandler_WrongMethod(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(HandleWebhook(
		DefaultWebhookConfig("secret", func([]byte) {}),
	)))
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestWebhookHandler_EventFilter(t *testing.T) {
	secret := "webhook-secret"
	payload := []byte(`{"event":"server.deleted","data":{"id":1}}`)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	signature := fmt.Sprintf("%x", mac.Sum(nil))

	var handled bool

	cfg := DefaultWebhookConfig(secret, func([]byte) { handled = true })
	cfg.AllowedEventTypes = []string{"server.started"}

	server := httptest.NewServer(http.HandlerFunc(HandleWebhook(cfg)))
	defer server.Close()

	req, err := http.NewRequest(http.MethodPost, server.URL, bytes.NewReader(payload))
	require.NoError(t, err)
	req.Header.Set("X-Pterodactyl-Signature", signature)
	req.Header.Set("X-Pterodactyl-Event", "server.deleted")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.False(t, handled)
}

func TestWebhookHandler_PayloadTooLarge(t *testing.T) {
	secret := "webhook-secret"
	payload := make([]byte, 300*1024)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	signature := fmt.Sprintf("%x", mac.Sum(nil))

	cfg := DefaultWebhookConfig(secret, func([]byte) {})
	cfg.MaxPayloadBytes = 256 * 1024

	server := httptest.NewServer(http.HandlerFunc(HandleWebhook(cfg)))
	defer server.Close()

	req, err := http.NewRequest(http.MethodPost, server.URL, bytes.NewReader(payload))
	require.NoError(t, err)
	req.Header.Set("X-Pterodactyl-Signature", signature)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// An oversized body is rejected before signature verification, so the
	// handler returns 413 Payload Too Large even when a (truncated) signature
	// is present. This prevents silent truncation of attacker-controlled data.
	assert.Equal(t, http.StatusRequestEntityTooLarge, resp.StatusCode)
}

func TestServerStatusResponse_JSON(t *testing.T) {
	payload := `{
		"id": 42,
		"identifier": "node-1",
		"name": "Survival",
		"current_state": "running",
		"suspended_port": 25565,
		"suspended": false,
		"favourite": true,
		"relationships": {
			"node": {
				"id": 1,
				"name": "Primary"
			}
		}
	}`

	var resp ServerStatusResponse
	require.NoError(t, json.Unmarshal([]byte(payload), &resp))

	assert.Equal(t, 42, resp.ID)
	assert.Equal(t, "node-1", resp.Identifier)
	assert.Equal(t, "Survival", resp.Name)
	assert.Equal(t, "running", resp.Status)
	assert.False(t, resp.Suspended)
	assert.True(t, resp.Favorite)
	assert.Equal(t, 1, resp.Relationships.Node.ID)
	assert.Equal(t, "Primary", resp.Relationships.Node.Name)
}

func TestNewClient_RejectsEmptyBaseURL(t *testing.T) {
	_, err := NewClient("", "key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "base URL is required")
}

func TestNewClient_RejectsEmptyAppKey(t *testing.T) {
	_, err := NewClient("https://panel.example.dev", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "application key is required")
}

func TestNewClient_TrimsTrailingSlash(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/client/servers/srv-1/startup/command", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(server.URL+"/", "key")
	require.NoError(t, err)

	err = client.StartServer(context.Background(), "srv-1")
	require.NoError(t, err)
}

func TestRetryBackoff_CapsAtMax(t *testing.T) {
	backoffs := make([]time.Duration, MaxRetries+1)
	for i := 0; i <= MaxRetries; i++ {
		backoffs[i] = retryBackoff(i)
	}

	for _, d := range backoffs {
		assert.LessOrEqual(t, d, 30*time.Second)
	}

	// Backoff should increase monotonically in expectation due to jitter, but
	// jitter can make the ordering non-strict. We only assert the cap here.
	assert.LessOrEqual(t, backoffs[MaxRetries], 30*time.Second)
}

func TestHandleWebhook_PanicsEarlyForInvalidConfig(t *testing.T) {
	assert.Panics(t, func() {
		HandleWebhook(WebhookConfig{
			Secret:            "",
			Handler:           func([]byte) {},
		})
	})

	assert.Panics(t, func() {
		HandleWebhook(WebhookConfig{
			Secret: "secret",
			Handler: nil,
		})
	})
}

func TestIsServerStopped(t *testing.T) {
	assert.True(t, IsServerStopped("server.stopped"))
	assert.False(t, IsServerStopped("server.started"))
}

func TestIsServerStarted(t *testing.T) {
	assert.True(t, IsServerStarted("server.started"))
	assert.False(t, IsServerStarted("server.stopped"))
}

func TestIsServerCrashed(t *testing.T) {
	// The current implementation treats "server.stopped" as a potential crash.
	assert.True(t, IsServerCrashed("server.stopped"))
	assert.False(t, IsServerCrashed("server.started"))
}

func TestVerifySignature_HexMatch(t *testing.T) {
	secret := "secret"
	body := []byte("hello")

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := fmt.Sprintf("%x", mac.Sum(nil))

	assert.True(t, verifySignature(secret, body, expected, 0))
	assert.True(t, verifySignature(secret, body, strings.ToUpper(expected), 0))
	assert.False(t, verifySignature(secret, body, "bad", 0))
}
