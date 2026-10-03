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
	"os"
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
}

// Default values. These are non-secret operational defaults only.
const (
	defaultListenAddr = ":8080"
	defaultDBPath     = "auth.db"
	defaultAccessTTL  = 15 * time.Minute
	defaultRefreshTTL = 30 * 24 * time.Hour // 30 days, per docs/AUTH.md
	defaultBcryptCost = 12
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
