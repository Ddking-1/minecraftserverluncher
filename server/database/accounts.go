// Package database owns the SQLite connection, schema initialization and
// the persisted models (accounts, devices, sessions, access tokens).
//
// The schema is initialized idempotently on startup. All token digests
// stored here are SHA-256 hashes of opaque tokens; raw tokens never touch
// the database.
package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// ErrNotFound is returned when a requested row does not exist.
var ErrNotFound = errors.New("not found")

// ErrConflict is returned when a uniqueness constraint is violated.
var ErrConflict = errors.New("conflict")

// ErrForbidden is returned when an operation is not permitted for the
// current caller (for example, deleting your own admin account).
var ErrForbidden = errors.New("forbidden")

// ErrUnauthorized is returned when the caller is not authenticated.
var ErrUnauthorized = errors.New("unauthorized")

// ErrRoleNotFound is returned when a requested role name does not exist.
var ErrRoleNotFound = errors.New("role not found")

// ErrPermissionDenied is returned when the caller lacks a required permission.
var ErrPermissionDenied = errors.New("permission denied")

// HWIDBan represents a banned hardware identifier
type HWIDBan struct {
	ID        int64
	DeviceID  string
	CreatedAt time.Time
}

// Invite represents an allowlisted registration identifier (username or
// email) for invite-only beta registration.
type Invite struct {
	ID         int64
	Identifier string
	CreatedAt  time.Time
}

// Account is a registered user account.
type Account struct {
	ID           int64
	Username     string
	PasswordHash string
	CreatedAt    time.Time
	// IsAdmin (AdminRole row present) is not exposed on Account for now;
	// admin lookup uses the roles/admin helpers instead so we can keep the
	// admin table name locals-friendly.
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

// Role is an RBAC role.
type Role struct {
	ID          int64
	Name        string
	Description string
	CreatedAt   time.Time
}

// AccountRole assigns a role to an account.
type AccountRole struct {
	AccountID int64
	RoleID    int64
}

// Permission is an RBAC permission (action on a resource).
type Permission struct {
	ID          int64
	Name        string
	Description string
	CreatedAt   time.Time
}

// RolePermission assigns a permission to a role.
type RolePermission struct {
	RoleID       int64
	PermissionID int64
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
	// minimal backend. Connection pool is tuned below explicitly.
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(0)

	db := &DB{sql: sqlDB}
	if err := db.migrate(); err != nil {
		sqlDB.Close()
		return nil, err
	}
	return db, nil
}

// OpenPooled is like Open but tuned for higher-concurrency access patterns.
// It is intended for testing and for environments where multiple concurrent
// readers may be active. SQLite still has a single active writer at a time.
func OpenPooled(path string) (*DB, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(wAL)", path)
	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("database: open %s: %w", path, err)
	}
	sqlDB.SetMaxOpenConns(4)
	sqlDB.SetMaxIdleConns(4)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

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

// SQL returns the underlying database handle for advanced operations such as
// connection pool tuning. The wrapper still owns migration and model methods.
func (db *DB) SQL() *sql.DB { return db.sql }	// QueryContext executes a query with the request context.
func (db *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.sql.QueryContext(ctx, query, args...)
}

// ExecContext executes a statement with the request context.
func (db *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.sql.ExecContext(ctx, query, args...)
}

// pingContextContext is unused but present to satisfy the context import
// until it is needed by a real context-aware helper.
var _ = func() { _ = context.Background() }


// migrate applies the schema idempotently.
func (db *DB) migrate() error {
	const schema = `
CREATE TABLE IF NOT EXISTS accounts (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    username         TEXT    NOT NULL UNIQUE,
    password_hash    TEXT    NOT NULL,
    created_at       TEXT    NOT NULL,
    minecraft_banned INTEGER NOT NULL DEFAULT 0,
    approved         INTEGER NOT NULL DEFAULT 1
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

CREATE TABLE IF NOT EXISTS password_resets (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    token_hash TEXT    NOT NULL,
    expires_at TEXT    NOT NULL,
    used       INTEGER NOT NULL DEFAULT 0,
    created_at TEXT    NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_password_resets_hash ON password_resets(token_hash);

-- RBAC tables
CREATE TABLE IF NOT EXISTS roles (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT    NOT NULL UNIQUE,
    description TEXT    NOT NULL DEFAULT '',
    created_at  TEXT    NOT NULL
);

CREATE TABLE IF NOT EXISTS permissions (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT    NOT NULL UNIQUE,
    description TEXT    NOT NULL DEFAULT '',
    created_at  TEXT    NOT NULL
);

CREATE TABLE IF NOT EXISTS account_roles (
    account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    role_id    INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (account_id, role_id)
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id       INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id INTEGER NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE INDEX IF NOT EXISTS idx_account_roles_account ON account_roles(account_id);
CREATE INDEX IF NOT EXISTS idx_account_roles_role ON account_roles(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_role ON role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission ON role_permissions(permission_id);
CREATE INDEX IF NOT EXISTS idx_permissions_name ON permissions(name);
CREATE INDEX IF NOT EXISTS idx_roles_name ON roles(name);

CREATE TABLE IF NOT EXISTS invite_allowlist (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    identifier  TEXT    NOT NULL UNIQUE,
    created_at  TEXT    NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_invite_allowlist_identifier ON invite_allowlist(identifier);

CREATE TABLE IF NOT EXISTS tripwire_events (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    kind       TEXT    NOT NULL,
    detail     TEXT    NOT NULL DEFAULT '',
    ip         TEXT    NOT NULL DEFAULT '',
    account_id INTEGER NOT NULL DEFAULT 0,
    created_at TEXT    NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_tripwire_events_kind ON tripwire_events(kind);
CREATE INDEX IF NOT EXISTS idx_tripwire_events_created ON tripwire_events(created_at);

CREATE TABLE IF NOT EXISTS skin_file_approvals (
    skin_id     TEXT    PRIMARY KEY,
    approved    INTEGER NOT NULL DEFAULT 1,
    reviewer    TEXT    NOT NULL DEFAULT '',
    reviewed_at TEXT    NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS account_login_ips (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    ip         TEXT    NOT NULL,
    seen_at    TEXT    NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_login_ips_account ON account_login_ips(account_id);
CREATE INDEX IF NOT EXISTS idx_login_ips_ip ON account_login_ips(ip);

CREATE TABLE IF NOT EXISTS bans (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    ban_type         TEXT    NOT NULL,
    scope            TEXT    NOT NULL DEFAULT 'both',
    target           TEXT    NOT NULL,
    account_id       INTEGER,
    reason           TEXT    NOT NULL DEFAULT '',
    duration_minutes INTEGER NOT NULL DEFAULT 0,
    expires_at       TEXT,
    created_at       TEXT    NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_bans_type_target ON bans(ban_type, target);
CREATE INDEX IF NOT EXISTS idx_bans_account ON bans(account_id);
`
	if _, err := db.sql.Exec(schema); err != nil {
		return fmt.Errorf("database: migrate: %w", err)
	}
	// Backfill columns added after the initial schema so existing databases
	// pick them up without a manual migration.
	for _, stmt := range []string{
		`ALTER TABLE accounts ADD COLUMN minecraft_banned INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE accounts ADD COLUMN approved INTEGER NOT NULL DEFAULT 1`,
	} {
		if _, err := db.sql.Exec(stmt); err != nil {
			// Ignore "duplicate column" errors; any other error is fatal.
			if !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
				return fmt.Errorf("database: migrate alter: %w", err)
			}
		}
	}
	// Seed the default admin role + permission so per-user RBAC works out of
	// the box (idempotent).
	if _, err := db.sql.Exec(`INSERT OR IGNORE INTO roles (name, description, created_at) VALUES ('admin', 'Administrator role', ?)`, nowText()); err != nil {
		return fmt.Errorf("database: seed admin role: %w", err)
	}
	if _, err := db.sql.Exec(`INSERT OR IGNORE INTO permissions (name, description, created_at) VALUES ('admin', 'Administrator permission', ?)`, nowText()); err != nil {
		return fmt.Errorf("database: seed admin permission: %w", err)
	}
	// Link the admin permission to the admin role.
	_, _ = db.sql.Exec(`INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
		SELECT r.id, p.id FROM roles r, permissions p WHERE r.name='admin' AND p.name='admin'`)
	return nil
}

