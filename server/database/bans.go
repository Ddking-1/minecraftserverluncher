// Package database ban helpers.
package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func trimIP(ip string) string { return strings.TrimSpace(ip) }

// RecordLoginIP stores an IP address observation for an account (best effort).
func (db *DB) RecordLoginIP(accountID int64, ip string) error {
	ip = trimIP(ip)
	if ip == "" {
		return nil
	}
	_, err := db.sql.Exec(
		`INSERT INTO account_login_ips (account_id, ip, seen_at) VALUES (?, ?, ?)`,
		accountID, ip, nowText(),
	)
	if err != nil {
		return fmt.Errorf("database: record login ip: %w", err)
	}
	return nil
}

// AccountsByIP returns distinct account ids that logged in from ip within days.
func (db *DB) AccountsByIP(ip string, days int) ([]int64, error) {
	ip = trimIP(ip)
	if days <= 0 {
		days = 30
	}
	since := time.Now().UTC().AddDate(0, 0, -days).Format(timeLayout)
	rows, err := db.sql.Query(
		`SELECT DISTINCT account_id FROM account_login_ips WHERE ip = ? AND seen_at >= ?`,
		ip, since,
	)
	if err != nil {
		return nil, fmt.Errorf("database: accounts by ip: %w", err)
	}
	defer rows.Close()
	var out []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// CountAccountsByIP returns number of distinct accounts using ip within days.
func (db *DB) CountAccountsByIP(ip string, days int) (int, error) {
	ids, err := db.AccountsByIP(ip, days)
	if err != nil {
		return 0, err
	}
	return len(ids), nil
}

// DevicesByAccountSince returns devices for account with last_seen >= since.
func (db *DB) DevicesByAccountSince(accountID int64, since time.Time) ([]Device, error) {
	rows, err := db.sql.Query(
		`SELECT id, account_id, device_id, device_name, public_key, created_at, last_seen
		 FROM devices WHERE account_id = ? AND last_seen >= ?`,
		accountID, since.UTC().Format(timeLayout),
	)
	if err != nil {
		return nil, fmt.Errorf("database: devices by account since: %w", err)
	}
	defer rows.Close()
	var out []Device
	for rows.Next() {
		var d Device
		var created, lastSeen string
		if err := rows.Scan(&d.ID, &d.AccountID, &d.DeviceID, &d.DeviceName, &d.PublicKey, &created, &lastSeen); err != nil {
			return nil, err
		}
		d.CreatedAt = mustParse(created)
		d.LastSeen = mustParse(lastSeen)
		out = append(out, d)
	}
	return out, rows.Err()
}

// BanRecord model.
type BanRecord struct {
	ID              int64
	BanType         string
	Scope           string
	Target          string
	AccountID       *int64
	Reason          string
	DurationMinutes int
	ExpiresAt       *time.Time
	CreatedAt       time.Time
}

func parseBanRow(id int64, banType, scope, target string, accountID sql.NullInt64, reason string, dur int, expires, created string) BanRecord {
	r := BanRecord{ID: id, BanType: banType, Scope: scope, Target: target, Reason: reason, DurationMinutes: dur}
	if accountID.Valid {
		v := accountID.Int64
		r.AccountID = &v
	}
	if expires != "" {
		t := mustParse(expires)
		r.ExpiresAt = &t
	}
	r.CreatedAt = mustParse(created)
	return r
}

// CreateBan inserts a ban record.
func (db *DB) CreateBan(banType, scope, target string, accountID *int64, reason string, durationMinutes int) (BanRecord, error) {
	var expires *time.Time
	if durationMinutes > 0 {
		t := time.Now().UTC().Add(time.Duration(durationMinutes) * time.Minute)
		expires = &t
	}
	var expiresStr any
	if expires != nil {
		expiresStr = expires.UTC().Format(timeLayout)
	} else {
		expiresStr = nil
	}
	var acc any
	if accountID != nil {
		acc = *accountID
	} else {
		acc = nil
	}
	res, err := db.sql.Exec(
		`INSERT INTO bans (ban_type, scope, target, account_id, reason, duration_minutes, expires_at, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		banType, scope, target, acc, reason, durationMinutes, expiresStr, nowText(),
	)
	if err != nil {
		return BanRecord{}, fmt.Errorf("database: create ban: %w", err)
	}
	id, _ := res.LastInsertId()
	rec := BanRecord{ID: id, BanType: banType, Scope: scope, Target: target, AccountID: accountID, Reason: reason, DurationMinutes: durationMinutes, ExpiresAt: expires, CreatedAt: time.Now().UTC()}
	return rec, nil
}

// IsActive reports whether a ban record is still active.
func (r BanRecord) IsActive() bool {
	if r.ExpiresAt == nil {
		return true
	}
	return time.Now().UTC().Before(*r.ExpiresAt)
}

// ActiveBansFor checks bans matching type/target that are active.
func (db *DB) ActiveBansFor(banType, target string) ([]BanRecord, error) {
	rows, err := db.sql.Query(
		`SELECT id, ban_type, scope, target, account_id, reason, duration_minutes, expires_at, created_at
		 FROM bans WHERE ban_type = ? AND target = ?`,
		banType, target,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []BanRecord
	for rows.Next() {
		var id int64
		var bt, scope, tgt, reason string
		var acc sql.NullInt64
		var dur int
		var expires sql.NullString
		var created string
		if err := rows.Scan(&id, &bt, &scope, &tgt, &acc, &reason, &dur, &expires, &created); err != nil {
			return nil, err
		}
		expiresStr := ""
		if expires.Valid {
			expiresStr = expires.String
		}
		r := parseBanRow(id, bt, scope, tgt, acc, reason, dur, expiresStr, created)
		if r.IsActive() {
			out = append(out, r)
		}
	}
	return out, rows.Err()
}

// IsAccountBanned checks active account bans.
func (db *DB) IsAccountBanned(accountID int64) (bool, string, string) {
	target := fmt.Sprintf("%d", accountID)
	bans, _ := db.ActiveBansFor("account", target)
	if len(bans) > 0 {
		return true, bans[0].Reason, bans[0].Scope
	}
	return false, "", ""
}

// ActiveHWIDBans returns active hwid bans.
func (db *DB) ActiveHWIDBans(deviceID string) ([]BanRecord, error) {
	return db.ActiveBansFor("hwid", deviceID)
}

// IsIPBanned checks active ip bans.
func (db *DB) IsIPBanned(ip string) (bool, string, string) {
	bans, _ := db.ActiveBansFor("ip", trimIP(ip))
	if len(bans) > 0 {
		return true, bans[0].Reason, bans[0].Scope
	}
	return false, "", ""
}

// IsHWIDBannedBans checks active hwid bans (new table) in addition to legacy.
func (db *DB) IsHWIDBannedNew(deviceID string) (bool, string, string) {
	bans, _ := db.ActiveBansFor("hwid", deviceID)
	if len(bans) > 0 {
		return true, bans[0].Reason, bans[0].Scope
	}
	return false, "", ""
}

// UnbanByTypeTarget removes bans matching type/target.
func (db *DB) UnbanByTypeTarget(banType, target string) error {
	_, err := db.sql.Exec(`DELETE FROM bans WHERE ban_type = ? AND target = ?`, banType, target)
	if err != nil {
		return fmt.Errorf("database: unban: %w", err)
	}
	if banType == "hwid" {
		_ = db.UnbanHWID(target)
	}
	return nil
}

// UnbanAccount removes account bans.
func (db *DB) UnbanAccount(accountID int64) error {
	target := fmt.Sprintf("%d", accountID)
	return db.UnbanByTypeTarget("account", target)
}

// ListBans returns recent bans.
func (db *DB) ListBans(limit int) ([]BanRecord, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := db.sql.Query(
		`SELECT id, ban_type, scope, target, account_id, reason, duration_minutes, expires_at, created_at
		 FROM bans ORDER BY created_at DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []BanRecord
	for rows.Next() {
		var id int64
		var bt, scope, tgt, reason string
		var acc sql.NullInt64
		var dur int
		var expires sql.NullString
		var created string
		if err := rows.Scan(&id, &bt, &scope, &tgt, &acc, &reason, &dur, &expires, &created); err != nil {
			return nil, err
		}
		expiresStr := ""
		if expires.Valid {
			expiresStr = expires.String
		}
		out = append(out, parseBanRow(id, bt, scope, tgt, acc, reason, dur, expiresStr, created))
	}
	return out, rows.Err()
}
