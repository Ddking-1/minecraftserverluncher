// Package api exposes the auth HTTP endpoints: registration, login,
// refresh, logout and logout-all. It wires the database, session and device
// services together and enforces request validation and rate limiting.
package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"minecraftserverluncher/server/auth"
	"minecraftserverluncher/server/config"
	"minecraftserverluncher/server/database"
	"minecraftserverluncher/server/devices"
	"minecraftserverluncher/server/sessions"
	"minecraftserverluncher/server/web"
	"sync"
)

// Server holds the dependencies for the HTTP handlers.
type Server struct {
	db       *database.DB
	sessions *sessions.Service
	devices  *devices.Service
	cfg      *config.Config
	limiter  *rateLimiter
	queue    *skinQueue
	bg       []func()
	mu       sync.Mutex

	// Skin/v1 tooling wired directly to the server/skins package.
	SkinHandler        *SkinHandler
	PterodactylWebhook *PterodactylWebhook
}

// NewServer constructs the API server.
func NewServer(db *database.DB, cfg *config.Config) *Server {
	return &Server{
		db:       db,
		sessions: sessions.NewService(db, cfg.AccessTokenTTL, cfg.RefreshTokenTTL),
		devices:  devices.NewService(db),
		cfg:      cfg,
		limiter:  newRateLimiter(10, time.Minute),
		queue:    newSkinQueue(),
	}
}

// Close shuts down background workers and connections held by the server so the
// process can exit cleanly and tests do not leak goroutines or DB connections.
// It is safe to call Close more than once.
func (s *Server) Close() {
	if s.queue != nil {
		s.queue.Close()
	}
	if s.db != nil {
		s.db.Close()
	}
	if s.bg != nil {
		s.mu.Lock()
		defer s.mu.Unlock()
		for _, fn := range s.bg {
			fn()
		}
		s.bg = nil
	}
}

// enhanceSecurityHeaders wraps the handler and applies a consistent set of
// security-focused HTTP response headers on every response.
func enhanceSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

// Handler returns the root HTTP handler with all routes registered.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/auth/register", s.limit(s.handleRegister))
	mux.HandleFunc("POST /api/v1/auth/login", s.limit(s.handleLogin))
	mux.HandleFunc("POST /api/v1/auth/refresh", s.limit(s.handleRefresh))
	mux.HandleFunc("POST /api/v1/auth/logout", s.limit(s.handleLogout))
	mux.HandleFunc("POST /api/v1/auth/logout-all", s.limit(s.handleLogoutAll))
	mux.HandleFunc("POST /api/v1/auth/forgot", s.limit(s.handleForgotPassword))
	mux.HandleFunc("POST /api/v1/auth/reset", s.limit(s.handleResetPassword))
	mux.HandleFunc("GET /api/v1/account/me", s.limit(s.handleAccountMe))
	mux.HandleFunc("POST /api/v1/account/password", s.limit(s.handleChangePassword))
	mux.HandleFunc("GET /api/v1/account/sessions", s.limit(s.handleListSessions))
	mux.HandleFunc("POST /api/v1/account/sessions/revoke", s.limit(s.handleRevokeSession))
	mux.HandleFunc("POST /api/v1/account/sessions/logout-others", s.limit(s.handleLogoutOtherDevices))
	mux.HandleFunc("POST /api/server/start", s.limit(s.handleStartServer))
	mux.HandleFunc("PUT /api/v1/skins/{variant}", s.limit(s.handleUploadSkin))
	mux.HandleFunc("GET /api/v1/skins", s.limit(s.handleListSkins))
	mux.HandleFunc("GET /api/v1/skins/{account}/{variant}", s.limit(s.handleGetSkin))
	mux.HandleFunc("DELETE /api/v1/skins/{variant}", s.limit(s.handleDeleteSkin))
	mux.HandleFunc("POST /api/v1/skins/upload/{variant}", s.limit(s.handleAsyncSkinUpload))
	mux.HandleFunc("GET /api/v1/tasks/{id}", s.limit(s.handleTaskStatus))
	mux.HandleFunc("POST /api/v1/proxy/auth", s.limit(s.handleProxyAuth))
	mux.HandleFunc("POST /api/admin/hwid/ban", s.limit(s.handleBanHWID))
	mux.HandleFunc("POST /api/admin/hwid/unban", s.limit(s.handleUnbanHWID))
	mux.HandleFunc("GET /api/admin/hwid/bans", s.limit(s.handleListHWIDBans))
	mux.HandleFunc("GET /api/admin/users", s.limit(s.handleAdminListUsers))
	mux.HandleFunc("DELETE /api/admin/users/{id}", s.limit(s.handleAdminDeleteUser))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /readyz", s.handleReadyz)
	mux.HandleFunc("GET /api/v1/updates/manifest", s.limit(s.handleUpdatesManifest))
	mux.Handle("/", web.Handler())

	if s.SkinHandler != nil {
		mux.HandleFunc("POST /api/v1/skins/upload", s.SkinHandler.handleUpload)
		mux.HandleFunc("GET /api/v1/skins/{id}", s.SkinHandler.handleGetSkin)
		mux.HandleFunc("GET /api/v1/skins/{id}/avatar", s.SkinHandler.handleGetAvatar)
		mux.HandleFunc("DELETE /api/v1/skins/{id}", s.SkinHandler.handleDeleteSkin)
	}
	if s.PterodactylWebhook != nil {
		mux.HandleFunc("POST /api/v1/webhooks/pterodactyl", s.PterodactylWebhook.Handle())
	}

	return s.withSecurityHeaders(s.withRecovery(withRequestID(s.withRequestLog(mux))))
}