const timeLayout = time.RFC3339Nano

func nowText() string { return time.Now().UTC().Format(timeLayout) }

func parseTime(s string) (time.Time, error) {
	return time.Parse(timeLayout, s)
}

// ---------------------------------------------------------------------------
// Accounts
// ---------------------------------------------------------------------------

// CreateAccount inserts a new account. It returns ErrConflict if the
// username is already taken.
func (db *DB) CreateAccount(username, passwordHash string) (*Account, error) {
	now := nowText()
	res, err := db.sql.Exec(
		`INSERT INTO accounts (username, password_hash, created_at) VALUES (?, ?, ?)`,
		username, passwordHash, now,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrConflict
		}
		return nil, fmt.Errorf("database: create account: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("database: create account id: %w", err)
	}
	return &Account{ID: id, Username: username, PasswordHash: passwordHash, CreatedAt: mustParse(now)}, nil
}

// UpdatePassword replaces the stored password hash for an account. It returns
// ErrNotFound if no account has the given id.
func (db *DB) UpdatePassword(accountID int64, passwordHash string) error {
	res, err := db.sql.Exec(`UPDATE accounts SET password_hash = ? WHERE id = ?`, passwordHash, accountID)
	if err != nil {
		return fmt.Errorf("database: update password: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("database: update password rows: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// AccountByUsername returns the account with the given username, or
// ErrNotFound.
func (db *DB) AccountByUsername(username string) (*Account, error) {
	row := db.sql.QueryRow(
		`SELECT id, username, password_hash, created_at FROM accounts WHERE username = ?`,
		username,
	)
	return scanAccount(row)
}

// AccountByID returns the account with the given id, or ErrNotFound.
func (db *DB) AccountByID(id int64) (*Account, error) {
	row := db.sql.QueryRow(
		`SELECT id, username, password_hash, created_at FROM accounts WHERE id = ?`,
		id,
	)
	return scanAccount(row)
}

func scanAccount(row *sql.Row) (*Account, error) {
	var a Account
	var created string
	if err := row.Scan(&a.ID, &a.Username, &a.PasswordHash, &created); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("database: scan account: %w", err)
	}
	a.CreatedAt = mustParse(created)
	return &a, nil
}

func isUniqueViolation(err error) bool {
	// modernc.org/sqlite surfaces constraint errors as strings containing
	// "UNIQUE constraint failed".
	return strings.Contains(strings.ToLower(err.Error()), "unique constraint failed")
}

func mustParse(s string) time.Time {
	t, err := parseTime(s)
	if err != nil {
		// Times are written by this package in a fixed layout; a parse
		// failure indicates corruption and is surfaced as a zero time.
		return time.Time{}
	}
	return t
}

// ---------------------------------------------------------------------------
// Devices
// ---------------------------------------------------------------------------

// UpsertDevice registers or updates a device for an account and returns the
// device row. The device is identified by the client-supplied stable
// device_id scoped to the account.
func (db *DB) UpsertDevice(accountID int64, deviceID, deviceName, publicKey string) (*Device, error) {
	now := nowText()
	_, err := db.sql.Exec(
		`INSERT INTO devices (account_id, device_id, device_name, public_key, created_at, last_seen)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(account_id, device_id) DO UPDATE SET
		   device_name = excluded.device_name,
		   public_key  = excluded.public_key,
		   last_seen   = excluded.last_seen`,
		accountID, deviceID, deviceName, publicKey, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("database: upsert device: %w", err)
	}

	row := db.sql.QueryRow(
		`SELECT id, account_id, device_id, device_name, public_key, created_at, last_seen
		 FROM devices WHERE account_id = ? AND device_id = ?`,
		accountID, deviceID,
	)
	return scanDevice(row)
}

// DeviceByID returns a device by its internal id, or ErrNotFound.
func (db *DB) DeviceByID(id int64) (*Device, error) {
	row := db.sql.QueryRow(
		`SELECT id, account_id, device_id, device_name, public_key, created_at, last_seen
		 FROM devices WHERE id = ?`,
		id,
	)
	return scanDevice(row)
}

func scanDevice(row *sql.Row) (*Device, error) {
	var d Device
	var created, lastSeen string
	if err := row.Scan(&d.ID, &d.AccountID, &d.DeviceID, &d.DeviceName, &d.PublicKey, &created, &lastSeen); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("database: scan device: %w", err)
	}
	d.CreatedAt = mustParse(created)
	d.LastSeen = mustParse(lastSeen)
	return &d, nil
}

// ---------------------------------------------------------------------------
// Sessions
// ---------------------------------------------------------------------------

// CreateSession inserts a new session row and returns it.
func (db *DB) CreateSession(accountID, deviceID int64, refreshHash string, refreshExpiresAt time.Time) (*Session, error) {
	now := nowText()
	res, err2 := db.sql.Exec(
		`INSERT INTO sessions (account_id, device_id, refresh_hash, refresh_expires_at, revoked, created_at, last_used_at)
		 VALUES (?, ?, ?, ?, 0, ?, ?)`,
		accountID, nullableDevice(deviceID), refreshHash, refreshExpiresAt.UTC().Format(timeLayout), now, now,
	)
	if err2 != nil {
		return nil, fmt.Errorf("database: create session: %w", err2)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("database: create session id: %w", err)
	}
	return &Session{
		ID:               id,
		AccountID:        accountID,
		DeviceID:         deviceID,
		RefreshHash:      refreshHash,
		RefreshExpiresAt: refreshExpiresAt,
		Revoked:          false,
		CreatedAt:        mustParse(now),
		LastUsedAt:       mustParse(now),
	}, nil
}

// SessionByRefreshHash returns the session matching a refresh token digest,
// or ErrNotFound.
func (db *DB) SessionByRefreshHash(refreshHash string) (*Session, error) {
	row := db.sql.QueryRow(
		`SELECT id, account_id, device_id, refresh_hash, refresh_expires_at, revoked, created_at, last_used_at
		 FROM sessions WHERE refresh_hash = ?`,
		refreshHash,
	)
	return scanSession(row)
}

// SessionByID returns a session by id, or ErrNotFound.
func (db *DB) SessionByID(id int64) (*Session, error) {
	row := db.sql.QueryRow(
		`SELECT id, account_id, device_id, refresh_hash, refresh_expires_at, revoked, created_at, last_used_at
		 FROM sessions WHERE id = ?`,
		id,
	)
	return scanSession(row)
}

// SessionsByAccount returns every session for an account, newest first.
func (db *DB) SessionsByAccount(accountID int64) ([]Session, error) {
	rows, err := db.sql.Query(
		`SELECT id, account_id, device_id, refresh_hash, refresh_expires_at, revoked, created_at, last_used_at
		 FROM sessions WHERE account_id = ? ORDER BY created_at DESC, id DESC`,
		accountID,
	)
	if err != nil {
		return nil, fmt.Errorf("database: list sessions: %w", err)
	}
	defer rows.Close()

	var out []Session
	for rows.Next() {
		var s Session
		var deviceID sql.NullInt64
		var refreshExpires, created, lastUsed string
		var revoked int
		if err := rows.Scan(&s.ID, &s.AccountID, &deviceID, &s.RefreshHash, &refreshExpires, &revoked, &created, &lastUsed); err != nil {
			return nil, fmt.Errorf("database: scan session row: %w", err)
		}
		if deviceID.Valid {
			s.DeviceID = deviceID.Int64
		}
		s.RefreshExpiresAt = mustParse(refreshExpires)
		s.CreatedAt = mustParse(created)
		s.LastUsedAt = mustParse(lastUsed)
		s.Revoked = revoked != 0
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("database: iterate sessions: %w", err)
	}
	return out, nil
}

// RevokeSession marks a single session revoked and deletes its access tokens.
func (db *DB) RevokeSession(id int64) error {
	tx, err := db.sql.Begin()
	if err != nil {
		return fmt.Errorf("database: revoke begin: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`UPDATE sessions SET revoked = 1 WHERE id = ?`, id); err != nil {
		return fmt.Errorf("database: revoke session: %w", err)
	}
	if _, err := tx.Exec(`DELETE FROM access_tokens WHERE session_id = ?`, id); err != nil {
		return fmt.Errorf("database: revoke access tokens: %w", err)
	}
	return tx.Commit()
}

// RevokeAllSessions marks every session for an account revoked and deletes
// their access tokens.
func (db *DB) RevokeAllSessions(accountID int64) error {
	tx, err := db.sql.Begin()
	if err != nil {
		return fmt.Errorf("database: revoke-all begin: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`UPDATE sessions SET revoked = 1 WHERE account_id = ?`, accountID); err != nil {
		return fmt.Errorf("database: revoke-all sessions: %w", err)
	}
	if _, err := tx.Exec(
		`DELETE FROM access_tokens WHERE session_id IN (SELECT id FROM sessions WHERE account_id = ?)`,
		accountID,
	); err != nil {
		return fmt.Errorf("database: revoke-all access tokens: %w", err)
	}
	return tx.Commit()
}

// TouchSession updates the last_used_at timestamp of a session.
func (db *DB) TouchSession(id int64) error {
	_, err := db.sql.Exec(`UPDATE sessions SET last_used_at = ? WHERE id = ?`, nowText(), id)
	if err != nil {
		return fmt.Errorf("database: touch session: %w", err)
	}
	return nil
}

// RotateRefreshHash replaces the stored refresh-token digest for a session
// with a new digest, and updates last_used_at. Used for refresh-token rotation
// so a stolen refresh token is invalidated once used.
func (db *DB) RotateRefreshHash(sessionID int64, newRefreshHash string) error {
	res, err := db.sql.Exec(
		`UPDATE sessions SET refresh_hash = ?, last_used_at = ? WHERE id = ?`,
		newRefreshHash, nowText(), sessionID,
	)
	if err != nil {
		return fmt.Errorf("database: rotate refresh hash: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("database: rotate refresh hash rows: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// RevokeAllExcept revokes every session of an account except the one with id
// keepSessionID, and deletes their access tokens. Used by "log out from all
// other devices".
func (db *DB) RevokeAllExcept(accountID, keepSessionID int64) error {
	tx, err := db.sql.Begin()
	if err != nil {
		return fmt.Errorf("database: revoke-except begin: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		`UPDATE sessions SET revoked = 1 WHERE account_id = ? AND id != ?`,
		accountID, keepSessionID,
	); err != nil {
		return fmt.Errorf("database: revoke-except sessions: %w", err)
	}
	if _, err := tx.Exec(
		`DELETE FROM access_tokens WHERE session_id IN
		   (SELECT id FROM sessions WHERE account_id = ? AND id != ?)`,
		accountID, keepSessionID,
	); err != nil {
		return fmt.Errorf("database: revoke-except access tokens: %w", err)
	}
	return tx.Commit()
}

// CreateAccessToken stores a short-lived access token digest bound to a
// session.
func (db *DB) CreateAccessToken(sessionID int64, hash string, expiresAt time.Time) error {
	_, err := db.sql.Exec(
		`INSERT INTO access_tokens (session_id, hash, expires_at, created_at) VALUES (?, ?, ?, ?)`,
		sessionID, hash, expiresAt.UTC().Format(timeLayout), nowText(),
	)
	if err != nil {
		return fmt.Errorf("database: create access token: %w", err)
	}
	return nil
}

// AccessTokenByHash returns the access token row for a token digest, or
// ErrNotFound.
func (db *DB) AccessTokenByHash(hash string) (*AccessToken, error) {
	row := db.sql.QueryRow(
		`SELECT id, session_id, hash, expires_at, created_at FROM access_tokens WHERE hash = ?`,
		hash,
	)
	var t AccessToken
	var expires, created string
	if err := row.Scan(&t.ID, &t.SessionID, &t.Hash, &expires, &created); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("database: scan access token: %w", err)
	}
	t.ExpiresAt = mustParse(expires)
	t.CreatedAt = mustParse(created)
	return &t, nil
}

// DeleteAccessToken removes an access token (used on logout).
func (db *DB) DeleteAccessToken(hash string) error {
	if _, err := db.sql.Exec(`DELETE FROM access_tokens WHERE hash = ?`, hash); err != nil {
		return fmt.Errorf("database: delete access token: %w", err)
	}
	return nil
}

func scanSession(row *sql.Row) (*Session, error) {
	var s Session
	var deviceID sql.NullInt64
	var refreshExpires, created, lastUsed string
	var revoked int
	if err := row.Scan(&s.ID, &s.AccountID, &deviceID, &s.RefreshHash, &refreshExpires, &revoked, &created, &lastUsed); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("database: scan session: %w", err)
	}
	if deviceID.Valid {
		s.DeviceID = deviceID.Int64
	}
	s.RefreshExpiresAt = mustParse(refreshExpires)
	s.CreatedAt = mustParse(created)
	s.LastUsedAt = mustParse(lastUsed)
	s.Revoked = revoked != 0
	return &s, nil
}

func nullableDevice(id int64) any {
	if id == 0 {
		return nil
	}
	return id
}

// ---------------------------------------------------------------------------
// HWID bans
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Admin / RBAC: roles, permissions, account_roles, role_permissions
// ---------------------------------------------------------------------------

// CreateRole inserts a new role. It returns ErrConflict when the name exists.
func (db *DB) CreateRole(name, description string) (*Role, error) {
	now := nowText()
	res, err := db.sql.Exec(
		`INSERT INTO roles (name, description, created_at) VALUES (?, ?, ?)`,
		name, description, now,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrConflict
		}
		return nil, fmt.Errorf("database: create role: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("database: create role id: %w", err)
	}
	return &Role{ID: id, Name: name, Description: description, CreatedAt: mustParse(now)}, nil
}

// RoleByName returns a role by name, or ErrRoleNotFound.
func (db *DB) RoleByName(name string) (*Role, error) {
	row := db.sql.QueryRow(
		`SELECT id, name, description, created_at FROM roles WHERE name = ?`,
		name,
	)
	var r Role
	var created string
	if err := row.Scan(&r.ID, &r.Name, &r.Description, &created); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRoleNotFound
		}
		return nil, fmt.Errorf("database: scan role: %w", err)
	}
	r.CreatedAt = mustParse(created)
	return &r, nil
}

// RoleByID returns a role by id, or ErrRoleNotFound.
func (db *DB) RoleByID(id int64) (*Role, error) {
	row := db.sql.QueryRow(
		`SELECT id, name, description, created_at FROM roles WHERE id = ?`,
		id,
	)
	var r Role
	var created string
	if err := row.Scan(&r.ID, &r.Name, &r.Description, &created); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRoleNotFound
		}
		return nil, fmt.Errorf("database: scan role by id: %w", err)
	}
	r.CreatedAt = mustParse(created)
	return &r, nil
}

// ListRoles returns every role, ordered by name.
func (db *DB) ListRoles() ([]Role, error) {
	rows, err := db.sql.Query(`SELECT id, name, description, created_at FROM roles ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("database: list roles: %w", err)
	}
	defer rows.Close()

	var out []Role
	for rows.Next() {
		var r Role
		var created string
		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &created); err != nil {
			return nil, fmt.Errorf("database: scan role row: %w", err)
		}
		r.CreatedAt = mustParse(created)
		out = append(out, r)
	}
	return out, rows.Err()
}

// DeleteRole removes a role. It is a no-op (returns nil) when the role does
// not exist. RolePermissions are cascaded by the foreign key.
func (db *DB) DeleteRole(id int64) error {
	res, err := db.sql.Exec(`DELETE FROM roles WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("database: delete role: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil
	}
	return nil
}

// CreatePermission inserts a new permission. It returns ErrConflict when the
// name exists.
func (db *DB) CreatePermission(name, description string) (*Permission, error) {
	now := nowText()
	res, err := db.sql.Exec(
		`INSERT INTO permissions (name, description, created_at) VALUES (?, ?, ?)`,
		name, description, now,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrConflict
		}
		return nil, fmt.Errorf("database: create permission: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("database: create permission id: %w", err)
	}
	return &Permission{ID: id, Name: name, Description: description, CreatedAt: mustParse(now)}, nil
}

// PermissionByName returns a permission by name, or ErrNotFound.
func (db *DB) PermissionByName(name string) (*Permission, error) {
	row := db.sql.QueryRow(
		`SELECT id, name, description, created_at FROM permissions WHERE name = ?`,
		name,
	)
	var p Permission
	var created string
	if err := row.Scan(&p.ID, &p.Name, &p.Description, &created); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("database: scan permission: %w", err)
	}
	p.CreatedAt = mustParse(created)
	return &p, nil
}

// ListPermissions returns every permission, ordered by name.
func (db *DB) ListPermissions() ([]Permission, error) {
	rows, err := db.sql.Query(`SELECT id, name, description, created_at FROM permissions ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("database: list permissions: %w", err)
	}
	defer rows.Close()

	var out []Permission
	for rows.Next() {
		var p Permission
		var created string
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &created); err != nil {
			return nil, fmt.Errorf("database: scan permission row: %w", err)
		}
		p.CreatedAt = mustParse(created)
		out = append(out, p)
	}
	return out, rows.Err()
}

// DeletePermission removes a permission. It is a no-op (returns nil) when the
// permission does not exist.
func (db *DB) DeletePermission(id int64) error {
	res, err := db.sql.Exec(`DELETE FROM permissions WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("database: delete permission: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil
	}
	return nil
}

// AddRolePermission assigns a permission to a role. It is idempotent.
func (db *DB) AddRolePermission(roleID, permissionID int64) error {
	_, err := db.sql.Exec(
		`INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)
		 ON CONFLICT(role_id, permission_id) DO NOTHING`,
		roleID, permissionID,
	)
	if err != nil {
		return fmt.Errorf("database: add role permission: %w", err)
	}
	return nil
}

// RemoveRolePermission removes a permission from a role. It is a no-op when
// the assignment does not exist.
func (db *DB) RemoveRolePermission(roleID, permissionID int64) error {
	if _, err := db.sql.Exec(
		`DELETE FROM role_permissions WHERE role_id = ? AND permission_id = ?`,
		roleID, permissionID,
	); err != nil {
		return fmt.Errorf("database: remove role permission: %w", err)
	}
	return nil
}

// RolePermissions returns every permission assigned to a role.
func (db *DB) RolePermissions(roleID int64) ([]Permission, error) {
	rows, err := db.sql.Query(
		`SELECT p.id, p.name, p.description, p.created_at
		 FROM permissions p
		 JOIN role_permissions rp ON rp.permission_id = p.id
		 WHERE rp.role_id = ?
		 ORDER BY p.name`,
		roleID,
	)
	if err != nil {
		return nil, fmt.Errorf("database: role permissions: %w", err)
	}
	defer rows.Close()

	var out []Permission
	for rows.Next() {
		var p Permission
		var created string
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &created); err != nil {
			return nil, fmt.Errorf("database: scan role permission: %w", err)
		}
		p.CreatedAt = mustParse(created)
		out = append(out, p)
	}
	return out, rows.Err()
}

// AccountRoleExists reports whether an account already has a given role.
func (db *DB) AccountRoleExists(accountID, roleID int64) (bool, error) {
	var n int
	if err := db.sql.QueryRow(
		`SELECT COUNT(*) FROM account_roles WHERE account_id = ? AND role_id = ?`,
		accountID, roleID,
	).Scan(&n); err != nil {
		return false, fmt.Errorf("database: account role exists: %w", err)
	}
	return n > 0, nil
}

// AddAccountRole assigns a role to an account. It is idempotent and returns
// ErrConflict only when the referenced role or account does not exist.
func (db *DB) AddAccountRole(accountID, roleID int64) error {
	// Verify both ends exist before inserting so callers get a clear error
	// instead of a foreign-key constraint message.
	if _, err := db.AccountByID(accountID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("database: add account role account check: %w", err)
	}
	if _, err := db.RoleByID(roleID); err != nil {
		if errors.Is(err, ErrRoleNotFound) {
			return ErrRoleNotFound
		}
		return fmt.Errorf("database: add account role role check: %w", err)
	}

	if exists, err := db.AccountRoleExists(accountID, roleID); err != nil {
		return err
	} else if exists {
		return nil
	}

	if _, err := db.sql.Exec(
		`INSERT INTO account_roles (account_id, role_id) VALUES (?, ?)`,
		accountID, roleID,
	); err != nil {
		return fmt.Errorf("database: add account role: %w", err)
	}
	return nil
}

// RemoveAccountRole removes a role from an account. It is a no-op when the
// assignment does not exist.
func (db *DB) RemoveAccountRole(accountID, roleID int64) error {
	if _, err := db.sql.Exec(
		`DELETE FROM account_roles WHERE account_id = ? AND role_id = ?`,
		accountID, roleID,
	); err != nil {
		return fmt.Errorf("database: remove account role: %w", err)
	}
	return nil
}


// AccountRoles returns every role assigned to an account, ordered by name.
func (db *DB) AccountRoles(accountID int64) ([]Role, error) {
	rows, err := db.sql.Query(
		`SELECT r.id, r.name, r.description, r.created_at
		 FROM roles r
		 JOIN account_roles ar ON ar.role_id = r.id
		 WHERE ar.account_id = ?
		 ORDER BY r.name`,
		accountID,
	)
	if err != nil {
		return nil, fmt.Errorf("database: account roles: %w", err)
	}
	defer rows.Close()

	var out []Role
	for rows.Next() {
		var r Role
		var created string
		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &created); err != nil {
			return nil, fmt.Errorf("database: scan account role: %w", err)
		}
		r.CreatedAt = mustParse(created)
		out = append(out, r)
	}
	return out, rows.Err()
}

