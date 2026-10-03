package api

import (
	"errors"
	"net/http"
	"time"

	"minecraftserverluncher/server/database"
)

func (s *Server) handleAccountMe(w http.ResponseWriter, r *http.Request) {
	accountID, ok := s.authenticate(w, r)
	if !ok {
		return
	}
	acc, err := s.db.AccountByID(accountID)
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
