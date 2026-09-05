// Package auth provides password hashing and opaque token handling.
//
// Passwords are hashed with bcrypt. Tokens are random opaque values; only
// their SHA-256 digests are ever stored, so a database leak does not expose
// usable tokens. Raw passwords and raw tokens are never logged.
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// ErrInvalidCredentials is returned when a password does not match.
var ErrInvalidCredentials = errors.New("invalid credentials")

// HashPassword returns a bcrypt hash of the given password.
func HashPassword(password string, cost int) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// VerifyPassword reports whether password matches the stored bcrypt hash.
func VerifyPassword(hash, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return ErrInvalidCredentials
	}
	return nil
}

// NewToken returns a new random opaque token encoded as base64url.
// The token is 32 random bytes (256 bits of entropy).
func NewToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// HashToken returns the SHA-256 digest of a token, base64url encoded.
// Only this digest should be persisted.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
