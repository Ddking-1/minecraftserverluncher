package models

import (
	"time"
)

// BanType represents the type of ban
type BanType string

const (
	AccountBan BanType = "ACCOUNT"
	HWIDBan    BanType = "HWID"
	IPBan      BanType = "IP"
)

// BanRecord represents a ban record in the system
type BanRecord struct {
	ID         int64     `json:"id"`
	AccountID  string    `json:"account_id"`
	IPAddress  string    `json:"ip_address"`
	HardwareID string    `json:"hardware_id"`
	Reason     string    `json:"reason"`
	BannedAt   time.Time `json:"banned_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	BanType    BanType   `json:"ban_type"`
}

// IsBanned checks if a user is banned based on account ID, IP address, or hardware ID
func IsBanned(accID, ip, hwid string) (bool, string) {
	// This would typically query the database to check for active bans
	// For now, we'll just return false as placeholder
	
	// Implementation would involve:
	// 1. Querying ban records where:
	//    - AccountID matches and BanType is ACCOUNT 
	//    - IPAddress matches and BanType is IP
	//    - HardwareID matches and BanType is HWID
	// 2. Filter for active bans (ExpiresAt > current time)
	// 3. Return true if any match found, along with the reason
	
	return false, ""
}