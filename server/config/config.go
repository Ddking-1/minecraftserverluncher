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
	// LauncherVersion is the current launcher version reported by the update
	// manifest.
	LauncherVersion string
	// UpdateBaseURL is the base URL from which launcher installers are served.
	UpdateBaseURL string
	// UpdateSHA256 is the SHA-256 of the currently shipped installer.
	UpdateSHA256 string
	// UpdateMandatory forces clients to update regardless of choice.
	UpdateMandatory bool
	// UpdateSigningKey signs the update manifest (HMAC-SHA256) so clients can
	// verify integrity. Never hardcoded; read from the environment.
	UpdateSigningKey string
	// RateLimitPerMinute is the per-client request budget applied by the API
	// rate limiter.
	RateLimitPerMinute int
	// TrustProxy makes rate limiting key on X-Forwarded-For / X-Real-IP.
	// Enable only when every request arrives through a trusted reverse proxy
	// that overwrites those headers.
	TrustProxy bool
	// SkinStorageDir is the directory where processed skins and avatars are
	// persisted. Skins survive a restart only when this is a writable path.
	SkinStorageDir string
	// TurnstileSiteKey is the public Cloudflare Turnstile site key (embedded in
	// the frontend). Empty disables CAPTCHA enforcement.
	TurnstileSiteKey string
	// TurnstileSecret is the server-side Cloudflare Turnstile secret used to
	// verify tokens. Empty disables CAPTCHA enforcement (dev mode).
	TurnstileSecret string
}

// Default values. These are non-secret operational defaults only.
const (
	defaultListenAddr      = ":8080"
	defaultDBPath          = "auth.db"
	defaultAccessTTL       = 15 * time.Minute
	defaultRefreshTTL      = 30 * 24 * time.Hour // 30 days, per docs/AUTH.md
	defaultBcryptCost      = 12
	defaultPteroScript     = "scripts/ptero_control.py"
	defaultLauncherVersion = "0.1.0"
	defaultUpdateBaseURL   = ""
	defaultRateLimitPerMin = 60
	defaultSkinStorageDir  = "./storage/skins"
	maxRateLimitPerMinute  = 100000
	minRateLimitPerMinute  = 1
)

// maxRefreshTTL caps the configured refresh/session lifetime at 30 days.
const maxRefreshTTL = 30 * 24 * time.Hour

// Load reads configuration from environment variables, applying defaults
// where a variable is unset. It returns an error if a provided value is
// invalid.
func Load() (*Config, error) {
	cfg := &Config{
		ListenAddr:         envOr("AUTH_LISTEN_ADDR", defaultListenAddr),
		DBPath:             envOr("AUTH_DB_PATH", defaultDBPath),
		AccessTokenTTL:     defaultAccessTTL,
		RefreshTokenTTL:    defaultRefreshTTL,
		BcryptCost:         defaultBcryptCost,
		AdminToken:         os.Getenv("AUTH_ADMIN_TOKEN"),
		PteroScript:        envOr("AUTH_PTERO_SCRIPT", defaultPteroScript),
		LauncherVersion:    envOr("AUTH_LAUNCHER_VERSION", defaultLauncherVersion),
		UpdateBaseURL:      os.Getenv("AUTH_UPDATE_BASE_URL"),
		UpdateSHA256:       os.Getenv("AUTH_UPDATE_SHA256"),
		UpdateMandatory:    os.Getenv("AUTH_UPDATE_MANDATORY") == "1",
		UpdateSigningKey:   os.Getenv("AUTH_UPDATE_SIGNING_KEY"),
		RateLimitPerMinute: defaultRateLimitPerMin,
		TrustProxy:         envBool("AUTH_TRUST_PROXY"),
		SkinStorageDir:     envOr("AUTH_SKIN_STORAGE_DIR", defaultSkinStorageDir),
		TurnstileSiteKey:   os.Getenv("AUTH_TURNSTILE_SITE_KEY"),
		TurnstileSecret:    os.Getenv("AUTH_TURNSTILE_SECRET"),
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

	if v := os.Getenv("AUTH_RATE_LIMIT_PER_MINUTE"); v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err != nil || n < minRateLimitPerMinute || n > maxRateLimitPerMinute {
			return nil, fmt.Errorf(
				"config: invalid AUTH_RATE_LIMIT_PER_MINUTE %q (must be %d-%d)",
				v, minRateLimitPerMinute, maxRateLimitPerMinute,
			)
		}
		cfg.RateLimitPerMinute = n
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

// envBool reports whether the named environment variable is set to a truthy
// value ("1", "true", "yes", "on", case-insensitive).
func envBool(name string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(name))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
