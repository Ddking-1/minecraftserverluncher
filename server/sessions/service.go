// Package sessions implements the session lifecycle: creation, refresh,
// revocation and access-token issuance. It enforces server-side expiry and
// revocation; client-side expiry alone is never trusted.
package sessions

import (
	"errors"
	"time"

	"minecraftserverluncher/server/auth"
	"minecraftserverluncher/server/database"
)

// ErrInvalidRefresh is returned when a refresh token is unknown, expired or
// revoked.
var ErrInvalidRefresh = errors.New("invalid or expired refresh token")

// ErrInvalidAccess is returned when an access token is unknown or expired.
var ErrInvalidAccess = errors.New("invalid or expired access token")

// Service coordinates session and access-token operations against the store.
type Service struct {
	db             *database.DB
	accessTTL      time.Duration
	refreshTTL     time.Duration
}

// NewService returns a session service with the given token lifetimes.
func NewService(db *database.DB, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{db: db, accessTTL: accessTTL, refreshTTL: refreshTTL}
}

// TokenPair is the result of creating or refreshing a session.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    time.Duration // access token lifetime
}

// Create establishes a new session for an account (optionally bound to a
// device) and returns fresh access and refresh tokens.
func (s *Service) Create(accountID, deviceID int64) (*TokenPair, error) {
	refresh, err := auth.NewToken()
	if err != nil {
		return nil, err
	}
	refreshExpires := time.Now().UTC().Add(s.refreshTTL)
	if _, err := s.db.CreateSession(accountID, deviceID, auth.HashToken(refresh), refreshExpires); err != nil {
		return nil, err
	}
	return s.issueAccess(refresh)
}

// Refresh validates a refresh token and issues a new access token. The
// session is revoked if the token is unknown, expired or already revoked.
func (s *Service) Refresh(refreshToken string) (*TokenPair, error) {
	sess, err := s.db.SessionByRefreshHash(auth.HashToken(refreshToken))
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, ErrInvalidRefresh
		}
		return nil, err
	}
	if sess.Revoked || time.Now().UTC().After(sess.RefreshExpiresAt) {
		return nil, ErrInvalidRefresh
	}
	if err := s.db.TouchSession(sess.ID); err != nil {
		return nil, err
	}
	return s.issueAccess(refreshToken)
}

// issueAccess creates a short-lived access token for the session identified
// by the given refresh token.
func (s *Service) issueAccess(refreshToken string) (*TokenPair, error) {
	sess, err := s.db.SessionByRefreshHash(auth.HashToken(refreshToken))
	if err != nil {
		return nil, err
	}
	access, err := auth.NewToken()
	if err != nil {
		return nil, err
	}
	accessExpires := time.Now().UTC().Add(s.accessTTL)
	if err := s.db.CreateAccessToken(sess.ID, auth.HashToken(access), accessExpires); err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken:  access,
		RefreshToken: refreshToken,
		ExpiresIn:    s.accessTTL,
	}, nil
}

// ValidateAccess checks an access token and returns the owning account id.
// It returns ErrInvalidAccess if the token is unknown or expired.
func (s *Service) ValidateAccess(accessToken string) (int64, error) {
	tok, err := s.db.AccessTokenByHash(auth.HashToken(accessToken))
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return 0, ErrInvalidAccess
		}
		return 0, err
	}
	if time.Now().UTC().After(tok.ExpiresAt) {
		return 0, ErrInvalidAccess
	}
	sess, err := s.db.SessionByID(tok.SessionID)
	if err != nil {
		return 0, ErrInvalidAccess
	}
	if sess.Revoked {
		return 0, ErrInvalidAccess
	}
	return sess.AccountID, nil
}

// Logout revokes the session associated with the given refresh token.
func (s *Service) Logout(refreshToken string) error {
	sess, err := s.db.SessionByRefreshHash(auth.HashToken(refreshToken))
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil // idempotent: nothing to revoke
		}
		return err
	}
	return s.db.RevokeSession(sess.ID)
}

// LogoutAll revokes every session for the account.
func (s *Service) LogoutAll(accountID int64) error {
	return s.db.RevokeAllSessions(accountID)
}
