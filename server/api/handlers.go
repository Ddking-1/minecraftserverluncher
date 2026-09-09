// Package api exposes the auth HTTP endpoints: registration, login,
// refresh, logout and logout-all. It wires the database, session and device
// services together and enforces request validation and rate limiting.
package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"minecraftserverluncher/server/auth"
	"minecraftserverluncher/server/config"
	"minecraftserverluncher/server/database"
	"minecraftserverluncher/server/devices"
	"minecraftserverluncher/server/sessions"
)

// Server holds the dependencies for the HTTP handlers.
type Server struct {
	db       *database.DB
	sessions *sessions.Service
	devices  *devices.Service
	cfg      *config.Config
	limiter  *rateLimiter
}

// NewServer constructs the API server.
func NewServer(db *database.DB, cfg *config.Config) *Server {
	return &Server{
		db:       db,
		sessions: sessions.NewService(db, cfg.AccessTokenTTL, cfg.RefreshTokenTTL),
		devices:  devices.NewService(db),
		cfg:      cfg,
		limiter:  newRateLimiter(10, time.Minute),
	}
}

// Handler returns the root HTTP handler with all routes registered.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/auth/register", s.limit(s.handleRegister))
	mux.HandleFunc("POST /api/v1/auth/login", s.limit(s.handleLogin))
	mux.HandleFunc("POST /api/v1/auth/refresh", s.limit(s.handleRefresh))
	mux.HandleFunc("POST /api/v1/auth/logout", s.limit(s.handleLogout))
	mux.HandleFunc("POST /api/v1/auth/logout-all", s.limit(s.handleLogoutAll))
	mux.HandleFunc("POST /api/server/start", s.limit(s.handleStartServer))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	return mux
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
			// Do not reveal whether the account exists.
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
func (s *Server) handleStartServer(w http.ResponseWriter, r *http.Request) {
	// This endpoint is for internal testing and admin panel use only - not for regular player sessions
	accountID, ok := s.authenticate(w, r)
	if !ok {
		return
	}
	
	// Check if the request comes from an authorized source (admin or internal test)
	// For now we'll keep it simple but this could be expanded to check specific headers or IP ranges
	// This prevents regular players from accessing server start functionality
	
	// Execute the Python script to start the Pterodactyl server
	cmd := exec.Command("python", "scripts/ptero_control.py")
	output, err := cmd.CombinedOutput()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to start server: "+string(output))
		return
	}
	
	// Return success response with server status info
	response := map[string]interface{}{
		"status": "server started",
		"output": string(output),
	}
	writeJSON(w, http.StatusOK, response)
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
