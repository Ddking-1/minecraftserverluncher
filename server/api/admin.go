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
	if s.cfg.AdminToken == "" {
		http.NotFound(w, r)
		return false
	}
	provided := r.Header.Get("X-Admin-Token")
	if subtle.ConstantTimeCompare([]byte(provided), []byte(s.cfg.AdminToken)) != 1 {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return false
	}
	return true
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
