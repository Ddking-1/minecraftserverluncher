// Update manifest endpoint. Serves the current launcher version and the
// matching download metadata the client needs to self-update. This is a public
// endpoint (no auth) because the manifest is not sensitive on its own; updates
// are integrity-protected by the per-file SHA-256 in the response and by
// signature verification on the client.
package api

import (
	"net/http"
	"strings"
)

// LauncherUpdate describes a downloadable launcher update.
type LauncherUpdate struct {
	Version     string `json:"version"`
	URL         string `json:"url"`
	SHA256      string `json:"sha256"`
	Mandatory   bool   `json:"mandatory,omitempty"`
}

// handleUpdatesManifest handles GET /api/v1/updates/manifest.
// It returns the currently shipped launcher version and download details.
// The URL and SHA-256 are configuration-driven, not user-controlled.
func (s *Server) handleUpdatesManifest(w http.ResponseWriter, r *http.Request) {
	// A minimal, valid manifest. In production these values are injected via
	// configuration (not hardcoded) and the SHA-256 is computed at build time.
	upd := LauncherUpdate{
		Version:   s.cfg.LauncherVersion,
		URL:       strings.TrimRight(s.cfg.UpdateBaseURL, "/") + "/MinecraftServerLauncher-" + s.cfg.LauncherVersion + "-setup.exe",
		SHA256:    s.cfg.UpdateSHA256,
		Mandatory: s.cfg.UpdateMandatory,
	}
	writeJSON(w, http.StatusOK, map[string]any{"update": upd})
}