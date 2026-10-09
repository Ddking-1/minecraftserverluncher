package pterodactyl

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// DefaultTimeout is the default HTTP client timeout used by Client if the
// caller does not supply a custom timeout.
const DefaultTimeout = 15 * time.Second

// MaxRetries is the maximum number of retry attempts for transient API errors.
const MaxRetries = 4

// Client communicates with a Pterodactyl Panel API.
//
// It is safe for concurrent use by multiple goroutines.
type Client struct {
	baseURL    string
	appKey     string
	httpClient *http.Client
}

// ClientOption configures a Client.
type ClientOption func(*Client) error

// WithTimeout sets the HTTP client timeout.
func WithTimeout(d time.Duration) ClientOption {
	return func(c *Client) error {
		c.httpClient.Timeout = d
		return nil
	}
}

// WithHTTPClient replaces the internal HTTP client entirely.
//
// This is primarily useful in tests or when a custom transport is required.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) error {
		c.httpClient = client
		return nil
	}
}

// NewClient returns a new Pterodactyl API client.
//
// baseURL must be the panel base URL, for example "https://panel.example.dev".
// The client will automatically prefix paths with this base URL.
func NewClient(baseURL string, appKey string, opts ...ClientOption) (*Client, error) {
	if baseURL == "" {
		return nil, errors.New("pterodactyl: base URL is required")
	}
	if appKey == "" {
		return nil, errors.New("pterodactyl: application key is required")
	}

	baseURL = strings.TrimRight(baseURL, "/")

	c := &Client{
		baseURL: baseURL,
		appKey:  appKey,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout: 5 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout:      5 * time.Second,
				ResponseHeaderTimeout:    10 * time.Second,
				ExpectContinueTimeout:    1 * time.Second,
				DisableCompression:       false,
				DisableKeepAlives:        false,
				ForceAttemptHTTP2:        false,
				MaxIdleConns:             100,
				MaxIdleConnsPerHost:      10,
				IdleConnTimeout:          90 * time.Second,
				TLSClientConfig:          nil,
				MaxResponseHeaderBytes:   0,
				Proxy:                    nil,
				GetProxyConnectHeader:    nil,
				DialTLSContext:           nil,
				DialTLS:                  nil,
			},
		},
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// StartServer starts the specified server.
func (c *Client) StartServer(ctx context.Context, serverID string) error {
	return c.postServerAction(ctx, serverID, "START")
}

// StopServer stops the specified server.
func (c *Client) StopServer(ctx context.Context, serverID string) error {
	return c.postServerAction(ctx, serverID, "STOP")
}

// RestartServer restarts the specified server.
func (c *Client) RestartServer(ctx context.Context, serverID string) error {
	return c.postServerAction(ctx, serverID, "RESTART")
}

func (c *Client) postServerAction(ctx context.Context, serverID, action string) error {
	if serverID == "" {
		return errors.New("pterodactyl: server id is required")
	}

	path := fmt.Sprintf("/api/client/servers/%s/startup/command", serverID)

	body := map[string]string{
		"command": action,
	}

	return c.doJSON(ctx, http.MethodPost, path, body, nil)
}

// GetServerStatus returns the current status of the server.
func (c *Client) GetServerStatus(ctx context.Context, serverID string) (*ServerStatusResponse, error) {
	if serverID == "" {
		return nil, errors.New("pterodactyl: server id is required")
	}

	path := fmt.Sprintf("/api/client/servers/%s", serverID)

	var resp ServerStatusResponse
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// ServerStatusResponse is a minimal subset of the Pterodactyl server state
// that the current backend cares about.
type ServerStatusResponse struct {
	ID             int    `json:"id"`
	Identifier     string `json:"identifier"`
	Name           string `json:"name"`
	Status         string `json:"current_state"`
	Port           int    `json:"port"`
	SuspendedPort  int    `json:"suspended_port"`
	Suspended      bool   `json:"suspended"`
	Favorite       bool   `json:"favourite"`
	Relationships  struct {
		Node struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"node"`
	} `json:"relationships"`
}

func (c *Client) doJSON(ctx context.Context, method, path string, body, target interface{}) error {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("pterodactyl: cannot marshal request: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("pterodactyl: cannot create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.appKey)

	resp, err := c.executeWithRetry(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("pterodactyl: cannot read response: %w", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if target != nil {
			if err := json.Unmarshal(respBody, target); err != nil {
				return fmt.Errorf("pterodactyl: cannot parse response: %w", err)
			}
		}
		return nil
	}

	return fmt.Errorf(
		"pterodactyl: request failed: %s %s returned %d: %s",
		method, path, resp.StatusCode,
		strings.TrimSpace(string(respBody)),
	)
}

// executeWithRetry executes an HTTP request and retries on transient errors.
//
// Retryable statuses are 429 (rate limit) and 5xx server errors. Retries use
// exponential backoff with a small jitter.
func (c *Client) executeWithRetry(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= MaxRetries; attempt++ {
		resp, err = c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == http.StatusTooManyRequests || isServerError(resp.StatusCode) {
			resp.Body.Close()

			if attempt == MaxRetries {
				return nil, fmt.Errorf(
					"pterodactyl: gave up after %d retries, last status %d",
					MaxRetries, resp.StatusCode,
				)
			}

			delay := retryBackoff(attempt)
			time.Sleep(delay)
			continue
		}

		return resp, nil
	}

	return nil, errors.New("pterodactyl: unreachable retry exit")
}

func isServerError(status int) bool {
	return status >= 500 && status < 600
}

func retryBackoff(attempt int) time.Duration {
	base := 1 * time.Second
	exp := time.Duration(attempt+1) * base
	jitter := time.Duration(nanosecondJitter(0, 100)) * time.Millisecond
	d := exp + jitter

	if d > 30*time.Second {
		d = 30 * time.Second
	}

	return d
}

func nanosecondJitter(min, max int) int {
	const modulus = 101
	current := min + int(time.Now().Nanosecond())%modulus
	if current > max {
		current = max
	}
	return current
}
