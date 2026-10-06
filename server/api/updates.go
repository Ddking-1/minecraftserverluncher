// Update manifest endpoint. Serves the current launcher version and the
// matching download metadata the client needs to self-update, plus a
// signature over the manifest so a client can verify the response was not
// tampered with in transit.
package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"
)

// LauncherUpdate describes a downloadable launcher update.
type LauncherUpdate struct {
	Version   string `json:"version"`
	URL       string `json:"url"`
	SHA256    string `json:"sha256"`
	Mandatory bool   `json:"mandatory,omitempty"`
	Signature string `json:"signature"`
}

// manifestSignature computes an HMAC-SHA256 over the canonical manifest fields
// using the update signing key. The client verifies this before trusting the
// manifest, so a tampered manifest (wrong URL/SHA) is rejected.
func (s *Server) manifestSignature(upd LauncherUpdate) string {
	if s.cfg.UpdateSigningKey == "" {
		return ""
	}
	mac := hmac.New(sha256.New, []byte(s.cfg.UpdateSigningKey))
	// Canonical string: version|url|sha256|mandatory
	canonical := strings.Join([]string{
		upd.Version, upd.URL, upd.SHA256,
	}, "|")
	if upd.Mandatory {
		canonical += "|1"
	} else {
		canonical += "|0"
	}
	mac.Write([]byte(canonical))
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyManifestSignature checks an HMAC signature over the given update fields
// with the provided signing key. Returns true only on a constant-time match.
func VerifyManifestSignature(key, version, url, sha256Hex string, mandatory bool, signature string) bool {
	if key == "" || signature == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(key))
	canonical := strings.Join([]string{version, url, sha256Hex}, "|")
	if mandatory {
		canonical += "|1"
	} else {
		canonical += "|0"
	}
	mac.Write([]byte(canonical))
	expected := hex.EncodeToString(mac.Sum(nil))
	return subtle.ConstantTimeCompare([]byte(expected), []byte(signature)) == 1
}

// handleUpdatesManifest handles GET /api/v1/updates/manifest.
// It is rate-limited (see Handler() route registration) and returns the
// shipped launcher version, download URL, SHA-256, mandatory flag and an
// HMAC signature over those fields.
func (s *Server) handleUpdatesManifest(w http.ResponseWriter, r *http.Request) {
	upd := LauncherUpdate{
		Version:   s.cfg.LauncherVersion,
		URL:       strings.TrimRight(s.cfg.UpdateBaseURL, "/") + "/MinecraftServerLauncher-" + s.cfg.LauncherVersion + "-setup.exe",
		SHA256:    s.cfg.UpdateSHA256,
		Mandatory: s.cfg.UpdateMandatory,
	}
	upd.Signature = s.manifestSignature(upd)
	writeJSON(w, http.StatusOK, map[string]any{"update": upd})
}