// AccountRoleNames returns the names of every role assigned to an account.
func (db *DB) AccountRoleNames(accountID int64) ([]string, error) {
	rows, err := db.sql.Query(
		`SELECT r.name FROM roles r
		 JOIN account_roles ar ON ar.role_id = r.id
		 WHERE ar.account_id = ?
		 ORDER BY r.name`,
		accountID,
	)
	if err != nil {
		return nil, fmt.Errorf("database: account role names: %w", err)
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("database: scan account role name: %w", err)
		}
		out = append(out, name)
	}
	return out, rows.Err()
}

// HasPermissionFast returns true when the account has a role that grants the
// named permission. It does not verify account existence first; use it only
// when the caller has already confirmed the account row exists.
func (db *DB) HasPermissionFast(accountID int64, permission string) (bool, error) {
	var n int
	if err := db.sql.QueryRow(
		`SELECT COUNT(*)
		 FROM permissions p
		 JOIN role_permissions rp ON rp.permission_id = p.id
		 JOIN account_roles ar ON ar.role_id = rp.role_id
		 WHERE ar.account_id = ? AND p.name = ?`,
		accountID, permission,
	).Scan(&n); err != nil {
		return false, fmt.Errorf("database: has permission: %w", err)
	}
	return n > 0, nil
}

