// Async skin upload and task polling endpoints. Uploads are validated then
// enqueued (HTTP 202 + task_id); the client polls GET /api/v1/tasks/{id}.
package api

import (
	"errors"
	"io"
	"net/http"

	"minecraftserverluncher/server/auth"
)

// handleAsyncSkinUpload handles POST /api/v1/skins/upload/{variant}. It
// validates the body like the synchronous upload, then enqueues a job and
// returns 202.
func (s *Server) handleAsyncSkinUpload(w http.ResponseWriter, r *http.Request) {
	accountID, ok := s.authenticate(w, r)
	if !ok {
		return
	}
	variant := r.PathValue("variant")
	if !validSkinVariant(variant) {
		writeError(w, http.StatusBadRequest, "invalid variant")
		return
	}

	allowed, err := s.db.AllowSkinUpload(accountID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !allowed {
		writeError(w, http.StatusTooManyRequests, "skin upload limit reached (max 2 per 3 hours)")
		return
	}

	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxSkinBytes+1))
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "skin too large")
			return
		}
		writeError(w, http.StatusBadRequest, "could not read body")
		return
	}
	if len(data) == 0 {
		writeError(w, http.StatusBadRequest, "empty skin data")
		return
	}
	// Reject non-image payloads at enqueue time.
	if _, err := auth.SniffImage(data); err != nil {
		writeError(w, http.StatusUnsupportedMediaType, "not a PNG or JPEG image")
		return
	}

	taskID, err := s.queue.enqueue(accountID, variant, data)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "skin queue is full, try again later")
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"task_id": taskID, "status": "queued"})
}

// handleTaskStatus handles GET /api/v1/tasks/{id}.
func (s *Server) handleTaskStatus(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.authenticate(w, r); !ok {
		return
	}
	id := r.PathValue("id")
	res, ok := s.queue.result(id)
	if !ok {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"task_id": res.ID,
		"status":  res.Status,
		"error":   res.Err,
	})
}
