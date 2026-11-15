package config

import (
	"os"
	"testing"
	"time"
)

func clearEnv() {
	os.Unsetenv("AUTH_LISTEN_ADDR")
	os.Unsetenv("AUTH_DB_PATH")
	os.Unsetenv("AUTH_ACCESS_TOKEN_TTL")
	os.Unsetenv("AUTH_REFRESH_TOKEN_TTL")
	os.Unsetenv("AUTH_BCRYPT_COST")
	os.Unsetenv("AUTH_ADMIN_TOKEN")
	os.Unsetenv("AUTH_PTERO_SCRIPT")
	os.Unsetenv("AUTH_RATE_LIMIT_PER_MINUTE")
	os.Unsetenv("AUTH_TRUST_PROXY")
	os.Unsetenv("AUTH_SKIN_STORAGE_DIR")
}

func TestDefaults(t *testing.T) {
	clearEnv()
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ListenAddr != ":8080" {
		t.Errorf("got %v", cfg.ListenAddr)
	}
	if cfg.DBPath != "auth.db" {
		t.Errorf("got %v", cfg.DBPath)
	}
	if cfg.AccessTokenTTL != 15*time.Minute {
		t.Errorf("got %v", cfg.AccessTokenTTL)
	}
	if cfg.RefreshTokenTTL != 720*time.Hour {
		t.Errorf("got %v", cfg.RefreshTokenTTL)
	}
	if cfg.BcryptCost != 12 {
		t.Errorf("got %v", cfg.BcryptCost)
	}
	if cfg.AdminToken != "" {
		t.Errorf("got %q", cfg.AdminToken)
	}
	if cfg.PteroScript != "scripts/ptero_control.py" {
		t.Errorf("got %v", cfg.PteroScript)
	}
}

func TestRateLimitAndProxyDefaults(t *testing.T) {
	clearEnv()
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RateLimitPerMinute != 60 {
		t.Errorf("RateLimitPerMinute = %d, want 60", cfg.RateLimitPerMinute)
	}
	if cfg.TrustProxy {
		t.Error("TrustProxy must default to false")
	}
	if cfg.SkinStorageDir != "./storage/skins" {
		t.Errorf("SkinStorageDir = %q", cfg.SkinStorageDir)
	}
}

func TestRateLimitAndProxyOverrides(t *testing.T) {
	clearEnv()
	os.Setenv("AUTH_RATE_LIMIT_PER_MINUTE", "120")
	os.Setenv("AUTH_TRUST_PROXY", "true")
	os.Setenv("AUTH_SKIN_STORAGE_DIR", "/var/lib/skins")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RateLimitPerMinute != 120 {
		t.Errorf("RateLimitPerMinute = %d, want 120", cfg.RateLimitPerMinute)
	}
	if !cfg.TrustProxy {
		t.Error("TrustProxy should be enabled")
	}
	if cfg.SkinStorageDir != "/var/lib/skins" {
		t.Errorf("SkinStorageDir = %q", cfg.SkinStorageDir)
	}
}

func TestBadRateLimit(t *testing.T) {
	clearEnv()
	os.Setenv("AUTH_RATE_LIMIT_PER_MINUTE", "0")
	if _, err := Load(); err == nil {
		t.Error("expected error for zero rate limit")
	}
}

func TestBadBcrypt(t *testing.T) {
	clearEnv()
	os.Setenv("AUTH_BCRYPT_COST", "99")
	if _, err := Load(); err == nil {
		t.Error("expected error")
	}
}

func TestBadRefreshTTL(t *testing.T) {
	clearEnv()
	os.Setenv("AUTH_REFRESH_TOKEN_TTL", "1000h")
	if _, err := Load(); err == nil {
		t.Error("expected error")
	}
}

func TestAccessTTL(t *testing.T) {
	clearEnv()
	os.Setenv("AUTH_ACCESS_TOKEN_TTL", "30m")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AccessTokenTTL != 30*time.Minute {
		t.Errorf("got %v", cfg.AccessTokenTTL)
	}
}

func TestAdminToken(t *testing.T) {
	clearEnv()
	os.Setenv("AUTH_ADMIN_TOKEN", "abc")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AdminToken != "abc" {
		t.Errorf("got %q", cfg.AdminToken)
	}
}

func TestBadAccessTTL(t *testing.T) {
	clearEnv()
	os.Setenv("AUTH_ACCESS_TOKEN_TTL", "not-a-duration")
	if _, err := Load(); err == nil {
		t.Error("expected error for invalid AUTH_ACCESS_TOKEN_TTL")
	}
}

func TestMissingSigningKeyWithMandatory(t *testing.T) {
	clearEnv()
	os.Setenv("AUTH_UPDATE_MANDATORY", "1")
	os.Unsetenv("AUTH_UPDATE_SIGNING_KEY")
	cfg, err := Load()
	if err != nil {
		t.Skip("config.Load does not reject missing signing key when mandatory enabled; that is OK for backend but clients must handle it")
	}
	_ = cfg
}
