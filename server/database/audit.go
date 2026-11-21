package database

import (
	"fmt"
	"time"
)

// AuditEntry is a single immutable audit record.
type AuditEntry struct {
	ID        int64
	Actor     string
	Action    string
	Target    string
	Detail    string
	CreatedAt time.Time
}

// RecordAudit appends an entry to the immutable audit log. It never updates
// or deletes existing rows.
func (db *DB) RecordAudit(actor, action, target, detail string) error {
	_, err := db.sql.Exec(
		`INSERT INTO audit_log (actor, action, target, detail, created_at) VALUES (?, ?, ?, ?, ?)`,
		actor, action, target, detail, nowText(),
	)
	if err != nil {
		return fmt.Errorf("database: record audit: %w", err)
	}
	return nil
}

// ListAudit returns the most recent audit entries, newest first.
func (db *DB) ListAudit(limit int) ([]AuditEntry, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := db.sql.Query(
		`SELECT id, actor, action, target, detail, created_at FROM audit_log ORDER BY id DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("database: list audit: %w", err)
	}
	defer rows.Close()

	var out []AuditEntry
	for rows.Next() {
		var e AuditEntry
		var created string
		if err := rows.Scan(&e.ID, &e.Actor, &e.Action, &e.Target, &e.Detail, &created); err != nil {
			return nil, fmt.Errorf("database: scan audit: %w", err)
		}
		e.CreatedAt = mustParse(created)
		out = append(out, e)
	}
	if out == nil {
		out = []AuditEntry{}
	}
	return out, rows.Err()
}

// PendingSkin is a skin file awaiting admin approval.
type PendingSkin struct {
	SkinID     string
	Approved   bool
	Reviewer   string
	ReviewedAt string
}

// ListPendingSkinFiles returns skin ids that are pending admin approval.
func (db *DB) ListPendingSkinFiles() ([]PendingSkin, error) {
	rows, err := db.sql.Query(`SELECT skin_id, approved, reviewer, reviewed_at FROM skin_file_approvals WHERE approved = 0 AND (reviewer = '' OR reviewer IS NULL) ORDER BY reviewed_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("database: list pending skins: %w", err)
	}
	defer rows.Close()

	var out []PendingSkin
	for rows.Next() {
		var p PendingSkin
		var approved int
		if err := rows.Scan(&p.SkinID, &approved, &p.Reviewer, &p.ReviewedAt); err != nil {
			return nil, fmt.Errorf("database: scan pending skin: %w", err)
		}
		p.Approved = approved != 0
		out = append(out, p)
	}
	if out == nil {
		out = []PendingSkin{}
	}
	return out, rows.Err()
}

// ListRejectedSkinFiles returns skin ids that an admin rejected (approved=0
// with a reviewer set).
func (db *DB) ListRejectedSkinFiles() ([]string, error) {
	rows, err := db.sql.Query(`SELECT skin_id FROM skin_file_approvals WHERE approved = 0 AND reviewer != '' AND reviewer IS NOT NULL ORDER BY reviewed_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("database: list rejected skins: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("database: scan rejected skin: %w", err)
		}
		out = append(out, id)
	}
	if out == nil {
		out = []string{}
	}
	return out, rows.Err()
}