// AdminAccounts returns every account that has the named admin role, ordered
// by id. The role must exist; otherwise ErrRoleNotFound is returned.
func (db *DB) AdminAccounts(role string) ([]Account, error) {
	if _, err := db.RoleByName(role); err != nil {
		return nil, err
	}

	rows, err := db.sql.Query(
		`SELECT a.id, a.username, a.password_hash, a.created_at
		 FROM accounts a
		 JOIN account_roles ar ON ar.account_id = a.id
		 JOIN roles r ON r.id = ar.role_id
		 WHERE r.name = ?
		 ORDER BY a.id`,
		role,
	)
	if err != nil {
		return nil, fmt.Errorf("database: admin accounts: %w", err)
	}
	defer rows.Close()

	var out []Account
	for rows.Next() {
		var a Account
		var created string
		if err := rows.Scan(&a.ID, &a.Username, &a.PasswordHash, &created); err != nil {
			return nil, fmt.Errorf("database: scan admin account: %w", err)
		}
		a.CreatedAt = mustParse(created)
		out = append(out, a)
	}
	return out, rows.Err()
}

// HasAdminRole reports whether the account has the named role.
func (db *DB) HasAdminRole(accountID int64, role string) (bool, error) {
	if _, err := db.RoleByName(role); err != nil {
		return false, err
	}
	var n int
	if err := db.sql.QueryRow(
		`SELECT COUNT(*)
		 FROM account_roles ar
		 JOIN roles r ON r.id = ar.role_id
		 WHERE ar.account_id = ? AND r.name = ?`,
		accountID, role,
	).Scan(&n); err != nil {
		return false, fmt.Errorf("database: has admin role: %w", err)
	}
	return n > 0, nil
}

