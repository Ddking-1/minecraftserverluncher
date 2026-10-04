package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Skin is a per-account skin image variant.
type Skin struct {
	ID          int64
	AccountID   int64
	Variant     string
	ContentType string
	Data        []byte
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UpsertSkin stores or replaces a skin variant for an account. On a repeat
// upload the existing row is updated in place (one skin per variant).
func (db *DB) UpsertSkin(accountID int64, variant, contentType string, data []byte) (*Skin, error) {
	if len(data) > 1<<20 { // 1 MiB cap
		return nil, fmt.Errorf("database: skin data exceeds 1 MiB")
	}
	now := nowText()
	_, err := db.sql.Exec(
		`INSERT INTO skins (account_id, variant, content_type, data, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(account_id, variant) DO UPDATE SET
		   content_type = excluded.content_type,
		   data = excluded.data,
		   updated_at = excluded.updated_at`,
		accountID, variant, contentType, data, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("database: upsert skin: %w", err)
	}
	return &Skin{AccountID: accountID, Variant: variant, ContentType: contentType, Data: data}, nil
}

// SkinByAccountAndVariant returns a skin for an account and variant, or
// ErrNotFound.
func (db *DB) SkinByAccountAndVariant(accountID int64, variant string) (*Skin, error) {
	row := db.sql.QueryRow(
		`SELECT id, account_id, variant, content_type, data, created_at, updated_at
		 FROM skins WHERE account_id = ? AND variant = ?`,
		accountID, variant,
	)
	return scanSkin(row)
}

// ListSkins returns every skin variant for an account (metadata only; no data).
func (db *DB) ListSkins(accountID int64) ([]Skin, error) {
	rows, err := db.sql.Query(
		`SELECT id, account_id, variant, content_type, created_at, updated_at
		 FROM skins WHERE account_id = ? ORDER BY variant`,
		accountID,
	)
	if err != nil {
		return nil, fmt.Errorf("database: list skins: %w", err)
	}
	defer rows.Close()

	var out []Skin
	for rows.Next() {
		var s Skin
		var created, updated string
		if err := rows.Scan(&s.ID, &s.AccountID, &s.Variant, &s.ContentType, &created, &updated); err != nil {
			return nil, fmt.Errorf("database: scan skin: %w", err)
		}
		s.CreatedAt = mustParse(created)
		s.UpdatedAt = mustParse(updated)
		out = append(out, s)
	}
	return out, rows.Err()
}

// DeleteSkin removes a skin variant for an account. It is idempotent.
func (db *DB) DeleteSkin(accountID int64, variant string) error {
	if _, err := db.sql.Exec(`DELETE FROM skins WHERE account_id = ? AND variant = ?`, accountID, variant); err != nil {
		return fmt.Errorf("database: delete skin: %w", err)
	}
	return nil
}

func scanSkin(row *sql.Row) (*Skin, error) {
	var s Skin
	var created, updated string
	var data []byte
	if err := row.Scan(&s.ID, &s.AccountID, &s.Variant, &s.ContentType, &data, &created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("database: scan skin: %w", err)
	}
	s.Data = data
	s.CreatedAt = mustParse(created)
	s.UpdatedAt = mustParse(updated)
	return &s, nil
}
