package api

import (
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"minecraftserverluncher/server/skins"
)

// SkinHandler serves the /api/v1/skins/{id} family of endpoints using the
// server/skins package directly (validation, storage, avatar extraction).
//
// Endpoints:// GET /api/v1/skins/{id}          - raw PNG from storage (image/png)
//   GET /api/v1/skins/{id}/avatar  - 64x64 avatar PNG (image/png)
//   DELETE /api/v1/skins/{id}      - delete skin + avatar
//   POST /api/v1/skins/upload      - multipart skin upload
type SkinHandler struct {
	Storage skins.SkinStorage
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

	data, err := io.ReadAll(f)
	if err != nil {
		writeError(w, http.StatusBadRequest, "could not read skin data")
		return
	}

	if _, err := skins.ValidateSkinPNG(data, skins.MaxSkinBytes, ""); err != nil {
		writeError(w, http.StatusBadRequest, "invalid skin: "+err.Error())
		return
	}

	id := uuid.New().String()
	if err := h.Storage.SaveSkin(id, data); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to store skin")
		return
	}

	avatar, err := skins.ExtractAvatarHead(data)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to extract avatar")
		return
	}

	avatarID := id + "-avatar"
	if err := h.Storage.SaveSkin(avatarID, avatar); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to store avatar")
		return
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

	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}
