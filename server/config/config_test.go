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