// AddAdminRole assigns the named role to an account. It returns
// ErrRoleNotFound when the role does not exist.
func (db *DB) AddAdminRole(accountID int64, role string) error {
	r, err := db.RoleByName(role)
	if err != nil {
		return err
	}
	return db.AddAccountRole(accountID, r.ID)
}

// RemoveAdminRole removes the named role from an account. It returns
// ErrRoleNotFound when the role does not exist.
func (db *DB) RemoveAdminRole(accountID int64, role string) error {
	r, err := db.RoleByName(role)
	if err != nil {
		return err
	}
	return db.RemoveAccountRole(accountID, r.ID)
}

// SearchAccountIDs returns the ids of accounts whose username matches query
// (substring), ordered by id, with pagination. An empty query returns all
// accounts. It is the auth-safe sibling of SearchAccounts: callers that only
// need identifiers should use this so password hashes are never loaded.
func (db *DB) SearchAccountIDs(query string, limit, offset int) ([]int64, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := db.sql.Query(
		`SELECT id FROM accounts
		 WHERE (? = '' OR username LIKE ?) ORDER BY id LIMIT ? OFFSET ?`,
		query, "%"+query+"%", limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("database: search account ids: %w", err)
	}
	defer rows.Close()

	var out []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("database: scan account id: %w", err)
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// SearchAccounts returns accounts whose username matches query (substring),
// ordered by id, with pagination. An empty query returns all accounts.
func (db *DB) SearchAccounts(query string, limit, offset int) ([]Account, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := db.sql.Query(
		`SELECT id, username, password_hash, created_at FROM accounts
		 WHERE (? = '' OR username LIKE ?) ORDER BY id LIMIT ? OFFSET ?`,
		query, "%"+query+"%", limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("database: search accounts: %w", err)
	}
	defer rows.Close()

	var out []Account
	for rows.Next() {
		var a Account
		var created string
		if err := rows.Scan(&a.ID, &a.Username, &a.PasswordHash, &created); err != nil {
			return nil, fmt.Errorf("database: scan account row: %w", err)
		}
		a.CreatedAt = mustParse(created)
		out = append(out, a)
	}
	return out, rows.Err()
}

// CountAccounts returns the total number of accounts (for pagination).
func (db *DB) CountAccounts() (int, error) {
	var n int
	if err := db.sql.QueryRow(`SELECT COUNT(*) FROM accounts`).Scan(&n); err != nil {
		return 0, fmt.Errorf("database: count accounts: %w", err)
	}
	return n, nil
}

// DeleteAccount permanently removes an account (cascading to devices,
// sessions, skins and roles via foreign keys).
func (db *DB) DeleteAccount(id int64) error {
	res, err := db.sql.Exec(`DELETE FROM accounts WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("database: delete account: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("database: delete account rows: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteSelf is like DeleteAccount but forbids deleting your own account.
func (db *DB) DeleteSelf(accountID, currentUserID int64) error {
	if accountID == currentUserID {
		return ErrForbidden
	}
	return db.DeleteAccount(accountID)
}

// ---------------------------------------------------------------------------
// Invite allowlist (beta registration)
// ---------------------------------------------------------------------------

// AddInvite inserts an identifier into the invite allowlist. It returns
// ErrConflict when the identifier is already present.
func (db *DB) AddInvite(identifier string) (*Invite, error) {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return nil, fmt.Errorf("database: empty invite identifier")
	}
	now := nowText()
	res, err := db.sql.Exec(
		`INSERT INTO invite_allowlist (identifier, created_at) VALUES (?, ?)`,
		identifier, now,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrConflict
		}
		return nil, fmt.Errorf("database: add invite: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("database: add invite id: %w", err)
	}
	return &Invite{ID: id, Identifier: identifier, CreatedAt: mustParse(now)}, nil
}

// RemoveInvite deletes an invite by id. It is a no-op when the id does not
// exist.
func (db *DB) RemoveInvite(id int64) error {
	if _, err := db.sql.Exec(`DELETE FROM invite_allowlist WHERE id = ?`, id); err != nil {
		return fmt.Errorf("database: remove invite: %w", err)
	}
	return nil
}

// IsInvited reports whether identifier is on the allowlist.
func (db *DB) IsInvited(identifier string) (bool, error) {
	identifier = strings.TrimSpace(identifier)
	var n int
	if err := db.sql.QueryRow(
		`SELECT COUNT(*) FROM invite_allowlist WHERE identifier = ?`,
		identifier,
	).Scan(&n); err != nil {
		return false, fmt.Errorf("database: is invited: %w", err)
	}
	return n > 0, nil
}

// InviteCount returns the number of entries on the allowlist.
func (db *DB) InviteCount() (int, error) {
	var n int
	if err := db.sql.QueryRow(`SELECT COUNT(*) FROM invite_allowlist`).Scan(&n); err != nil {
		return 0, fmt.Errorf("database: invite count: %w", err)
	}
	return n, nil
}

// ListInvites returns all allowlisted identifiers, ordered by id.
func (db *DB) ListInvites() ([]Invite, error) {
	rows, err := db.sql.Query(`SELECT id, identifier, created_at FROM invite_allowlist ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("database: list invites: %w", err)
	}
	defer rows.Close()

	var out []Invite
	for rows.Next() {
		var inv Invite
		var created string
		if err := rows.Scan(&inv.ID, &inv.Identifier, &created); err != nil {
			return nil, fmt.Errorf("database: scan invite: %w", err)
		}
		inv.CreatedAt = mustParse(created)
		out = append(out, inv)
	}
	if out == nil {
		out = []Invite{}
	}
	return out, rows.Err()
}

// ---------------------------------------------------------------------------
// Minecraft-level ban (distinct from system/HWID bans)
// ---------------------------------------------------------------------------

// SetMinecraftBanned toggles the minecraft_banned flag for an account.
func (db *DB) SetMinecraftBanned(accountID int64, banned bool) error {
	val := 0
	if banned {
		val = 1
	}
	res, err := db.sql.Exec(
		`UPDATE accounts SET minecraft_banned = ? WHERE id = ?`, val, accountID,
	)
	if err != nil {
		return fmt.Errorf("database: set minecraft banned: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("database: set minecraft banned rows: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// IsMinecraftBanned reports whether an account is banned from the Minecraft
// network (proxy joins are rejected).
func (db *DB) IsMinecraftBanned(accountID int64) (bool, error) {
	var v int
	if err := db.sql.QueryRow(
		`SELECT minecraft_banned FROM accounts WHERE id = ?`, accountID,
	).Scan(&v); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, ErrNotFound
		}
		return false, fmt.Errorf("database: is minecraft banned: %w", err)
	}
	return v != 0, nil
}

// ---------------------------------------------------------------------------
// Account approval workflow
// ---------------------------------------------------------------------------

// SetAccountApproved sets the approved flag for an account.
func (db *DB) SetAccountApproved(accountID int64, approved bool) error {
	val := 0
	if approved {
		val = 1
	}
	res, err := db.sql.Exec(`UPDATE accounts SET approved = ? WHERE id = ?`, val, accountID)
	if err != nil {
		return fmt.Errorf("database: set account approved: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("database: set account approved rows: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// IsAccountApproved reports whether an account is approved (may log in).
func (db *DB) IsAccountApproved(accountID int64) (bool, error) {
	var v int
	if err := db.sql.QueryRow(`SELECT approved FROM accounts WHERE id = ?`, accountID).Scan(&v); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, ErrNotFound
		}
		return false, fmt.Errorf("database: is account approved: %w", err)
	}
	return v != 0, nil
}

// ---------------------------------------------------------------------------
// Skin file approval workflow
// ---------------------------------------------------------------------------

// SetSkinFileApproval records an admin approval decision for a skin file id.
func (db *DB) SetSkinFileApproval(skinID string, approved bool, reviewer string) error {
	val := 0
	if approved {
		val = 1
	}
	_, err := db.sql.Exec(
		`INSERT INTO skin_file_approvals (skin_id, approved, reviewer, reviewed_at)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(skin_id) DO UPDATE SET
		   approved = excluded.approved,
		   reviewer = excluded.reviewer,
		   reviewed_at = excluded.reviewed_at`,
		skinID, val, reviewer, nowText(),
	)
	if err != nil {
		return fmt.Errorf("database: set skin file approval: %w", err)
	}
	return nil
}

// IsSkinFileApproved reports whether a skin file is approved for public read.
// A missing row is treated as approved (back-compat for existing skins).
func (db *DB) IsSkinFileApproved(skinID string) (bool, error) {
	var v int
	err := db.sql.QueryRow(`SELECT approved FROM skin_file_approvals WHERE skin_id = ?`, skinID).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("database: is skin file approved: %w", err)
	}
	return v != 0, nil
}

// ---------------------------------------------------------------------------
// Tripwire events (anomaly / tamper detection)
// ---------------------------------------------------------------------------

// TripwireEvent records a suspicious client or request anomaly.
type TripwireEvent struct {
	ID        int64
	Kind      string
	Detail    string
	IP        string
	AccountID int64
	CreatedAt string
}

// RecordTripwire persists a tripwire event and returns its id.
func (db *DB) RecordTripwire(kind, detail, ip string, accountID int64) (int64, error) {
	res, err := db.sql.Exec(
		`INSERT INTO tripwire_events (kind, detail, ip, account_id, created_at) VALUES (?, ?, ?, ?, ?)`,
		kind, detail, ip, accountID, nowText(),
	)
	if err != nil {
		return 0, fmt.Errorf("database: record tripwire: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("database: tripwire id: %w", err)
	}
	return id, nil
}

// ListTripwireEvents returns the most recent tripwire events (newest first).
func (db *DB) ListTripwireEvents(limit int) ([]TripwireEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := db.sql.Query(
		`SELECT id, kind, detail, ip, account_id, created_at
		 FROM tripwire_events ORDER BY id DESC LIMIT ?`, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("database: list tripwire: %w", err)
	}
	defer rows.Close()
	var out []TripwireEvent
	for rows.Next() {
		var e TripwireEvent
		var created string
		if err := rows.Scan(&e.ID, &e.Kind, &e.Detail, &e.IP, &e.AccountID, &created); err != nil {
			return nil, fmt.Errorf("database: scan tripwire: %w", err)
		}
		e.CreatedAt = created
		out = append(out, e)
	}
	if out == nil {
		out = []TripwireEvent{}
	}
	return out, rows.Err()
}
