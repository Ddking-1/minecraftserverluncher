// Proxy authorization endpoints. These are called by the Velocity proxy
// plugin (server-side), never by the client. They are gated by the admin
// token so a third party cannot use them to validate arbitrary tokens.
package api

import (
	"net/http"
	"strings"
	"time"

	"minecraftserverluncher/server/sessions"
)

type proxyAuthRequest struct {
	AccessToken string `json:"access_token"`
}

// handleProxyAuth validates an access token presented by the proxy and reports
// whether it is currently valid. It never returns the token or any account
// data; it only returns "ok" or "denied".
func (s *Server) handleProxyAuth(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(w, r) {
		return
	}
	var req proxyAuthRequest
	if !decodeBody(w, r, &req) {
		return
	}
	token := strings.TrimSpace(req.AccessToken)
	if token == "" {
		writeJSON(w, http.StatusOK, map[string]string{"status": "denied"})
		return
	}
	accountID, err := s.sessions.ValidateAccess(token)
	if err != nil {
		// Distinguish "invalid token" from a server error, but never leak
		// account info.
		if err == sessions.ErrInvalidAccess {
			writeJSON(w, http.StatusOK, map[string]string{"status": "denied"})
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	// A Minecraft-level ban is distinct from a system ban: the account may
	// still log in to the launcher, but proxy joins are rejected. Return the
	// stored reason and duration so the proxy can display them to the player.
	if info, ok := s.activeMCBan(accountID); ok {
		resp := map[string]any{"status": "denied", "source": info.Source}
		if info.Reason != "" {
			resp["reason"] = info.Reason
		}
		if info.DurationMinutes > 0 {
			resp["duration_minutes"] = info.DurationMinutes
		}
		if info.ExpiresAt != nil {
			resp["expires_at"] = info.ExpiresAt.UTC().Format(time.RFC3339)
		}
		writeJSON(w, http.StatusOK, resp)
		return
	}
	// Legacy Minecraft-ban flag (set by older admin flows).
	if banned, err := s.db.IsMinecraftBanned(accountID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	} else if banned {
		writeJSON(w, http.StatusOK, map[string]string{"status": "denied", "reason": "banned from Minecraft"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
