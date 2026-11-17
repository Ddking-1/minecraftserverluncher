// Forgot-password and account-reset endpoints.
//
// Forgot never reveals whether the identifier exists (no user enumeration):
// it always returns success and only mints a reset token when the account is
// actually found. Resets are rate-limited per identifier via the shared IP
// limiter in addition to validation.
package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
	"unicode"

	"minecraftserverluncher/server/auth"
	"minecraftserverluncher/server/database"
)

const passwordResetTTL = 30 * time.Minute

type forgotRequest struct {
	Identifier string `json:"identifier"`
}

type resetRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// handleForgotPassword handles POST /api/v1/auth/forgot.
func (s *Server) handleForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotRequest
	if !decodeBody(w, r, &req) {
		return
	}
	identifier := strings.TrimSpace(req.Identifier)
	if identifier == "" {
		writeError(w, http.StatusBadRequest, "identifier is required")
		return
	}

	// Look up by username (the backend currently has no email field). Never
	// reveal whether the account exists.
	acc, err := s.db.AccountByUsername(identifier)
	if err == nil {
		token, terr := auth.NewToken()
		if terr != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		if derr := s.db.CreatePasswordReset(acc.ID, auth.HashToken(token)); derr != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		// In a real deployment the token is delivered out-of-band (e.g. email).
		// It is intentionally not returned here.
	}

	// Always the same response regardless of whether the account exists.
	writeJSON(w, http.StatusOK, map[string]string{"status": "if that account exists, a reset link was sent"})
}

// handleResetPassword handles POST /api/v1/auth/reset.
func (s *Server) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if !validPassword(req.NewPassword) {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}
	if req.Token == "" {
		writeError(w, http.StatusBadRequest, "token is required")
		return
	}

	accountID, err := s.db.ConsumePasswordReset(auth.HashToken(req.Token))
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			writeError(w, http.StatusBadRequest, "invalid or expired reset token")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
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
	// Revoke all sessions so any active tokens are invalidated.
	if err := s.sessions.LogoutAll(accountID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "password updated"})
}

// handleCapabilities reports the server version and the enabled feature flags
// so clients (e.g. the launcher) can adapt their UI without hardcoding.
func (s *Server) handleCapabilities(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"version": s.cfg.LauncherVersion,
		"features": map[string]bool{
			"invites":           s.cfg.FeatureInvites,
			"captcha":           s.cfg.FeatureCaptcha,
			"crash_reports":     s.cfg.FeatureCrashReports,
			"modpacks":          s.cfg.FeatureModpacks,
			"offline_mode":      s.cfg.FeatureOfflineMode,
			"require_device_id": s.cfg.RequireDeviceID,
		},
	})
}

// maxCrashFieldLen bounds how much of a single crash-report field is logged.
const maxCrashFieldLen = 2048

// sanitizeLogField makes client-supplied text safe to write into a single log
// line: control characters are replaced so a malicious report cannot forge
// log entries or inject terminal sequences, and the length is bounded.
func sanitizeLogField(v any) string {
	text := strings.TrimSpace(fmt.Sprint(v))
	text = strings.Map(func(r rune) rune {
		switch r {
		case '\n', '\r', '\t':
			return ' '
		}
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, text)
	// Truncate on a rune boundary so the log line stays valid UTF-8.
	runes := []rune(text)
	if len(runes) > maxCrashFieldLen {
		text = string(runes[:maxCrashFieldLen]) + "...(truncated)"
	}
	return text
}

// handleCrashReport accepts a crash report submission from the launcher.
// It is best-effort: the report is logged server-side and 202 is returned.
func (s *Server) handleCrashReport(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	if !decodeBody(w, r, &payload) {
		return
	}
	// Best-effort logging; never fail the request on log errors. Every field is
	// sanitized and truncated: client input must not shape log structure.
	if len(payload) == 0 {
		log.Printf("crash report received: (empty)")
	} else {
		for _, key := range sortedKeys(payload) {
			log.Printf("crash report %s=%s", key, sanitizeLogField(payload[key]))
		}
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "accepted"})
}

// sortedKeys returns the map keys in a deterministic order for logging.
func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// handleNews returns a small list of news items for the launcher news panel.
func (s *Server) handleNews(w http.ResponseWriter, r *http.Request) {
	items := []map[string]string{
		{"title": "Server status", "body": "All systems operational."},
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}
