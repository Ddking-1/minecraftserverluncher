// Forgot-password and account-reset endpoints.
//
// Forgot never reveals whether the identifier exists (no user enumeration):
// it always returns success and only mints a reset token when the account is
// actually found. Resets are rate-limited per identifier via the shared IP
// limiter in addition to validation.
package api

import (
	"errors"
	"net/http"
	"strings"
	"time"

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
