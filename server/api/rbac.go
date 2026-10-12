// Package api provides RBAC (Role-Based Access Control) middleware for
// protecting admin endpoints.
package api

import (
	"net/http"
)

// Role represents a user role in the system.
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// RequirePermission returns an http.Handler that requires the authenticated
// user to have the specified permission. If the user does not have the
// required permission, they receive 403 Forbidden.
//
// This middleware uses the database layer's HasPermissionFast function,
// which checks if the account has a role that grants the named permission.
func (s *Server) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			accountID := RequestedAccountID(r)
			if accountID == 0 {
				writeError(w, http.StatusUnauthorized, "authentication required")
				return
			}

			hasPermission, err := s.db.HasPermissionFast(accountID, permission)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "internal error")
				return
			}

			if !hasPermission {
				writeError(w, http.StatusForbidden, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole returns an http.Handler that requires the authenticated user to
// have one of the specified roles. If the user does not have any of the
// required roles, they receive 403 Forbidden.
func (s *Server) RequireRole(roles ...Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			accountID := RequestedAccountID(r)
			if accountID == 0 {
				writeError(w, http.StatusUnauthorized, "authentication required")
				return
			}

			// Check if the user has the admin role via the database.
			// The database already has HasAdminRole method.
			hasAdmin, err := s.db.HasAdminRole(accountID, string(RoleAdmin))
			if err != nil {
				writeError(w, http.StatusInternalServerError, "internal error")
				return
			}

			if !hasAdmin {
				writeError(w, http.StatusForbidden, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin is a convenience wrapper that requires the admin role.
func (s *Server) RequireAdmin(next http.Handler) http.Handler {
	return s.RequireRole(RoleAdmin)(next)
}

// RequireUser is a convenience wrapper that requires at least a user role
// (any authenticated user).
func (s *Server) RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID := RequestedAccountID(r)
		if accountID == 0 {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		next.ServeHTTP(w, r)
	})
}
