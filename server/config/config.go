// Package config loads backend configuration from the environment.
//
// No secrets are hardcoded. All sensitive values (database path, listen
// address, session lifetimes) are read from environment variables with
// sensible non-secret defaults. Secret material such as a signing key or
// database credentials must be supplied via environment variables and is
// never embedded in source.
package config

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

// Config holds the runtime configuration for the auth backend.
type Config struct {
	// ListenAddr is the host:port the HTTP server binds to.
	ListenAddr string
	// DBPath is the filesystem path of the SQLite database file.
	DBPath string
	// AccessTokenTTL is the lifetime of a short-lived access token.
	AccessTokenTTL time.Duration
	// RefreshTokenTTL is the lifetime of a refresh token (session).
	// The maximum allowed session lifetime is 30 days.
	RefreshTokenTTL time.Duration
	// BcryptCost is the bcrypt cost used when hashing passwords.
	BcryptCost int
	// AdminToken gates administrative endpoints (e.g. /api/server/start).
	// When empty, those endpoints are disabled. Never hardcoded.
	AdminToken string
	// PteroScript is the path of the server-control script invoked by the
	// admin start endpoint.
	PteroScript string
	// TrustedProxyHeaders enables resolving the client IP from forwarding
	// headers, but only when the immediate peer is in TrustedProxyCIDRs.
	TrustedProxyHeaders bool
	// TrustedProxyCIDRs is the allowlist of reverse proxies that may supply
	// forwarding headers for client IP resolution.
	TrustedProxyCIDRs []*net.IPNet
}

// Default values. These are non-secret operational defaults only.
const (
	defaultListenAddr  = ":8080"
	defaultDBPath      = "auth.db"
	defaultAccessTTL   = 15 * time.Minute
	defaultRefreshTTL  = 30 * 24 * time.Hour // 30 days, per docs/AUTH.md
	defaultBcryptCost  = 12
	defaultPteroScript = "scripts/ptero_control.py"
)

// maxRefreshTTL caps the configured refresh/session lifetime at 30 days.
const maxRefreshTTL = 30 * 24 * time.Hour

// Load reads configuration from environment variables, applying defaults
// where a variable is unset. It returns an error if a provided value is
// invalid.
func Load() (*Config, error) {
	cfg := &Config{
		ListenAddr:      envOr("AUTH_LISTEN_ADDR", defaultListenAddr),
		DBPath:          envOr("AUTH_DB_PATH", defaultDBPath),
		AccessTokenTTL:  defaultAccessTTL,
		RefreshTokenTTL: defaultRefreshTTL,
		BcryptCost:      defaultBcryptCost,
		AdminToken:      os.Getenv("AUTH_ADMIN_TOKEN"),
		PteroScript:     envOr("AUTH_PTERO_SCRIPT", defaultPteroScript),
	}

	if v := os.Getenv("AUTH_ACCESS_TOKEN_TTL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("config: invalid AUTH_ACCESS_TOKEN_TTL %q: %w", v, err)
		}
		cfg.AccessTokenTTL = d
	}

	if v := os.Getenv("AUTH_REFRESH_TOKEN_TTL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("config: invalid AUTH_REFRESH_TOKEN_TTL %q: %w", v, err)
		}
		if d > maxRefreshTTL {
			return nil, fmt.Errorf("config: AUTH_REFRESH_TOKEN_TTL %s exceeds maximum session lifetime of %s", d, maxRefreshTTL)
		}
		cfg.RefreshTokenTTL = d
	}

	if v := os.Getenv("AUTH_BCRYPT_COST"); v != "" {
		var c int
		if _, err := fmt.Sscanf(v, "%d", &c); err != nil || c < 4 || c > 31 {
			return nil, fmt.Errorf("config: invalid AUTH_BCRYPT_COST %q (must be 4-31)", v)
		}
		cfg.BcryptCost = c
	}

	if v := os.Getenv("AUTH_TRUST_PROXY_HEADERS"); v != "" {
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "1", "true", "yes", "on":
			cfg.TrustedProxyHeaders = true
		case "0", "false", "no", "off":
			cfg.TrustedProxyHeaders = false
		default:
			return nil, fmt.Errorf("config: invalid AUTH_TRUST_PROXY_HEADERS %q", v)
		}
	}
	if cfg.TrustedProxyHeaders {
		cidrs := strings.TrimSpace(os.Getenv("AUTH_TRUSTED_PROXY_CIDRS"))
		if cidrs == "" {
			return nil, fmt.Errorf("config: AUTH_TRUSTED_PROXY_CIDRS is required when AUTH_TRUST_PROXY_HEADERS is enabled")
		}
		parsed, err := parseTrustedCIDRs(cidrs)
		if err != nil {
			return nil, err
		}
		cfg.TrustedProxyCIDRs = parsed
	}

	return cfg, nil
}

// envOr returns the value of the named environment variable, or fallback if
// the variable is unset or empty.
func envOr(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}

func parseTrustedCIDRs(value string) ([]*net.IPNet, error) {
	parts := strings.Split(value, ",")
	out := make([]*net.IPNet, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.Contains(part, "/") {
			_, n, err := net.ParseCIDR(part)
			if err != nil {
				return nil, fmt.Errorf("config: invalid AUTH_TRUSTED_PROXY_CIDRS entry %q", part)
			}
			out = append(out, n)
			continue
		}
		ip := net.ParseIP(part)
		if ip == nil {
			return nil, fmt.Errorf("config: invalid AUTH_TRUSTED_PROXY_CIDRS entry %q", part)
		}
		if v4 := ip.To4(); v4 != nil {
			_, n, _ := net.ParseCIDR(v4.String() + "/32")
			out = append(out, n)
		} else {
			_, n, _ := net.ParseCIDR(ip.String() + "/128")
			out = append(out, n)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("config: AUTH_TRUSTED_PROXY_CIDRS has no valid entries")
	}
	return out, nil
}
