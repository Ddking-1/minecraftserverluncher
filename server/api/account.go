package api

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"minecraftserverluncher/server/auth"
	"minecraftserverluncher/server/database"
)

// cacheKeyForSessions returns the cache key for a session list.
func cacheKeyForSessions(accountID int64) string {
	return "sessions:" + strconv.FormatInt(accountID, 10)
}

func (s *Server) handleAccountMe(w http.ResponseWriter, r *http.Request) {
	accountID, ok := s.authenticate(w, r)
	if !ok {
		return
	}
	acc, err := s.CachedAccount(accountID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			writeError(w, http.StatusNotFound, "account not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": acc.ID, "username": acc.Username, "created_at": acc.CreatedAt.UTC().Format(time.RFC3339)})
}

type passwordChangeRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (s *Server) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	var req passwordChangeRequest
	if !decodeBody(w, r, &req) {
		return
	}
	accountID, ok := s.authenticate(w, r)
	if !ok {
		return
	}
	if !validPassword(req.NewPassword) {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}
	acc, err := s.db.AccountByID(accountID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if auth.VerifyPassword(acc.PasswordHash, req.CurrentPassword) != nil {
		writeError(w, http.StatusUnauthorized, "invalid current password")
		return
	}
	hash, err := auth.HashPassword(req.NewPassword, s.cfg.BcryptCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if err := s.db.UpdatePassword(accountID, hash); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	// Invalidate the account cache since the account data may have changed.
	s.accountCache.Delete(accountID)
	if err := s.sessions.LogoutAll(accountID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	// Invalidate the session list cache since all sessions have been logged out.
	s.statsCache.Delete(cacheKeyForSessions(accountID))
	writeJSON(w, http.StatusOK, map[string]string{"status": "password changed"})
}

func (s *Server) handleListSessions(w http.ResponseWriter, r *http.Request) {
	accountID, ok := s.authenticate(w, r)
	if !ok {
		return
	}
	cacheKey := "sessions:" + strconv.FormatInt(accountID, 10)
	if sessions, ok := s.statsCache.Get(cacheKey); ok {
		writeJSON(w, http.StatusOK, map[string]any{"sessions": sessions})
		return
	}
	sess, err := s.db.SessionsByAccount(accountID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	list := make([]map[string]any, 0, len(sess))
	for _, srow := range sess {
		list = append(list, map[string]any{
			"id":                 srow.ID,
			"device_id":          srow.DeviceID,
			"revoked":            srow.Revoked,
			"created_at":         srow.CreatedAt.UTC().Format(time.RFC3339),
			"last_used_at":       srow.LastUsedAt.UTC().Format(time.RFC3339),
			"refresh_expires_at": srow.RefreshExpiresAt.UTC().Format(time.RFC3339),
		})
	}
	s.statsCache.Set(cacheKey, list)
	writeJSON(w, http.StatusOK, map[string]any{"sessions": list})
}

func (s *Server) handleRevokeSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID int64 `json:"session_id"`
	}
	if !decodeBody(w, r, &req) {
		return
	}
	accountID, ok := s.authenticate(w, r)
	if !ok {
		return
	}
	if req.SessionID == 0 {
		writeError(w, http.StatusBadRequest, "invalid session id")
		return
	}
	srow, err := s.db.SessionByID(req.SessionID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			writeError(w, http.StatusNotFound, "session not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if srow.AccountID != accountID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	if err := s.db.RevokeSession(srow.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	// Invalidate the session list cache since the session list may have changed.
	s.statsCache.Delete(cacheKeyForSessions(accountID))
	writeJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
}

// handleLogoutOtherDevices handles POST /api/v1/account/sessions/logout-others.
// It revokes every session of the account except the current one.
func (s *Server) handleLogoutOtherDevices(w http.ResponseWriter, r *http.Request) {
	accountID, ok := s.authenticate(w, r)
	if !ok {
		return
	}
	// Identify the current session from the presented access token.
	token, _ := bearerToken(r)
	sess, err := s.sessions.SessionByAccessToken(token)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid access token")
		return
	}
	if err := s.db.RevokeAllExcept(accountID, sess.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	// Invalidate the session list cache since the session list may have changed.
	s.statsCache.Delete(cacheKeyForSessions(accountID))
	writeJSON(w, http.StatusOK, map[string]string{"status": "revoked other devices"})
}
