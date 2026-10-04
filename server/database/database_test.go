package database

import (
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func openTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestExpiryIndexesExist(t *testing.T) {
	db := openTestDB(t)

	want := []string{"idx_access_expires", "idx_sessions_refresh_expires"}
	for _, name := range want {
		var found string
		err := db.sql.QueryRow(
			`SELECT name FROM sqlite_master WHERE type = 'index' AND name = ?`, name,
		).Scan(&found)
		if err != nil {
			t.Fatalf("index %s not found: %v", name, err)
		}
	}
}

func TestCreateAndGetAccount(t *testing.T) {
	db := openTestDB(t)

	acc, err := db.CreateAccount("alice", "hash1")
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	if acc.ID == 0 || acc.Username != "alice" {
		t.Fatalf("unexpected account: %+v", acc)
	}

	got, err := db.AccountByUsername("alice")
	if err != nil {
		t.Fatalf("AccountByUsername: %v", err)
	}
	if got.ID != acc.ID || got.PasswordHash != "hash1" {
		t.Fatalf("unexpected fetched account: %+v", got)
	}

	if _, err := db.AccountByUsername("nobody"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestCreateAccountDuplicateUsername(t *testing.T) {
	db := openTestDB(t)
	if _, err := db.CreateAccount("bob", "h"); err != nil {
		t.Fatalf("first create: %v", err)
	}
	if _, err := db.CreateAccount("bob", "h2"); !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestDeviceUpsert(t *testing.T) {
	db := openTestDB(t)
	acc, _ := db.CreateAccount("carol", "h")

	d1, err := db.UpsertDevice(acc.ID, "dev-1", "PC", "")
	if err != nil {
		t.Fatalf("UpsertDevice: %v", err)
	}
	if d1.ID == 0 || d1.DeviceID != "dev-1" {
		t.Fatalf("unexpected device: %+v", d1)
	}

	// Upserting the same device_id returns the same row, updated.
	d2, err := db.UpsertDevice(acc.ID, "dev-1", "Laptop", "")
	if err != nil {
		t.Fatalf("UpsertDevice second: %v", err)
	}
	if d2.ID != d1.ID {
		t.Fatalf("expected same device id, got %d vs %d", d2.ID, d1.ID)
	}
	if d2.DeviceName != "Laptop" {
		t.Fatalf("expected updated name, got %q", d2.DeviceName)
	}
}

func TestSessionLifecycle(t *testing.T) {
	db := openTestDB(t)
	acc, _ := db.CreateAccount("dave", "h")

	exp := time.Now().UTC().Add(time.Hour)
	sess, err := db.CreateSession(acc.ID, 0, "refresh-hash", exp)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if sess.Revoked {
		t.Fatal("new session must not be revoked")
	}

	got, err := db.SessionByRefreshHash("refresh-hash")
	if err != nil {
		t.Fatalf("SessionByRefreshHash: %v", err)
	}
	if got.ID != sess.ID {
		t.Fatalf("unexpected session id %d", got.ID)
	}

	if err := db.CreateAccessToken(sess.ID, "access-hash", time.Now().UTC().Add(time.Minute)); err != nil {
		t.Fatalf("CreateAccessToken: %v", err)
	}
	if _, err := db.AccessTokenByHash("access-hash"); err != nil {
		t.Fatalf("AccessTokenByHash: %v", err)
	}

	if err := db.RevokeSession(sess.ID); err != nil {
		t.Fatalf("RevokeSession: %v", err)
	}
	revoked, _ := db.SessionByID(sess.ID)
	if !revoked.Revoked {
		t.Fatal("session should be revoked")
	}
	// Access tokens for the revoked session are deleted.
	if _, err := db.AccessTokenByHash("access-hash"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected access token deleted, got %v", err)
	}
}

func TestRevokeAllSessions(t *testing.T) {
	db := openTestDB(t)
	acc, _ := db.CreateAccount("erin", "h")

	for i := 0; i < 3; i++ {
		if _, err := db.CreateSession(acc.ID, 0, "rh", time.Now().UTC().Add(time.Hour)); err != nil {
			t.Fatalf("CreateSession: %v", err)
		}
	}
	if err := db.RevokeAllSessions(acc.ID); err != nil {
		t.Fatalf("RevokeAllSessions: %v", err)
	}

	// Verify all sessions for the account are revoked by re-querying.
	rows, err := db.sql.Query(`SELECT revoked FROM sessions WHERE account_id = ?`, acc.ID)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var revoked int
		if err := rows.Scan(&revoked); err != nil {
			t.Fatalf("scan: %v", err)
		}
		if revoked != 1 {
			t.Fatal("expected all sessions revoked")
		}
		count++
	}
	if count != 3 {
		t.Fatalf("expected 3 sessions, got %d", count)
	}
}
