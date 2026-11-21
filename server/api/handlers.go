// Package api exposes the auth HTTP endpoints: registration, login,
// refresh, logout and logout-all. It wires the database, session and device
// services together and enforces request validation and rate limiting.
//
// IMPORTANT: Server, NewServer, Handler, routing and middleware chains are
// defined in router.go. This file contains ONLY handler method implementations
// and helper functions that are called by the router.
package api

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"minecraftserverluncher/server/auth"
	"minecraftserverluncher/server/database"
	"minecraftserverluncher/server/sessions"
)

// The following types and functions are used by handlers but defined in
// router.go: Server, NewServer, Handler, registerRequest, loginRequest,
// refreshRequest, logoutRequest, tokenResponse, errorResponse, writeJSON,
// writeError, decodeBody, bearerToken, toTokenResponse, validUsername,
// validPassword, enhanceSecurityHeaders, limitMiddleware, limitHandler,
// setConnPoolSettings, dbQueryContext, dbExecContext, CachedAccount,
// CachedStats, issueSession, authenticate, dummyPasswordHash.

// handleRegister handles POST /api/v1/auth/register.
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if s.cfg.RequireDeviceID && strings.TrimSpace(req.DeviceID) == "" {
		writeError(w, http.StatusBadRequest, "device_id is required")
		return
	}
	username := strings.TrimSpace(req.Username)
	// Anti-bot: honeypot field must be empty.
	if strings.TrimSpace(req.Website) != "" {
		s.recordTripwire(r, "honeypot_register", username)
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	// CAPTCHA (Cloudflare Turnstile) verification when configured.
	if err := s.verifyTurnstile(req.TurnstileToken, clientIPForTripwire(r)); err != nil {
		s.recordTripwire(r, "captcha_failed_register", username)
		writeError(w, http.StatusBadRequest, "captcha verification failed")
		return
	}
	if !validUsername(username) {
		writeError(w, http.StatusBadRequest, "invalid username")
		return
	}
	if !validPassword(req.Password) {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}
	// Anti-bot: per-username rate limit.
	if s.antiBot != nil && !s.antiBot.allow("register:"+strings.ToLower(username)) {
		writeError(w, http.StatusTooManyRequests, "too many requests")
		return
	}

	// Invite-only beta registration: when an allowlist is configured, only
	// allowlisted identifiers may register. An empty allowlist keeps the
	// open-registration behaviour so development and tests keep working.
	// Fail-closed: a database error denies registration rather than silently
	// bypassing the invite gate.
	inviteCount, err := s.db.InviteCount()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if inviteCount > 0 {
		invited, err := s.db.IsInvited(username)
		if err != nil || !invited {
			writeError(w, http.StatusForbidden, "nem beta tester")
			return
		}
	}

	hash, err := auth.HashPassword(req.Password, s.cfg.BcryptCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	acc, err := s.db.CreateAccount(username, hash)
	if err != nil {
		if errors.Is(err, database.ErrConflict) {
			writeError(w, http.StatusConflict, "username already taken")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	_ = s.db.RecordLoginIP(acc.ID, s.clientIP(r))

	pair, err := s.issueSession(w, acc.ID, req.DeviceID, req.DeviceName)
	if err != nil {
		return
	}
	writeJSON(w, http.StatusCreated, pair)
}

// handleLogin handles POST /api/v1/auth/login.
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if !decodeBody(w, r, &req) {
		return
	}
	// HWID-ban hardening: when enabled, refuse requests without a device id so
	// a banned client cannot bypass the ban by omitting the identifier.
	if s.cfg.RequireDeviceID && strings.TrimSpace(req.DeviceID) == "" {
		writeError(w, http.StatusBadRequest, "device_id is required")
		return
	}
	// Anti-bot: honeypot field must be empty.
	if strings.TrimSpace(req.Website) != "" {
		s.recordTripwire(r, "honeypot_login", strings.TrimSpace(req.Username))
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	// CAPTCHA (Cloudflare Turnstile) verification when configured.
	if err := s.verifyTurnstile(req.TurnstileToken, clientIPForTripwire(r)); err != nil {
		s.recordTripwire(r, "captcha_failed_login", strings.TrimSpace(req.Username))
		writeError(w, http.StatusBadRequest, "captcha verification failed")
		return
	}
	// Anti-bot: per-username rate limit.
	if s.antiBot != nil && !s.antiBot.allow("login:"+strings.ToLower(strings.TrimSpace(req.Username))) {
		writeError(w, http.StatusTooManyRequests, "too many requests")
		return
	}
	acc, err := s.db.AccountByUsername(strings.TrimSpace(req.Username))
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// Run a bcrypt comparison against a dummy hash so the response
			// time does not reveal whether the account exists (timing-based
			// user enumeration).
			_ = auth.VerifyPassword(dummyPasswordHash, req.Password)
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if err := auth.VerifyPassword(acc.PasswordHash, req.Password); err != nil {
		s.recordTripwire(r, "invalid_credentials", strings.TrimSpace(req.Username))
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Account approval workflow: unapproved accounts cannot log in.
	// Fail-closed: a database error denies login rather than silently
	// bypassing the approval gate.
	approved, err := s.db.IsAccountApproved(acc.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !approved {
		writeError(w, http.StatusForbidden, "account pending approval")
		return
	}

	// Legacy HWID ban check.
	isBanned, err := s.db.IsHWIDBanned(req.DeviceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if isBanned {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Account or hardware is banned"})
		return
	}

	// Advanced ban checks (platform scope blocks login). Scope is matched
	// explicitly so an mc-scope record cannot mask a platform ban on the same
	// target, and so an mc-only ban does not block the launcher.
	if b, ok := s.db.ActiveBanForScope("account", strconv.FormatInt(acc.ID, 10), "platform", "both"); ok {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "account banned: " + b.Reason})
		return
	}
	if b, ok := s.db.ActiveBanForScope("hwid", req.DeviceID, "platform", "both"); ok {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "hardware banned: " + b.Reason})
		return
	}
	ip := s.clientIP(r)
	if b, ok := s.db.ActiveBanForScope("ip", ip, "platform", "both"); ok {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "ip banned: " + b.Reason})
		return
	}

	pair, err := s.issueSession(w, acc.ID, req.DeviceID, req.DeviceName)
	if err != nil {
		return
	}

	_ = s.db.RecordLoginIP(acc.ID, s.clientIP(r))

	// If the user has already accepted cookies, set a session cookie in
	// addition to the bearer token. The cookie is httpOnly so the JS UI cannot
	// read it; the UI still sends the bearer token in the Authorization header.
	if s.ConsultCookieConsent(r) {
		sessionCookie := &http.Cookie{
			Name:     "alpha_session",
			Path:     "/",
			HttpOnly: true,
			Secure:   r.TLS != nil,
			SameSite: http.SameSiteLaxMode,
			Expires:  time.Now().Add(time.Duration(pair.ExpiresIn) * time.Second),
			Value:    pair.AccessToken,
		}
		http.SetCookie(w, sessionCookie)
	}

	writeJSON(w, http.StatusOK, pair)
}

