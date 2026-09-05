# Auth backend (Go + SQLite)

Project-owned backend authentication service. Implements the common account
system described in `docs/AUTH.md`: registration, login, refresh, logout and
logout-all, with server-side session expiry and revocation, and device
identity mapping.

This is a scaffold/first implementation. Microsoft auth, the Velocity plugin,
LimboAuth and the UI are intentionally out of scope.

## Layout

- `main.go` — entry point (HTTP server bootstrap).
- `config/` — environment-driven configuration; no hardcoded secrets.
- `database/` — SQLite connection, idempotent schema, persisted models
  (accounts, devices, sessions, access tokens).
- `auth/` — bcrypt password hashing and opaque token generation/hashing.
- `sessions/` — session lifecycle (create, refresh, validate, revoke).
- `devices/` — device identity registration.
- `api/` — HTTP handlers, routing and per-IP rate limiting.

## Security model

- Passwords are hashed with bcrypt; never stored or logged in plaintext.
- Tokens are random opaque values; only SHA-256 digests are stored, so a
  database leak does not expose usable tokens.
- Access tokens are short-lived (default 15 minutes). Refresh tokens are
  server-side revocable and capped at a 30-day session lifetime.
- Expiry and revocation are enforced server-side; client-side expiry alone is
  never trusted.
- Registration, login and refresh are rate-limited per client IP.

## Configuration

All configuration is read from environment variables (see `.env.example`).
No secrets are hardcoded.

## Run

```
cd server
go run .            # reads AUTH_* env vars, defaults to :8080 and auth.db
```

## Endpoints

| Method | Path                        | Auth | Description                          |
|--------|-----------------------------|------|--------------------------------------|
| POST   | /api/v1/auth/register       | no   | Create account + session             |
| POST   | /api/v1/auth/login          | no   | Log in, create session               |
| POST   | /api/v1/auth/refresh        | no   | Exchange refresh token for access    |
| POST   | /api/v1/auth/logout         | no   | Revoke the session (refresh token)   |
| POST   | /api/v1/auth/logout-all     | yes  | Revoke all sessions for the account  |
| GET    | /healthz                   | no   | Liveness probe                       |

Register/login/refresh return `access_token`, `refresh_token`, `token_type`
and `expires_in`. Protected endpoints take the access token as a `Bearer`
token in the `Authorization` header.

## Tests

```
cd server
go test ./...
```
