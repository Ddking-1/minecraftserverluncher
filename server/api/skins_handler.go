package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"minecraftserverluncher/server/skins"
)

// SkinHandler serves the /api/v1/skin-files family of endpoints using the
// server/skins package directly (processing, validation, storage, avatar extraction).
//
// The endpoints are mounted on the authenticated route group by router.go.
//
// Endpoints:
//
//	POST   /api/v1/skin-files              - multipart skin upload (ProcessSkin + recent-skin rotation)
//	GET    /api/v1/skin-files/{id}         - raw processed PNG from storage (image/png)
//	GET    /api/v1/skin-files/{id}/avatar  - 64x64 head avatar PNG (image/png)
//	DELETE /api/v1/skin-files/{id}         - delete skin + avatar
//	POST   /api/v1/skin-files/{id}/active  - quick-switch active skin (rate-limited)
type SkinHandler struct {
	Storage     *skins.SkinStorage
	History     *skins.ThreadSafeSkinHistory
	RateLimiter *RateLimiter
}

func (h *SkinHandler) handleUpload(w http.ResponseWriter, r *http.Request) {
	accountID := RequestedAccountID(r)
	if accountID == 0 {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "malformed multipart form: "+err.Error())
		return
	}
	defer r.MultipartForm.RemoveAll()

	file, ok := r.MultipartForm.File["skin"]
	if !ok || len(file) == 0 {
		writeError(w, http.StatusBadRequest, "missing or invalid skin field")
		return
	}
	f, err := file[0].Open()
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing or invalid skin field")
		return
	}
	defer f.Close()

	processed, err := skins.ProcessSkin(f, file[0].Size)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid skin: "+err.Error())
		return
	}

	// The id carries the owning account so delete/activate can be authorized
	// without a separate ownership table. Avatar ids append "-avatar" after
	// the UUID, so the owner prefix is still the first segment.
	id := strconv.FormatInt(accountID, 10) + "-" + uuid.New().String()
	if err := h.Storage.SaveSkin(id, processed); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to store skin")
		return
	}

	avatar, err := skins.ExtractAvatarHead(processed)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to extract avatar")
		return
	}

	avatarID := id + "-avatar"
	if err := h.Storage.SaveSkin(avatarID, avatar); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to store avatar")
		return
	}

	if h.History != nil {
		evicted, err := h.History.AddSkin(id, true)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to record skin history")
			return
		}
		if evicted != "" {
			// Drop the evicted skin and its avatar from storage as well, so
			// rotation does not leak files or memory entries.
			if removeErr := h.Storage.DeleteSkin(evicted); removeErr != nil {
				writeError(w, http.StatusInternalServerError, "failed to delete oldest skin")
				return
			}
			if removeErr := h.Storage.DeleteSkin(evicted + "-avatar"); removeErr != nil {
				writeError(w, http.StatusInternalServerError, "failed to delete oldest avatar")
				return
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success":    true,
		"skin_id":    id,
		"avatar_url": "/api/v1/skin-files/" + id + "/avatar",
	})
}

// skinFileOwner parses the owning account id from a skin-file id. Ids are
// created as "<accountID>-<uuid>" (avatars add "-avatar"), so the first dash
// separates the owner.
func skinFileOwner(id string) (int64, bool) {
	dash := strings.IndexByte(id, '-')
	if dash <= 0 {
		return 0, false
	}
	owner, err := strconv.ParseInt(id[:dash], 10, 64)
	if err != nil || owner <= 0 {
		return 0, false
	}
	return owner, true
}

// ownsSkinFile reports whether the authenticated account owns the skin file.
// A request without an authenticated account never owns anything.
func ownsSkinFile(r *http.Request, id string) bool {
	accountID := RequestedAccountID(r)
	if accountID == 0 {
		return false
	}
	owner, ok := skinFileOwner(id)
	return ok && owner == accountID
}

func (h *SkinHandler) handleGetSkin(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusNotFound, "skin not found")
		return
	}
	data, err := h.Storage.GetSkin(id)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			writeError(w, http.StatusNotFound, "skin not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to read skin")
		return
	}

	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "default-src 'none'")
	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (h *SkinHandler) handleGetAvatar(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusNotFound, "avatar not found")
		return
	}
	avatarID := id + "-avatar"
	data, err := h.Storage.GetSkin(avatarID)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			writeError(w, http.StatusNotFound, "avatar not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to read avatar")
		return
	}

	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "default-src 'none'")
	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (h *SkinHandler) handleDeleteSkin(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" || !ownsSkinFile(r, id) {
		writeError(w, http.StatusNotFound, "skin not found")
		return
	}
	if _, err := h.Storage.GetSkin(id); err != nil {
		writeError(w, http.StatusNotFound, "skin not found")
		return
	}

	if err := h.Storage.DeleteSkin(id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete skin")
		return
	}
	// The avatar is derived data; a missing avatar must not fail the delete.
	if err := h.Storage.DeleteSkin(id + "-avatar"); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete avatar")
		return
	}

	if h.History != nil {
		h.History.RemoveSkin(id)
	}

	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (h *SkinHandler) handleSetActiveSkin(w http.ResponseWriter, r *http.Request) {
	if h.RateLimiter != nil {
		h.RateLimiter.HandlerMiddleware(http.HandlerFunc(h.rateLimitedSetActiveSkin)).ServeHTTP(w, r)
		return
	}
	h.rateLimitedSetActiveSkin(w, r)
}

func (h *SkinHandler) rateLimitedSetActiveSkin(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" || !ownsSkinFile(r, id) {
		writeError(w, http.StatusNotFound, "skin not found")
		return
	}

	if h.History == nil {
		writeError(w, http.StatusInternalServerError, "skin history unavailable")
		return
	}

	if err := h.History.SetActive(id); err != nil {
		if strings.Contains(err.Error(), "is not in history") {
			writeError(w, http.StatusNotFound, "skin not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to set active skin")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success":        true,
		"active_skin_id": id,
	})
}
