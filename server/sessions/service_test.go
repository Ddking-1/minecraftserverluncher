package sessions

import (
	"path/filepath"
	"testing"
	"time"

	"minecraftserverluncher/server/database"
)

func newDB(t *testing.T) *database.DB {
	t.Helper()
	db, err := database.Open(filepath.Join(t.TempDir(), "s.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestSessionLifecycle(t *testing.T) {
	db := newDB(t)
	acc, err := db.CreateAccount("alice", "hash")
	if err != nil {
		t.Fatal(err)
	}
	svc := NewService(db, time.Hour, 30*24*time.Hour)
	pair, err := svc.Create(acc.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatalf("empty tokens")
	}
	id, err := svc.ValidateAccess(pair.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	if id != acc.ID {
		t.Fatalf("account id %d", id)
	}
	pair2, err := svc.Refresh(pair.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if pair2.AccessToken == "" {
		t.Fatal("no new access token")
	}
	if err := svc.Logout(pair.RefreshToken); err != nil {
		t.Fatal(err)
	}
	_, err = svc.Refresh(pair.RefreshToken)
	if err != ErrInvalidRefresh {
		t.Fatalf("expected ErrInvalidRefresh, got %v", err)
	}
}

func TestValidateAccessRejectsUnknown(t *testing.T) {
	db := newDB(t)
	svc := NewService(db, time.Hour, 30*24*time.Hour)
	_, err := svc.ValidateAccess("nope")
	if err != ErrInvalidAccess {
		t.Fatalf("got %v", err)
	}
}

func TestLogoutAllRevokesAccess(t *testing.T) {
	db := newDB(t)
	acc, err := db.CreateAccount("bob", "hash")
	if err != nil {
		t.Fatal(err)
	}
	svc := NewService(db, time.Hour, 30*24*time.Hour)
	pair, err := svc.Create(acc.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.LogoutAll(acc.ID); err != nil {
		t.Fatal(err)
	}
	_, err = svc.ValidateAccess(pair.AccessToken)
	if err != ErrInvalidAccess {
		t.Fatalf("expected ErrInvalidAccess, got %v", err)
	}
}

func TestLogoutIdempotent(t *testing.T) {
	db := newDB(t)
	svc := NewService(db, time.Hour, time.Hour)
	if err := svc.Logout("unknown"); err != nil {
		t.Fatalf("got %v", err)
	}
}
