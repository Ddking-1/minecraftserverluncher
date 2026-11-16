// Package api exposes the auth HTTP endpoints.
//
// Route groups and middleware chains are defined here; handler implementations
// live in handlers.go, account.go, admin.go, admin_users.go, skins.go,
// skins_handler.go and queue_handlers.go.
//
// This file is the SINGLE source of truth for Server, NewServer, Handler and
// all routing/middleware wiring.
package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"minecraftserverluncher/server/auth"
	"minecraftserverluncher/server/cache"
	"minecraftserverluncher/server/config"
	"minecraftserverluncher/server/database"
	"minecraftserverluncher/server/devices"
	"minecraftserverluncher/server/sessions"
	"minecraftserverluncher/server/web"
)

// Server holds the dependencies for the HTTP handlers.
type Server struct {
	db       *database.DB
	sessions *sessions.Service
	devices  *devices.Service
	cfg      *config.Config
	limiter  *RateLimiter
	queue    *skinQueue
	bg       []func()
	mu       sync.Mutex

	// Skin/v1 tooling wired directly to the server/skins package.
	SkinHandler        *SkinHandler
	PterodactylWebhook *PterodactylWebhook

	// Caches for read-heavy endpoints.
	statsCache   *cache.Cache[string, any]
	accountCache *cache.Cache[int64, *database.Account]
	sessionCache *cache.Cache[int64, *sessions.TokenPair]
}

// NewServer constructs the API server with caches for read-heavy endpoints.
func NewServer(db *database.DB, cfg *config.Config, skinHandler *SkinHandler) *Server {
	s := &Server{
		db:          db,
		sessions:    sessions.NewService(db, cfg.AccessTokenTTL, cfg.RefreshTokenTTL),
		devices:     devices.NewService(db),
		cfg:         cfg,
		SkinHandler: skinHandler,
		limiter: NewRateLimiter(Config{
			RequestsPerMinute: cfg.RateLimitPerMinute,
			TrustProxy:        cfg.TrustProxy,
		}),
		// TTL caches for read-heavy endpoints.
		statsCache:   cache.New[string, any](30 * time.Second),
		accountCache: cache.New[int64, *database.Account](5 * time.Minute),
		sessionCache: cache.New[int64, *sessions.TokenPair](2 * time.Minute),
	}

	// The async skin worker persists validated uploads and invalidates the
	// per-account skin list cache so a poll right after completion sees it.
	s.queue = newSkinQueue(func(job SkinJob) error {
		info, err := auth.SniffImage(job.Data)
		if err != nil {
			return err
		}
		if !auth.ValidSkinDimensions(info.Width, info.Height) {
			return errors.New("invalid skin dimensions")
		}
		if _, err := s.db.UpsertSkin(job.AccountID, job.Variant, info.ContentType, job.Data); err != nil {
			return err
		}
		s.statsCache.Delete("skins:list:" + strconv.FormatInt(job.AccountID, 10))
		return nil
	})

	return s
}

