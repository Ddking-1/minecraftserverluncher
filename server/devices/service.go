// Package devices maps client device identities to accounts. A device is
// identified by a client-supplied stable device_id scoped to the account.
// Private device keys remain on the client; only an optional public key is
// registered here (reserved for future device-bound authorization).
package devices

import (
	"minecraftserverluncher/server/database"
)

// Service registers and resolves device identities.
type Service struct {
	db *database.DB
}

// NewService returns a device service backed by the given store.
func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

// Register creates or updates a device for an account and returns its
// internal id. deviceName and publicKey may be empty.
func (s *Service) Register(accountID int64, deviceID, deviceName, publicKey string) (int64, error) {
	d, err := s.db.UpsertDevice(accountID, deviceID, deviceName, publicKey)
	if err != nil {
		return 0, err
	}
	return d.ID, nil
}
