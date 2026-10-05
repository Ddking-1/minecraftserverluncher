package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// PasswordReset is a short-lived token allowing an account to reset its password.
type PasswordReset struct {
	ID        int64
	AccountID int64
	TokenHash string
	ExpiresAt time.Time
	Used      bool
}

// PasswordResetTTL is the lifetime of a reset token (30 minutes).
const PasswordResetTTL = 30 * time.Minute

// CreatePasswordReset stores a hashed reset token for an account, invalidating
// any prior token.
func (db *DB) CreatePasswordReset(accountID int64, tokenHash string) error {
	now := nowText()
	expires := time.Now().UTC().Add(PasswordResetTTL)
	tx, err := db.sql.Begin()
	if err != nil {
		return fmt.Errorf("database: reset begin: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`UPDATE password_resets SET used = 1 WHERE account_id = ?`, accountID); err != nil {
		return fmt.Errorf("database: invalidate prior resets: %w", err)
	}
	if _, err := tx.Exec(
		`INSERT INTO password_resets (account_id, token_hash, expires_at, used, created_at) VALUES (?, ?, ?, 0, ?)`,
		accountID, tokenHash, expires.UTC().Format(timeLayout), now,
	); err != nil {
		return fmt.Errorf("database: create reset: %w", err)
	}
	return tx.Commit()
}

// ConsumePasswordReset validates and consumes a reset token, returning the
// account id. It returns ErrNotFound for an unknown, used or expired token.
func (db *DB) ConsumePasswordReset(tokenHash string) (int64, error) {
	row := db.sql.QueryRow(
		`SELECT id, account_id, expires_at, used FROM password_resets WHERE token_hash = ?`,
		tokenHash,
	)
	var id, accountID int64
	var expires string
	var used int
	if err := row.Scan(&id, &accountID, &expires, &used); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrNotFound
		}
		return 0, fmt.Errorf("database: scan reset: %w", err)
	}
	if used == 1 || time.Now().UTC().After(mustParse(expires)) {
		return 0, ErrNotFound
	}
	if _, err := db.sql.Exec(`UPDATE password_resets SET used = 1 WHERE id = ?`, id); err != nil {
		return 0, fmt.Errorf("database: consume reset: %w", err)
	}
	return accountID, nil
}
