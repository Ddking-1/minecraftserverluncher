# Server

This directory contains the Go backend for the private Minecraft network. It
provides HTTP authentication (register/login/refresh/logout), account
management, session and device handling, HWID ban enforcement, an admin
server-start endpoint, and health probes. It uses SQLite for storage.

## Layout

- `api/` — HTTP handlers, routing, middleware (request logging, request ID,
  panic recovery, security headers, rate limiting).
- `auth/` — bcrypt password hashing and opaque token generation/validation.
- `config/` — configuration loading and defaults.
- `database/` — SQLite schema and queries: accounts, sessions, access tokens,
  devices, HWID bans.
- `sessions/` — session and access/refresh token lifecycle service.
- `devices/` — device registration.

## How it works

Registration creates an account (bcrypt cost 12) and a session, returning an
access token + refresh token. Login verifies the password and enforces the
HWID ban list, then issues a fresh token pair. Refresh exchanges a valid
refresh token for a new access token; logout and logout-all revoke sessions.

Tokens are opaque random values; only their SHA-256 digests are stored, so a
database leak does not expose usable tokens. All `/api/v1/*` endpoints are
rate limited per client IP.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/register` | Create an account and start a session |
| POST | `/api/v1/auth/login` | Authenticate and issue tokens |
| POST | `/api/v1/auth/refresh` | Exchange a refresh token for a new access token |
| POST | `/api/v1/auth/logout` | Revoke a single session |
| POST | `/api/v1/auth/logout-all` | Revoke every session of the account |
| GET | `/api/v1/account/me` | Current account profile |
| POST | `/api/v1/account/password` | Change password (revokes sessions) |
| GET | `/api/v1/account/sessions` | List the account's sessions |
| POST | `/api/v1/account/sessions/revoke` | Revoke one session |
| PUT | `/api/v1/skins/{variant}` | Upload a named skin (bearer token) |
| GET | `/api/v1/skins` | List the account's named skins |
| GET | `/api/v1/skins/{account}/{variant}` | Fetch a named skin |
| DELETE | `/api/v1/skins/{variant}` | Delete a named skin |
| POST | `/api/v1/skins/upload/{variant}` | Async named-skin upload (`202` + `task_id`) |
| POST | `/api/v1/skin-files` | Upload a processed skin file (multipart field `skin`) |
| GET | `/api/v1/skin-files/{id}` | Processed PNG |
| GET | `/api/v1/skin-files/{id}/avatar` | 64x64 head avatar |
| DELETE | `/api/v1/skin-files/{id}` | Delete own skin file and avatar |
| POST | `/api/v1/skin-files/{id}/active` | Activate a recent skin (owner only) |
| GET | `/api/v1/tasks/{id}` | Async task status (creator only) |
| POST | `/api/v1/webhooks/pterodactyl` | Pterodactyl webhook (HMAC signature) |
| POST | `/api/server/start` | Start the game server (admin token) |
| POST | `/api/admin/hwid/ban` | Ban a device ID (admin token) |
| POST | `/api/admin/hwid/unban` | Unban a device ID (admin token) |
| GET | `/api/admin/hwid/bans` | List HWID bans (admin token) |
| GET | `/healthz` | Liveness probe |
| GET | `/readyz` | Readiness probe (DB ping) |

## Configuration

- `ListenAddr` — address to listen on (default `:8080`).
- `DBPath` — SQLite database file path.
- `AccessTokenTTL` — access token lifetime (default `15m`).
- `RefreshTokenTTL` — refresh token lifetime (default `720h`, max 30 days).
- `BcryptCost` — bcrypt cost (default `12`).
- `AdminToken` — admin token for privileged endpoints (`""` disables them).
- `PteroScript` — path to the server-start script.
- `RateLimitPerMinute` — per-client request budget (default `60`, env `AUTH_RATE_LIMIT_PER_MINUTE`).
- `TrustProxy` — trust `X-Forwarded-For` / `X-Real-IP` for rate-limit keys (default off, env `AUTH_TRUST_PROXY=1`).
- `SkinStorageDir` — persistent skin/avatar directory (default `./storage/skins`, env `AUTH_SKIN_STORAGE_DIR`).

## Build & test

```bash
cd server
go build ./...
go vet ./...
go test ./... -count=1
```

## Security notes

- No raw tokens stored; only SHA-256 digests. Raw passwords and tokens are
  never logged.
- Login performs a dummy bcrypt comparison when the account does not exist to
  resist timing-based user enumeration.
- Admin token comparison uses `subtle.ConstantTimeCompare`.
- Server start never uses a shell and never echoes command output to clients.
- Request bodies are capped with `MaxBytesReader` (1 MiB).
