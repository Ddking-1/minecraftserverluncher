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
	os.Unsetenv("AUTH_TRUST_PROXY_HEADERS")
	os.Unsetenv("AUTH_TRUSTED_PROXY_CIDRS")
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
	if cfg.TrustedProxyHeaders {
		t.Errorf("expected trusted proxy headers disabled by default")
	}
	if len(cfg.TrustedProxyCIDRs) != 0 {
		t.Errorf("expected no trusted proxy cidrs by default")
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

func TestTrustedProxyRequiresCIDRs(t *testing.T) {
	clearEnv()
	os.Setenv("AUTH_TRUST_PROXY_HEADERS", "true")
	if _, err := Load(); err == nil {
		t.Fatal("expected error when trusted proxy mode enabled without cidrs")
	}
}

func TestTrustedProxyCIDRParsing(t *testing.T) {
	clearEnv()
	os.Setenv("AUTH_TRUST_PROXY_HEADERS", "true")
	os.Setenv("AUTH_TRUSTED_PROXY_CIDRS", "127.0.0.1,10.0.0.0/8")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.TrustedProxyHeaders {
		t.Fatal("expected trusted proxy mode enabled")
	}
	if len(cfg.TrustedProxyCIDRs) != 2 {
		t.Fatalf("expected 2 cidrs, got %d", len(cfg.TrustedProxyCIDRs))
	}
}