// Close shuts down background workers and connections held by the server so the
// process can exit cleanly and tests do not leak goroutines or DB connections.
// It is safe to call Close more than once.
func (s *Server) Close() {
	if s.queue != nil {
		s.queue.Close()
	}
	if s.limiter != nil {
		s.limiter.Stop()
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

// Handler returns the root HTTP handler with all routes registered.
//
// Route groups:
//   - Public: /api/v1/auth/*, /api/v1/updates/manifest, /api/v1/proxy/auth,
//     /api/v1/webhooks/pterodactyl, /healthz, /readyz
//   - Protected (authenticated users): /api/v1/account/*, /api/v1/skins/*,
//     /api/v1/skin-files/*, /api/v1/tasks/*
//   - Admin (X-Admin-Token): /api/admin/*, /api/server/*
//
// Rate limiting is applied once per route group. Handlers must not wrap
// themselves in the limiter again or every request would consume two tokens.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// ---- PUBLIC routes (no bearer token required) ----
	public := http.NewServeMux()
	public.HandleFunc("POST /api/v1/auth/register", s.handleRegister)
	public.HandleFunc("POST /api/v1/auth/login", s.handleLogin)
	public.HandleFunc("POST /api/v1/auth/refresh", s.handleRefresh)
	public.HandleFunc("POST /api/v1/auth/logout", s.handleLogout)
	public.HandleFunc("POST /api/v1/auth/logout-all", s.handleLogoutAll)
	public.HandleFunc("POST /api/v1/auth/forgot", s.handleForgotPassword)
	public.HandleFunc("POST /api/v1/auth/reset", s.handleResetPassword)
	public.HandleFunc("GET /api/v1/updates/manifest", s.handleUpdatesManifest)
	public.HandleFunc("GET /readyz", s.handleReadyz)
	public.HandleFunc("POST /api/v1/proxy/auth", s.handleProxyAuth)
	if s.PterodactylWebhook != nil {
		// Authenticated by the HMAC signature over the raw body, not a bearer
		// token, so it belongs to the public group.
		public.HandleFunc("POST /api/v1/webhooks/pterodactyl", s.PterodactylWebhook.Handle())
	}

	// ---- PROTECTED routes (require a valid bearer token) ----
	protected := http.NewServeMux()
	protected.HandleFunc("GET /api/v1/account/me", s.handleAccountMe)
	protected.HandleFunc("POST /api/v1/account/password", s.handleChangePassword)
	protected.HandleFunc("GET /api/v1/account/sessions", s.handleListSessions)
	protected.HandleFunc("POST /api/v1/account/sessions/revoke", s.handleRevokeSession)
	protected.HandleFunc("POST /api/v1/account/sessions/logout-others", s.handleLogoutOtherDevices)

	// Account-scoped, named skins (variant is a user-chosen name such as
	// "default"). These are backed by the skins table.
	protected.HandleFunc("PUT /api/v1/skins/{variant}", s.handleUploadSkin)
	protected.HandleFunc("GET /api/v1/skins", s.handleListSkins)
	protected.HandleFunc("GET /api/v1/skins/{account}/{variant}", s.handleGetSkin)
	protected.HandleFunc("DELETE /api/v1/skins/{variant}", s.handleDeleteSkin)
	protected.HandleFunc("POST /api/v1/skins/upload/{variant}", s.handleAsyncSkinUpload)

	// Processing tasks for async uploads.
	protected.HandleFunc("GET /api/v1/tasks/{id}", s.handleTaskStatus)

	// ID-addressed processed skin files (upload -> immutable image + avatar).
	// They live under a separate prefix because "{id}" would collide with the
	// account-scoped "{variant}" pattern above.
	if s.SkinHandler != nil {
		protected.HandleFunc("POST /api/v1/skin-files", s.SkinHandler.handleUpload)
		protected.HandleFunc("GET /api/v1/skin-files/{id}", s.SkinHandler.handleGetSkin)
		protected.HandleFunc("GET /api/v1/skin-files/{id}/avatar", s.SkinHandler.handleGetAvatar)
		protected.HandleFunc("DELETE /api/v1/skin-files/{id}", s.SkinHandler.handleDeleteSkin)
		protected.HandleFunc("POST /api/v1/skin-files/{id}/active", s.SkinHandler.handleSetActiveSkin)
	}

	// ---- ADMIN routes (require X-Admin-Token) ----
	admin := http.NewServeMux()
	admin.HandleFunc("POST /api/admin/hwid/ban", s.handleBanHWID)
	admin.HandleFunc("POST /api/admin/hwid/unban", s.handleUnbanHWID)
	admin.HandleFunc("GET /api/admin/hwid/bans", s.handleListHWIDBans)
	admin.HandleFunc("GET /api/admin/users", s.handleAdminListUsers)
	admin.HandleFunc("DELETE /api/admin/users/{id}", s.handleAdminDeleteUser)
	admin.HandleFunc("POST /api/admin/users/{id}/minecraft-ban", s.handleAdminMinecraftBan)
	admin.HandleFunc("GET /api/admin/users/{id}/roles", s.handleAdminUserRoles)
	admin.HandleFunc("POST /api/admin/users/{id}/roles", s.handleAdminGrantRole)
	admin.HandleFunc("DELETE /api/admin/users/{id}/roles/{role}", s.handleAdminRevokeRole)
	admin.HandleFunc("POST /api/admin/users/{id}/permissions", s.handleAdminGrantPermission)
	admin.HandleFunc("GET /api/admin/invites", s.handleAdminListInvites)
	admin.HandleFunc("POST /api/admin/invites", s.handleAdminAddInvite)
	admin.HandleFunc("DELETE /api/admin/invites/{id}", s.handleAdminRemoveInvite)
	admin.HandleFunc("GET /api/admin/roles", s.handleAdminListRoles)
	admin.HandleFunc("GET /api/admin/permissions", s.handleAdminListPermissions)
	admin.HandleFunc("POST /api/server/start", s.handleStartServer)

	// Middleware-wrapped route groups. Admin routes only require the
	// X-Admin-Token header; protected routes only require a bearer token.
	publicGroup := s.limitMiddleware(public)
	protectedGroup := s.limitMiddleware(s.authenticateMiddleware(protected))
	adminGroup := s.limitMiddleware(s.adminAuthMiddleware(admin))

	wrap := func(next http.Handler) http.Handler {
		return s.withSecurityHeaders(s.withRecovery(withRequestID(s.withRequestLog(next))))
	}

	mux.Handle("/api/v1/auth/", wrap(publicGroup))
	mux.Handle("/api/v1/updates/", wrap(publicGroup))
	mux.Handle("/api/v1/proxy/", wrap(publicGroup))
	mux.Handle("/api/v1/webhooks/", wrap(publicGroup))
	mux.Handle("/api/v1/account/", wrap(protectedGroup))
	// "/api/v1/skins" is an exact pattern; "/api/v1/skins/" is the prefix that
	// also covers the sub-paths. Both are needed: Go's mux does not redirect a
	// prefix pattern to a missing exact match and vice versa.
	mux.Handle("/api/v1/skins", wrap(protectedGroup))
	mux.Handle("/api/v1/skins/", wrap(protectedGroup))
	mux.Handle("/api/v1/skin-files", wrap(protectedGroup))
	mux.Handle("/api/v1/skin-files/", wrap(protectedGroup))
	mux.Handle("/api/v1/tasks/", wrap(protectedGroup))
	mux.Handle("/api/admin/", wrap(adminGroup))
	// The server-control endpoint lives under /api/server/, not /api/admin/,
	// and must be mounted on the admin group or it is unreachable.
	mux.Handle("/api/server/", wrap(adminGroup))

	mux.Handle("/healthz", wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})))
	// /readyz touches the database, so it stays rate limited but needs no auth.
	mux.Handle("/readyz", wrap(s.limitMiddleware(http.HandlerFunc(s.handleReadyz))))

	// Frontend web handler as fallback.
	mux.Handle("/", web.Handler())

	return wrap(mux)
}

// limitMiddleware returns an http.Handler that applies rate limiting to the
// given handler.
func (s *Server) limitMiddleware(next http.Handler) http.Handler {
	return s.limiter.HandlerMiddleware(next)
}

// adminAuthMiddleware requires a valid X-Admin-Token header. Requests without a
// valid admin token receive 401 Unauthorized (or 404 if no admin token is
// configured).
func (s *Server) adminAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.adminAuthorized(w, r) {
			return
		}
		next.ServeHTTP(w, r)
	})
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

// CachedAccount retrieves an account from cache or DB, caching the result.
func (s *Server) CachedAccount(id int64) (*database.Account, error) {
	if acc, ok := s.accountCache.Get(id); ok {
		return acc, nil
	}
	acc, err := s.db.AccountByID(id)
	if err != nil {
		return nil, err
	}
	s.accountCache.Set(id, acc)
	return acc, nil
}

// CachedStats retrieves statistics from cache or computes them fresh.
func (s *Server) CachedStats(key string, compute func() (any, error)) (any, error) {
	if val, ok := s.statsCache.Get(key); ok {
		return val, nil
	}
	val, err := compute()
	if err != nil {
		return nil, err
	}
	s.statsCache.Set(key, val)
	return val, nil
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
