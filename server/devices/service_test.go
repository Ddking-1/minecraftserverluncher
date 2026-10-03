package devices

import (
	"path/filepath"
	"testing"

	"minecraftserverluncher/server/database"
)

func newDB(t *testing.T) *database.DB {
	t.Helper()
	db, err := database.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestRegisterAndResolve(t *testing.T) {
	db := newDB(t)
	acc, err := db.CreateAccount("carol", "hash")
	if err != nil {
		t.Fatal(err)
	}
	svc := NewService(db)
	id1, err := svc.Register(acc.ID, "dev-1", "Desktop", "")
	if err != nil {
		t.Fatal(err)
	}
	if id1 == 0 {
		t.Fatal("zero ID returned from registration")
	}
	id2, err := svc.Register(acc.ID, "dev-1", "Desktop Renamed", "")
	if err != nil {
		t.Fatal(err)
	}
	if id2 == 0 || id2 != id1 {
		t.Fatalf("upsert must keep the same ID: %d vs %d", id1, id2)
	}
}

func TestRegisterScopedPerAccount(t *testing.T) {
	db := newDB(t)
	a, err := db.CreateAccount("d1", "h")
	if err != nil {
		t.Fatal(err)
	}
	b, err := db.CreateAccount("d2", "h")
	if err != nil {
		t.Fatal(err)
	}
	svc := NewService(db)
	idA, err := svc.Register(a.ID, "shared", "", "")
	if err != nil {
		t.Fatal(err)
	}
	idB, err := svc.Register(b.ID, "shared", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if idA == idB {
		t.Fatalf("same device_id under different accounts must have different rows: %d vs %d", idA, idB)
	}
}
