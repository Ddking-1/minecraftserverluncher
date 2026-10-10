package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"minecraftserverluncher/server/skins"
)

// SkinHandler serves the /api/v1/skins/{id} family of endpoints using the
// server/skins package directly (processing, validation, storage, avatar extraction).
//
// Endpoints:
//   POST   /api/v1/skins/upload                - multipart skin upload (ProcessSkin + recent-skin rotation)
//   GET    /api/v1/skins/{id}                  - raw processed PNG from storage (image/png)
//   GET    /api/v1/skins/{id}/avatar           - 64x64 avatar PNG (image/png)
//   DELETE /api/v1/skins/{id}                  - delete skin + avatar
//   POST   /api/v1/skins/{id}/active           - quick-switch active skin (rate-limited)
type SkinHandler struct {
	Storage        skins.SkinStorage
	History        *skins.ThreadSafeSkinHistory
	SkinStorageDir string
	RateLimiter    *RateLimiter
}

func (h *SkinHandler) handleUpload(w http.ResponseWriter, r *http.Request) {
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

	id := uuid.New().String()
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
		if evicted != "" && h.SkinStorageDir != "" {
			if removeErr := os.Remove(filepath.Join(h.SkinStorageDir, evicted)); removeErr != nil && !os.IsNotExist(removeErr) {
				writeError(w, http.StatusInternalServerError, "failed to delete oldest skin")
				return
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success":   true,
		"skin_id":   id,
		"avatar_url": "/api/v1/skins/" + id + "/avatar",
	})
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
	if id == "" {
		writeError(w, http.StatusNotFound, "skin not found")
		return
	}
	avatarID := id + "-avatar"

	if err := h.Storage.DeleteSkin(id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete skin")
		return
	}
	if err := h.Storage.DeleteSkin(avatarID); err != nil {
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
	if id == "" {
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
		"success": true,
		"active_skin_id": id,
	})
}
