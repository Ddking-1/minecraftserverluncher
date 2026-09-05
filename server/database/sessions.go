package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// CreateSession inserts a new session row and returns it.
func (db *DB) CreateSession(accountID, deviceID int64, refreshHash string, refreshExpiresAt time.Time) (*Session, error) {
	now := nowText()
	res, err := db.sql.Exec(
		`INSERT INTO sessions (account_id, device_id, refresh_hash, refresh_expires_at, revoked, created_at, last_used_at)
		 VALUES (?, ?, ?, ?, 0, ?, ?)`,
		accountID, nullableDevice(deviceID), refreshHash, refreshExpiresAt.UTC().Format(timeLayout), now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("database: create session: %w", err)
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
