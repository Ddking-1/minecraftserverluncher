package database

import (
	"testing"
	"time"
)

func TestSkinUploadRateLimit(t *testing.T) {
	db := openTestDB(t)
	acc, _ := db.CreateAccount("uploader", "h")

	// First two uploads allowed.
	for i := 0; i < SkinUploadLimit; i++ {
		ok, err := db.AllowSkinUpload(acc.ID)
		if err != nil {
			t.Fatalf("AllowSkinUpload %d: %v", i, err)
		}
		if !ok {
			t.Fatalf("upload %d should be allowed", i+1)
		}
	}
	// Third rejected.
	ok, err := db.AllowSkinUpload(acc.ID)
	if err != nil {
		t.Fatalf("AllowSkinUpload third: %v", err)
	}
	if ok {
		t.Fatal("third upload should be rejected")
	}

	// Simulate the window having elapsed.
	past := time.Now().UTC().Add(-SkinUploadWindow - time.Minute).Format(timeLayout)
	if _, err := db.sql.Exec(
		`UPDATE skin_upload_limits SET window_start = ? WHERE account_id = ?`, past, acc.ID,
	); err != nil {
		t.Fatalf("age window: %v", err)
	}
	ok, err = db.AllowSkinUpload(acc.ID)
	if err != nil {
		t.Fatalf("AllowSkinUpload after reset: %v", err)
	}
	if !ok {
		t.Fatal("upload after window reset should be allowed")
	}
}
