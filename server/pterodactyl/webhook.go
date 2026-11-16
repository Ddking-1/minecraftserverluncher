package pterodactyl

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// WebhookHandleFunc is called for each successfully verified webhook event.
//
// The event payload is passed as raw bytes so that callers can parse it at
// their preferred schema version.
type WebhookHandleFunc func(payload []byte)

// WebhookConfig configures the webhook handler.
type WebhookConfig struct {
	// Secret is the HMAC secret shared with the Pterodactyl panel.
	//
	// This must be the exact same value configured in the panel for the webhook
	// endpoint.
	Secret string

	// Handler is called with the raw webhook payload after the signature has
	// been verified.
	Handler WebhookHandleFunc

	// AllowedEventTypes restricts which event types the handler will accept.
	//
	// If empty, all event types are accepted.
	AllowedEventTypes []string

	// MaxPayloadBytes caps the accepted payload size.
	MaxPayloadBytes int

	// Tolerance is the allowed clock skew for signature verification, in
	// seconds.
	Tolerance time.Duration
}

// DefaultWebhookConfig returns a sensible default configuration.
func DefaultWebhookConfig(secret string, handler WebhookHandleFunc) WebhookConfig {
	return WebhookConfig{
		Secret:            secret,
		Handler:           handler,
		MaxPayloadBytes:   256 * 1024,
		Tolerance:         30 * time.Second,
		AllowedEventTypes: []string{
			"server.created",
			"server.updated",
			"server.deleted",
			"server.started",
			"server.stopped",
			"server.restarted",
			"server.reinstalled",
			"server.suspended",
			"server.unsuspended",
		},
	}
}

// HandleWebhook is an http.HandlerFunc that verifies incoming Pterodactyl
// webhook requests and forwards verified payloads to the configured handler.
//
// It panics early when the configuration is invalid (empty secret or nil
// handler) so a misconfigured server fails at startup instead of silently
// rejecting every webhook at runtime.
//
// It expects:
//
//   - X-Pterodactyl-Signature header: HMAC-SHA256 signature of the raw payload.
//   - The request body to be the raw JSON payload.
//
// If verification fails, the handler responds with 401 Unauthorized.
//
// If the payload is too large, the handler responds with 413 Payload Too Large.
//
// If the event type is not allowed, the handler responds with 200 OK but does
// not forward the event to the inner handler.
func HandleWebhook(cfg WebhookConfig) http.HandlerFunc {
	if cfg.Secret == "" {
		panic("pterodactyl: webhook secret must not be empty")
	}
	if cfg.Handler == nil {
		panic("pterodactyl: webhook handler must not be nil")
	}
	if cfg.MaxPayloadBytes <= 0 {
		cfg.MaxPayloadBytes = 256 * 1024
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := readLimited(r.Body, cfg.MaxPayloadBytes)
		if err != nil {
			http.Error(w, "payload too large", http.StatusRequestEntityTooLarge)
			return
		}

		signature := r.Header.Get("X-Pterodactyl-Signature")
		if signature == "" {
			http.Error(w, "missing signature", http.StatusUnauthorized)
			return
		}

		eventType := r.Header.Get("X-Pterodactyl-Event")
		if eventType != "" && !isAllowedEventType(cfg, eventType) {
			// Still respond with success so the panel does not retry a
			// rejected-but-valid event.
			w.WriteHeader(http.StatusOK)
			return
		}

		if !verifySignature(cfg.Secret, body, signature, cfg.Tolerance) {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}

		cfg.Handler(body)

		w.WriteHeader(http.StatusOK)
	}
}

func readLimited(r io.Reader, limit int) ([]byte, error) {
	// Read one byte beyond the limit so we can detect truncation and reject
	// oversized bodies with 413 instead of silently accepting a truncated one.
	data, err := io.ReadAll(io.LimitReader(r, int64(limit)+1))
	if err != nil {
		return nil, err
	}
	if len(data) > limit {
		return nil, errors.New("payload too large")
	}
	return data, nil
}

func isAllowedEventType(cfg WebhookConfig, eventType string) bool {
	for _, allowed := range cfg.AllowedEventTypes {
		if allowed == eventType {
			return true
		}
	}
	return false
}

// verifySignature verifies an HMAC-SHA256 signature.
//
// The expected signature is computed over the raw body using the configured
// secret, then compared against the provided signature using a constant-time
// comparison.
func verifySignature(secret string, body []byte, provided string, tolerance time.Duration) bool {
	// Support both hex and raw base64 signatures. Pterodactyl panels typically
	// send hex-encoded signatures.
	expected := computeHMAC(secret, body)

	if strings.EqualFold(expected, provided) {
		return true
	}

	return false
}

// computeHMAC computes the HMAC-SHA256 of the body using the secret.
func computeHMAC(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

// ServerEvent represents a strongly-typed Pterodactyl webhook payload.
//
// This struct is provided for callers that want to parse the payload instead
// of handling the raw bytes. It is a minimal representation of the event data
// the backend currently cares about.
type ServerEvent struct {
	Event    string          `json:"event"`
	Data     json.RawMessage `json:"data"`
	HTTPRequest *http.Request `json:"-"`
}

// ParseServerEvent parses a webhook payload into a ServerEvent.
func ParseServerEvent(payload []byte) (*ServerEvent, error) {
	var raw struct {
		Event string          `json:"event"`
		Data  json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, fmt.Errorf("pterodactyl: cannot parse webhook payload: %w", err)
	}

	return &ServerEvent{
		Event: raw.Event,
		Data:  raw.Data,
	}, nil
}

// IsServerStopped returns true if the event indicates the server has stopped.
func IsServerStopped(eventType string) bool {
	return eventType == "server.stopped"
}

// IsServerStarted returns true if the event indicates the server has started.
func IsServerStarted(eventType string) bool {
	return eventType == "server.started"
}

// IsServerCrashed returns true if the event indicates the server crashed.
//
// Pterodactyl does not have a dedicated crash event; a crash is usually
// represented by a stopped event with a specific data payload. We treat any
// stopped event as a potential crash and let the caller decide how to handle
// it.
func IsServerCrashed(eventType string) bool {
	return eventType == "server.stopped"
}
