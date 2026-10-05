// Proxy authorization endpoints. These are called by the Velocity proxy
// plugin (server-side), never by the client. They are gated by the admin
// token so a third party cannot use them to validate arbitrary tokens.
package api

import (
	"net/http"
	"strings"

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
	_, err := s.sessions.ValidateAccess(token)
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
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
