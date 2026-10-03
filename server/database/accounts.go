package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

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
