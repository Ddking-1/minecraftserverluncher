// Package api provides HTTP middleware for authentication and authorization.
package api

import (
	"context"
	"errors"
	"net/http"

	"minecraftserverluncher/server/sessions"
)

// authenticateMiddleware returns an http.Handler that requires a valid bearer
// access token and injects the authenticated account ID into the request
// context. Requests without a valid token receive 401 Unauthorized.
func (s *Server) authenticateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	accountID, err := s.validateAccessToken(r)
	if err != nil {
		if errors.Is(err, sessions.ErrInvalidAccess) {
			writeError(w, http.StatusUnauthorized, "invalid or expired access token")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	ctx := context.WithValue(r.Context(), accountIDCtxKey, accountID)
	next.ServeHTTP(w, r.WithContext(ctx))
})
}

// accountIDCtxKey is the context key used to store the authenticated account ID.
type accountIDCtxKeyType struct{}
var accountIDCtxKey = accountIDCtxKeyType{}

// RequestedAccountID returns the authenticated account ID from the request
// context, or 0 if the request has not been authenticated.
func RequestedAccountID(r *http.Request) int64 {
	if id, ok := r.Context().Value(accountIDCtxKey).(int64); ok {
		return id
	}
	return 0
}

// validateAccessToken extracts and validates the bearer token from the request,
// returning the associated account ID or an error.
func (s *Server) validateAccessToken(r *http.Request) (int64, error) {
	token, ok := bearerToken(r)
	if !ok {
		return 0, sessions.ErrInvalidAccess
	}
	accountID, err := s.sessions.ValidateAccess(token)
	if err != nil {
		return 0, err
	}
	return accountID, nil
}

// requireAuth is a helper that authenticates the request and returns the
// account ID, or false if authentication failed. Handlers can use this
// instead of embedding the middleware for endpoints that need manual control.
// Note: This is defined here for consistency, but authenticate() in router.go
// is the primary implementation used by existing handlers.
func (s *Server) requireAuth(w http.ResponseWriter, r *http.Request) (int64, bool) {
	return s.authenticate(w, r)
}

// withAuthenticatedUser is an alias for authenticateMiddleware for backwards
// compatibility with existing handler wiring.
func (s *Server) withAuthenticatedUser(next http.Handler) http.Handler {
	return s.authenticateMiddleware(next)
}
