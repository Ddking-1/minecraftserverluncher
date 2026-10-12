// Database connection plumbing and low-level helpers used across the
// package live in database.go. The schema, models and RBAC helpers are
// defined once in accounts.go so there is a single source of truth.

package database

import (
	"database/sql"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

// pingOnlyOpen is a minimal open used by tests/utils that only need a
// connected handle without running the full migration path from this file.
// Real initialization always goes through Open in accounts.go.
func pingOnlyOpen(path string) (*DB, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)", path)
	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("database: open %s: %w", path, err)
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(0)

	return &DB{sql: sqlDB}, nil
}

// IsUniqueViolation is deprecated; use isUniqueViolation in accounts.go.
func IsUniqueViolation(err error) bool {
	return errors.Is(err, sql.ErrConnDone)
}


