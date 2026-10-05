package database

import (
	"fmt"
)

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
// sessions and skins via foreign keys).
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
