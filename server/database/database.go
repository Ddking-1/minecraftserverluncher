// Package database owns the SQLite connection, schema initialization and
// the persisted models (accounts, devices, sessions, access tokens).
//
// The schema is initialized idempotently on startup. All token digests
// stored here are SHA-256 hashes of opaque tokens; raw tokens never touch
// the database.
package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// ErrNotFound is returned when a requested row does not exist.
var ErrNotFound = errors.New("not found")

// ErrConflict is returned when a uniqueness constraint is violated.
var ErrConflict = errors.New("conflict")

// Account is a registered user account.
type Account struct {
	ID           int64
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}

// Device is a client device associated with an account.
type Device struct {
	ID        int64
	AccountID int64
	DeviceID  string // client-generated stable device identifier
	DeviceName string
	PublicKey string // optional registered device key (reserved)
	CreatedAt time.Time
	LastSeen  time.Time
}

// Session is a login session bound to an account and (optionally) a device.
type Session struct {
	ID               int64
	AccountID        int64
	DeviceID         int64 // 0 when not associated
	RefreshHash      string
	RefreshExpiresAt time.Time
	Revoked          bool
	CreatedAt        time.Time
	LastUsedAt       time.Time
}

// AccessToken is a short-lived token bound to a session.
type AccessToken struct {
	ID        int64
	SessionID int64
	Hash      string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// DB wraps the SQLite handle.
type DB struct {
	sql *sql.DB
}

// Open opens (creating if needed) the SQLite database at path and applies
// the schema.
func Open(path string) (*DB, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)", path)
	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("database: open %s: %w", path, err)
	}
	// modernc.org/sqlite is pure Go; a single writer is expected for this
	// minimal backend.
	sqlDB.SetMaxOpenConns(1)

	db := &DB{sql: sqlDB}
	if err := db.migrate(); err != nil {
		sqlDB.Close()
		return nil, err
	}
	return db, nil
}

// Close closes the underlying database.
func (db *DB) Close() error { return db.sql.Close() }

// migrate applies the schema idempotently.
func (db *DB) migrate() error {
	const schema = `
CREATE TABLE IF NOT EXISTS accounts (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    username      TEXT    NOT NULL UNIQUE,
    password_hash TEXT    NOT NULL,
    created_at    TEXT    NOT NULL
);

CREATE TABLE IF NOT EXISTS devices (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id  INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    device_id   TEXT    NOT NULL,
    device_name TEXT    NOT NULL DEFAULT '',
    public_key  TEXT    NOT NULL DEFAULT '',
    created_at  TEXT    NOT NULL,
    last_seen   TEXT    NOT NULL,
    UNIQUE (account_id, device_id)
);

CREATE TABLE IF NOT EXISTS sessions (
    id                 INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id         INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    device_id          INTEGER REFERENCES devices(id) ON DELETE SET NULL,
    refresh_hash       TEXT    NOT NULL,
    refresh_expires_at TEXT    NOT NULL,
    revoked            INTEGER NOT NULL DEFAULT 0,
    created_at         TEXT    NOT NULL,
    last_used_at       TEXT    NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_sessions_account ON sessions(account_id);
CREATE INDEX IF NOT EXISTS idx_sessions_refresh ON sessions(refresh_hash);

CREATE TABLE IF NOT EXISTS access_tokens (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id INTEGER NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    hash       TEXT    NOT NULL UNIQUE,
    expires_at TEXT    NOT NULL,
    created_at TEXT    NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_access_session ON access_tokens(session_id);
`
	if _, err := db.sql.Exec(schema); err != nil {
		return fmt.Errorf("database: migrate: %w", err)
	}
	return nil
}

const timeLayout = time.RFC3339Nano

func nowText() string { return time.Now().UTC().Format(timeLayout) }

func parseTime(s string) (time.Time, error) {
	return time.Parse(timeLayout, s)
}
