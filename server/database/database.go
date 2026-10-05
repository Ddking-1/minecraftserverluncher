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

// HWIDBan represents a banned hardware identifier
type HWIDBan struct {
	ID        int64
	DeviceID  string
	CreatedAt time.Time
}

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

// Ping verifies that the database is reachable.
func (db *DB) Ping() error { return db.sql.Ping() }

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

CREATE TABLE IF NOT EXISTS hwid_bans (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    device_id   TEXT    NOT NULL UNIQUE,
    created_at  TEXT    NOT NULL
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
CREATE INDEX IF NOT EXISTS idx_hwid_bans_device ON hwid_bans(device_id);

CREATE TABLE IF NOT EXISTS access_tokens (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id INTEGER NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    hash       TEXT    NOT NULL UNIQUE,
    expires_at TEXT    NOT NULL,
    created_at TEXT    NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_access_session ON access_tokens(session_id);
CREATE INDEX IF NOT EXISTS idx_access_expires ON access_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_sessions_refresh_expires ON sessions(refresh_expires_at);

CREATE TABLE IF NOT EXISTS skins (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id   INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    variant      TEXT    NOT NULL,
    content_type TEXT    NOT NULL,
    data         BLOB    NOT NULL,
    created_at   TEXT    NOT NULL,
    updated_at   TEXT    NOT NULL,
    UNIQUE (account_id, variant)
);

CREATE TABLE IF NOT EXISTS skin_upload_limits (
    account_id   INTEGER PRIMARY KEY REFERENCES accounts(id) ON DELETE CASCADE,
    count        INTEGER NOT NULL DEFAULT 0,
    window_start TEXT    NOT NULL
);
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

// IsHWIDBanned checks if a device ID is banned
func (db *DB) IsHWIDBanned(deviceID string) (bool, error) {
	row := db.sql.QueryRow(
		`SELECT COUNT(*) FROM hwid_bans WHERE device_id = ?`,
		deviceID,
	)
	var count int
	if err := row.Scan(&count); err != nil {
		return false, fmt.Errorf("database: check hwid ban: %w", err)
	}
	return count > 0, nil
}

// BanHWID adds a device identifier to the ban list. It is idempotent.
func (db *DB) BanHWID(deviceID string) error {
	_, err := db.sql.Exec(
		`INSERT INTO hwid_bans (device_id, created_at) VALUES (?, ?)
		 ON CONFLICT(device_id) DO NOTHING`,
		deviceID, nowText(),
	)
	if err != nil {
		return fmt.Errorf("database: ban hwid: %w", err)
	}
	return nil
}

// UnbanHWID removes a device identifier from the ban list.
func (db *DB) UnbanHWID(deviceID string) error {
	if _, err := db.sql.Exec(`DELETE FROM hwid_bans WHERE device_id = ?`, deviceID); err != nil {
		return fmt.Errorf("database: unban hwid: %w", err)
	}
	return nil
}

// ListHWIDBans returns all banned device identifiers, newest first.
func (db *DB) ListHWIDBans() ([]string, error) {
	rows, err := db.sql.Query(`SELECT device_id FROM hwid_bans ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("database: list hwid bans: %w", err)
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("database: scan hwid ban: %w", err)
		}
		out = append(out, id)
	}
	return out, rows.Err()
}
