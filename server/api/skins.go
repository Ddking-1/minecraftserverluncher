// Skin endpoints. All /api/v1/skins routes require a valid bearer access
// token. Uploads are validated (size cap, content-type allowlist) and are
// always scoped to the authenticated account.
package api

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"minecraftserverluncher/server/auth"
	"minecraftserverluncher/server/database"
)

// maxSkinBytes caps a skin upload at 1 MiB.
const maxSkinBytes = 1 << 20

// handleUploadSkin handles PUT /api/v1/skins/{variant}.
func (s *Server) handleUploadSkin(w http.ResponseWriter, r *http.Request) {
	accountID, ok := s.authenticate(w, r)
	if !ok {
		return
	}
	variant := r.PathValue("variant")
	if !validSkinVariant(variant) {
		writeError(w, http.StatusBadRequest, "invalid variant")
		return
	}

	// Enforce the per-user upload rate limit before doing any work.
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

	// Never trust the declared Content-Type: sniff the actual bytes (T-C).
	info, err := auth.SniffImage(data)
	if err != nil {
		writeError(w, http.StatusUnsupportedMediaType, "not a PNG or JPEG image")
		return
	}
	// Validate the pixel dimensions against the Minecraft skin standard (T-B).
	if !auth.ValidSkinDimensions(info.Width, info.Height) {
		writeError(w, http.StatusBadRequest, "invalid skin dimensions (must be 64x64 or 64x32)")
		return
	}
	if _, err := s.db.UpsertSkin(accountID, variant, info.ContentType, data); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	// Invalidate the per-account list cache so the next GET sees the new skin.
	s.statsCache.Delete("skins:list:" + strconv.FormatInt(accountID, 10))
	writeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

// handleListSkins handles GET /api/v1/skins (the authenticated account's skins).
func (s *Server) handleListSkins(w http.ResponseWriter, r *http.Request) {
	accountID, ok := s.authenticate(w, r)
	if !ok {
		return
	}
	cacheKey := "skins:list:" + strconv.FormatInt(accountID, 10)
	if skins, ok := s.statsCache.Get(cacheKey); ok {
		writeJSON(w, http.StatusOK, map[string]any{"skins": skins})
		return
	}
	skins, err := s.db.ListSkins(accountID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	list := make([]map[string]any, 0, len(skins))
	for _, sk := range skins {
		list = append(list, map[string]any{
			"variant":      sk.Variant,
			"content_type": sk.ContentType,
			"created_at":   sk.CreatedAt.UTC().Format(time.RFC3339),
			"updated_at":   sk.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}
	s.statsCache.Set(cacheKey, list)
	writeJSON(w, http.StatusOK, map[string]any{"skins": list})
}

// handleGetSkin handles GET /api/v1/skins/{account}/{variant}. It requires a
// bearer token but returns any account's public skin.
func (s *Server) handleGetSkin(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.authenticate(w, r); !ok {
		return
	}
	accountName := r.PathValue("account")
	variant := r.PathValue("variant")
	if !validSkinVariant(variant) {
		writeError(w, http.StatusBadRequest, "invalid variant")
		return
	}
	acc, err := s.db.AccountByUsername(accountName)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			writeError(w, http.StatusNotFound, "account not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	sk, err := s.db.SkinByAccountAndVariant(acc.ID, variant)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			writeError(w, http.StatusNotFound, "skin not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.Header().Set("Content-Type", sk.ContentType)
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(sk.Data)
}

// handleDeleteSkin handles DELETE /api/v1/skins/{variant}.
func (s *Server) handleDeleteSkin(w http.ResponseWriter, r *http.Request) {
	accountID, ok := s.authenticate(w, r)
	if !ok {
		return
	}
	variant := r.PathValue("variant")
	if !validSkinVariant(variant) {
		writeError(w, http.StatusBadRequest, "invalid variant")
		return
	}
	if err := s.db.DeleteSkin(accountID, variant); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	// Invalidate the skin list cache for this account.
	s.statsCache.Delete("skins:list:" + strconv.FormatInt(accountID, 10))
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// validSkinVariant restricts variants to lowercase letters, digits and dashes.
func validSkinVariant(v string) bool {
	if len(v) < 1 || len(v) > 32 {
		return false
	}
	for _, c := range v {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}
	return true
}
