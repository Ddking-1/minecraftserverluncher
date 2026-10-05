package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// SkinUploadWindow is the per-user upload rate-limit window (3 hours).
const SkinUploadWindow = 3 * time.Hour

// SkinUploadLimit is the maximum number of uploads per user per window.
const SkinUploadLimit = 2

// AllowSkinUpload records an upload attempt for an account and reports whether
// it is within the limit (max SkinUploadLimit per SkinUploadWindow). The
// counter resets automatically once the window has elapsed.
func (db *DB) AllowSkinUpload(accountID int64) (bool, error) {
	now := time.Now().UTC()
	var count int
	var windowStart string
	err := db.sql.QueryRow(
		`SELECT count, window_start FROM skin_upload_limits WHERE account_id = ?`,
		accountID,
	).Scan(&count, &windowStart)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("database: read skin upload limit: %w", err)
		}
		// First ever upload.
		if _, err := db.sql.Exec(
			`INSERT INTO skin_upload_limits (account_id, count, window_start) VALUES (?, 1, ?)`,
			accountID, nowText(),
		); err != nil {
			return false, fmt.Errorf("database: insert skin upload limit: %w", err)
		}
		return true, nil
	}

	start := mustParse(windowStart)
	if now.Sub(start) >= SkinUploadWindow {
		// Window elapsed: reset.
		if _, err := db.sql.Exec(
			`UPDATE skin_upload_limits SET count = 1, window_start = ? WHERE account_id = ?`,
			nowText(), accountID,
		); err != nil {
			return false, fmt.Errorf("database: reset skin upload limit: %w", err)
		}
		return true, nil
	}
	if count >= SkinUploadLimit {
		return false, nil
	}
	if _, err := db.sql.Exec(
		`UPDATE skin_upload_limits SET count = count + 1 WHERE account_id = ?`,
		accountID,
	); err != nil {
		return false, fmt.Errorf("database: increment skin upload limit: %w", err)
	}
	return true, nil
}
