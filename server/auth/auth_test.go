package auth

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple", 10)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "correct horse battery staple" {
		t.Fatal("hash must not equal plaintext")
	}
	if err := VerifyPassword(hash, "correct horse battery staple"); err != nil {
		t.Fatalf("VerifyPassword with correct password: %v", err)
	}
	if err := VerifyPassword(hash, "wrong password"); err == nil {
		t.Fatal("VerifyPassword with wrong password should fail")
	}
}

func TestNewTokenUniqueAndHashable(t *testing.T) {
	a, err := NewToken()
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}
	b, err := NewToken()
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}
	if a == b {
		t.Fatal("two tokens must differ")
	}
	if HashToken(a) == a {
		t.Fatal("token hash must not equal the raw token")
	}
	if HashToken(a) == HashToken(b) {
		t.Fatal("hashes of distinct tokens must differ")
	}
}
