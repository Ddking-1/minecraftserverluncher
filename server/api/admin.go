// Administrative endpoints. All admin routes are gated by the X-Admin-Token
// header and are disabled entirely when no admin token is configured.
package api

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

// adminAuthorized reports whether the request carries the configured admin
// token. If no admin token is configured, admin endpoints are disabled.
func (s *Server) adminAuthorized(w http.ResponseWriter, r *http.Request) bool {
	// Legacy shared admin token (backward compatible).
	if s.cfg.AdminToken != "" {
		provided := r.Header.Get("X-Admin-Token")
		if provided != "" && subtle.ConstantTimeCompare([]byte(provided), []byte(s.cfg.AdminToken)) == 1 {
			return true
		}
	}

	// Per-user RBAC migration: a bearer token for an account that has the
	// "admin" role or the "admin" permission is also accepted. This lets
	// individual accounts act as admins without sharing the global token.
	if token, ok := bearerToken(r); ok {
		accountID, err := s.sessions.ValidateAccess(token)
		if err == nil && accountID > 0 {
			if has, _ := s.db.HasPermissionFast(accountID, "admin"); has {
				return true
			}
			// Fall back to a role named "admin".
			roles, _ := s.db.AccountRoles(accountID)
			for _, role := range roles {
				if role.Name == "admin" {
					return true
				}
			}
		}
	}

	if s.cfg.AdminToken == "" {
		http.NotFound(w, r)
		return false
	}
	writeError(w, http.StatusUnauthorized, "unauthorized")
	return false
}

type hwidRequest struct {
	DeviceID string `json:"device_id"`
}

// handleBanHWID handles POST /api/admin/hwid/ban.
func (s *Server) handleBanHWID(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(w, r) {
		return
	}
	var req hwidRequest
	if !decodeBody(w, r, &req) {
		return
	}
	deviceID := strings.TrimSpace(req.DeviceID)
	if deviceID == "" || len(deviceID) > 256 {
		writeError(w, http.StatusBadRequest, "invalid device_id")
		return
	}
	if err := s.db.BanHWID(deviceID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "banned"})
}

// handleUnbanHWID handles POST /api/admin/hwid/unban.
func (s *Server) handleUnbanHWID(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(w, r) {
		return
	}
	var req hwidRequest
	if !decodeBody(w, r, &req) {
		return
	}
	deviceID := strings.TrimSpace(req.DeviceID)
	if deviceID == "" || len(deviceID) > 256 {
		writeError(w, http.StatusBadRequest, "invalid device_id")
		return
	}
	if err := s.db.UnbanHWID(deviceID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "unbanned"})
}

// handleListHWIDBans handles GET /api/admin/hwid/bans.
func (s *Server) handleListHWIDBans(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(w, r) {
		return
	}
	bans, err := s.db.ListHWIDBans()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if bans == nil {
		bans = []string{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"bans": bans})
}