// setConnPoolSettings adjusts database connection pool limits so a production
// backend does not exhaust SQLite/WAL connections under concurrent load.
func setConnPoolSettings(db *database.DB) {
	if db != nil {
		db.SQL().SetMaxOpenConns(20)
		db.SQL().SetMaxIdleConns(5)
		db.SQL().SetConnMaxLifetime(time.Hour)
	}
}

// Context-aware database helpers: all query paths used by handlers go through
// these so they are always bound to the request context with a timeout.
func (s *Server) dbQueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return s.db.QueryContext(ctx, query, args...)
}

func (s *Server) dbExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return s.db.ExecContext(ctx, query, args...)
}

// ---- request/response types ----

type registerRequest struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
}

type loginRequest struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// dummyPasswordHash is a valid bcrypt hash used to keep the login response
// time constant when the account does not exist. It must never match a real
// password.
var dummyPasswordHash = func() string {
	h, err := auth.HashPassword("password-not-in-use", 12)
	if err != nil {
		panic(err)
	}
	return h
}()

// ---- handlers ----

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if !decodeBody(w, r, &req) {
		return
	}
	username := strings.TrimSpace(req.Username)
	if !validUsername(username) {
		writeError(w, http.StatusBadRequest, "invalid username")
		return
	}
	if !validPassword(req.Password) {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
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

	pair, err := s.issueSession(w, acc.ID, req.DeviceID, req.DeviceName)
	if err != nil {
		return
	}
	writeJSON(w, http.StatusCreated, pair)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if !decodeBody(w, r, &req) {
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
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Check if the device ID is banned
	isBanned, err := s.db.IsHWIDBanned(req.DeviceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if isBanned {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Account or hardware is banned"})
		return
	}

	pair, err := s.issueSession(w, acc.ID, req.DeviceID, req.DeviceName)
	if err != nil {
		return
	}
	writeJSON(w, http.StatusOK, pair)
}

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


// ---- helpers ----

// issueSession registers the device (if a device_id is supplied) and creates
// a session, writing the token response on success.
func (s *Server) issueSession(w http.ResponseWriter, accountID int64, deviceID, deviceName string) (*tokenResponse, error) {
	var devID int64
	if strings.TrimSpace(deviceID) != "" {
		id, err := s.devices.Register(accountID, strings.TrimSpace(deviceID), strings.TrimSpace(deviceName), "")
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return nil, err
		}
		devID = id
	}
	pair, err := s.sessions.Create(accountID, devID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return nil, err
	}
	return toTokenResponse(pair), nil
}

// authenticate validates the bearer access token and returns the account id.
func (s *Server) authenticate(w http.ResponseWriter, r *http.Request) (int64, bool) {
	token, ok := bearerToken(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing bearer token")
		return 0, false
	}
	accountID, err := s.sessions.ValidateAccess(token)
	if err != nil {
		if errors.Is(err, sessions.ErrInvalidAccess) {
			writeError(w, http.StatusUnauthorized, "invalid or expired access token")
			return 0, false
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return 0, false
	}
	return accountID, true
}

func bearerToken(r *http.Request) (string, bool) {
	h := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if len(h) > len(prefix) && strings.EqualFold(h[:len(prefix)], prefix) {
		return strings.TrimSpace(h[len(prefix):]), true
	}
	return "", false
}

func toTokenResponse(pair *sessions.TokenPair) *tokenResponse {
	return &tokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(pair.ExpiresIn.Seconds()),
	}
}

func validUsername(u string) bool {
	if len(u) < 3 || len(u) > 32 {
		return false
	}
	for _, c := range u {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-') {
			return false
		}
	}
	return true
}

func validPassword(p string) bool { return len(p) >= 8 }

func decodeBody(w http.ResponseWriter, r *http.Request, dst any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	if err := dec.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}
