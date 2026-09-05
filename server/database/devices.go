package database

import (
	"database/sql"
	"errors"
	"fmt"
)

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