// handleRefresh handles POST /api/v1/auth/refresh.
func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if !decodeBody(w, r, &req) {
		return
	}
	pair, err := s.sessions.Refresh(req.RefreshToken)
	if err != nil {
		if errors.Is(err, sessions.ErrInvalidRefresh) {
			writeError(w, http.StatusUnauthorized, "invalid or expired refresh token")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, toTokenResponse(pair))
}

// handleLogout handles POST /api/v1/auth/logout.
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	var req logoutRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if err := s.sessions.Logout(req.RefreshToken); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleReadyz reports readiness, including database connectivity.
func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	if err := s.db.Ping(); err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

// handleLogoutAll handles POST /api/v1/auth/logout-all.
func (s *Server) handleLogoutAll(w http.ResponseWriter, r *http.Request) {
	accountID, ok := s.authenticate(w, r)
	if !ok {
		return
	}
	if err := s.sessions.LogoutAll(accountID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleStartServer handles the POST /api/server/start endpoint.
//
// This is an administrative endpoint. It is disabled entirely unless an admin
// token is configured (AUTH_ADMIN_TOKEN); requests must present it in the
// X-Admin-Token header. It never echoes command output to the client.
func (s *Server) handleStartServer(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(w, r) {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// No shell is involved: the script path is fixed by configuration and is
	// passed as a single argv element.
	cmd := exec.CommandContext(ctx, "python", s.cfg.PteroScript)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Structured operational log entry for startup failures (machine-readable).
		log.Printf("event=srv_start_failed err=%v output=%s", err, strings.TrimSpace(string(output)))
		writeError(w, http.StatusInternalServerError, "failed to start server")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "server started"})
}
