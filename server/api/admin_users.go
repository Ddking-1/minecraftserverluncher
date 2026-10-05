// Account administration endpoints (user-level, beyond HWID bans). All routes
// require the admin token via adminAuthorized and are disabled when no admin
// token is configured.
package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"minecraftserverluncher/server/database"
)

// handleAdminListUsers handles GET /api/admin/users with search + pagination.
func (s *Server) handleAdminListUsers(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(w, r) {
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	accounts, err := s.db.SearchAccounts(q, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	total, err := s.db.CountAccounts()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	users := make([]map[string]any, 0, len(accounts))
	for _, acc := range accounts {
		users = append(users, map[string]any{
			"id":         acc.ID,
			"username":   acc.Username,
			"created_at": acc.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": users, "total": total})
}

// handleAdminDeleteUser handles DELETE /api/admin/users/{id}.
func (s *Server) handleAdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(w, r) {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	if err := s.db.DeleteAccount(id); err != nil {
		if errors.Is(err, database.ErrNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
