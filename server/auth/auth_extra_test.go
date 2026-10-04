package auth

import "testing"

func TestHashTokenDeterministic(t *testing.T) {
	a := HashToken("abc")
	b := HashToken("abc")
	if a == "" || b == "" {
		t.Fatal("token hash must not be empty")
	}
	if a != b {
		t.Fatal("token hashing must be deterministic")
	}
	if a == HashToken("abcd") {
		t.Fatal("distinct inputs must produce distinct hashes")
	}
}

func TestVerifyPasswordRejectsWrong(t *testing.T) {
	hash, err := HashPassword("right-password", 10)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if err := VerifyPassword(hash, "wrong-password"); err == nil {
		t.Fatal("VerifyPassword must reject a wrong password")
	}
	if err := VerifyPassword(hash, "right-password"); err != nil {
		t.Fatalf("VerifyPassword with correct password: %v", err)
	}
}
