# BACKEND API

Reference for the project-owned Go + SQLite auth backend (`server/`). All
responses are JSON. Unless stated otherwise, request bodies must be JSON objects
capped at 1 MiB.

Base path: `/api/v1`.

## Authentication

Protected endpoints expect a bearer access token:

```
Authorization: Bearer <access_token>
```

Access tokens are short-lived (default 15 minutes). Refresh tokens establish a
session and are capped at 30 days. Tokens are opaque; only their SHA-256
digests are stored server-side.

## Rate limiting

All route groups are rate limited per client IP (token bucket, default 60
requests/minute, configurable with `AUTH_RATE_LIMIT_PER_MINUTE`). The limiter
keys on the TCP peer address unless `AUTH_TRUST_PROXY` is enabled; only then
are `X-Forwarded-For` / `X-Real-IP` honoured. Exceeding the limit returns
`429 Too Many Requests` with `Retry-After`.

## Auth endpoints

### POST /api/v1/auth/register

Creates an account and a session.

Request:
```json
{ "username": "player1", "password": "supersecret", "device_id": "dev-1", "device_name": "Desktop" }
```
- `username`: 3-32 chars, `[A-Za-z0-9_-]`.
- `password`: minimum 8 chars.

Responses:
- `201 Created` — token response (below).
- `400 Bad Request` — invalid username/password.
- `409 Conflict` — username already taken.

### POST /api/v1/auth/login

Request:
```json
{ "username": "player1", "password": "supersecret", "device_id": "dev-1", "device_name": "Desktop" }
```

Responses:
- `200 OK` — token response.
- `401 Unauthorized` — invalid credentials (unknown user and bad password are
  indistinguishable).
- `403 Forbidden` — account or hardware is banned.

### POST /api/v1/auth/refresh

Request:
```json
{ "refresh_token": "<token>" }
```
Responses:
- `200 OK` — token response with a fresh access token.
- `401 Unauthorized` — invalid or expired refresh token.

### POST /api/v1/auth/logout

Revokes the session identified by the refresh token. Idempotent.

Request:
```json
{ "refresh_token": "<token>" }
```
Response: `200 OK` `{ "status": "ok" }`.

### POST /api/v1/auth/logout-all

Requires a bearer access token. Revokes every session for the account.

Response: `200 OK` `{ "status": "ok" }`.

## Account endpoints

### GET /api/v1/account/me

Requires bearer auth. Returns the current account profile.

Response `200 OK`:
```json
{ "id": 1, "username": "player1", "created_at": "2026-01-01T00:00:00Z" }
```

### POST /api/v1/account/password

Requires bearer auth. Changes the password and revokes all sessions.

Request:
```json
{ "current_password": "supersecret", "new_password": "newsecret1" }
```
Responses:
- `200 OK` `{ "status": "password changed" }`
- `400 Bad Request` — new password too short.
- `401 Unauthorized` — current password incorrect.

### GET /api/v1/account/sessions

Requires bearer auth. Lists the account's sessions, newest first.

Response `200 OK`:
```json
{ "sessions": [
  { "id": 2, "device_id": 0, "revoked": false,
    "created_at": "...", "last_used_at": "...", "refresh_expires_at": "..." }
] }
```

### POST /api/v1/account/sessions/revoke

Requires bearer auth. Revokes one of the caller's sessions.

Request:
```json
{ "session_id": 2 }
```
Responses:
- `200 OK` `{ "status": "revoked" }`
- `403 Forbidden` — the session belongs to another account.
- `404 Not Found` — unknown session id.

## Token response

```json
{
  "access_token": "<opaque>",
  "refresh_token": "<opaque>",
  "token_type": "Bearer",
  "expires_in": 900
}
```

## Admin endpoints

Admin endpoints are disabled (return `404`) unless `AUTH_ADMIN_TOKEN` is
configured. Requests must present the token in the `X-Admin-Token` header,
compared in constant time.

### POST /api/admin/hwid/ban

Request: `{ "device_id": "<hwid>" }`. Response: `200 OK`.

### POST /api/admin/hwid/unban

Request: `{ "device_id": "<hwid>" }`. Response: `200 OK`.

### GET /api/admin/hwid/bans

Response `200 OK`: `{ "bans": ["<hwid>", ...] }`.

### POST /api/server/start

Administrative server-start hook. Disabled unless `AUTH_ADMIN_TOKEN` is set.
Runs the configured Pterodactyl control script with a 30s timeout and never
returns command output to the client.

## Skin endpoints

All skin endpoints require a bearer access token.

Account-scoped named skins (variant is a user-chosen name, e.g. `default`):

| Method | Path | Purpose |
|--------|------|---------|
| PUT | `/api/v1/skins/{variant}` | Upload/replace a named skin (PNG or JPEG, 64x64 or 64x32, max 1 MiB, max 2 uploads / 3h). |
| GET | `/api/v1/skins` | List the authenticated account's skins. |
| GET | `/api/v1/skins/{account}/{variant}` | Fetch another account's named skin. |
| DELETE | `/api/v1/skins/{variant}` | Delete own named skin. |
| POST | `/api/v1/skins/upload/{variant}` | Async upload: `202 Accepted` + `task_id`. |

Processed skin files (immutable PNG plus derived 64x64 head avatar). Ids are
`<accountID>-<uuid>`; delete and activate are owner-only and answer 404 for
other accounts:

| Method | Path | Purpose |
|--------|------|---------|
| POST | `/api/v1/skin-files` | Multipart upload; field name `skin`. Returns `skin_id` and `avatar_url`. |
| GET | `/api/v1/skin-files/{id}` | Raw processed PNG. |
| GET | `/api/v1/skin-files/{id}/avatar` | 64x64 avatar PNG (face + hat overlay composited). |
| DELETE | `/api/v1/skin-files/{id}` | Delete the file and avatar (owner only). |
| POST | `/api/v1/skin-files/{id}/active` | Mark a recent skin as active (owner only, rate limited). |

Async task status: `GET /api/v1/tasks/{id}` (only the account that created the
task can read it).

## Operational endpoints

### GET /healthz

Liveness. Always `200 OK` `{ "status": "ok" }` while the process is running.

### GET /readyz

Readiness. Pings the database. `200 OK` `{ "status": "ready" }` when reachable,
`503 Service Unavailable` otherwise.

## Error format

```json
{ "error": "human-readable message" }
```

## Security headers

Every response includes `X-Content-Type-Options: nosniff`,
`X-Frame-Options: DENY`, `Cache-Control: no-store` and
`Referrer-Policy: no-referrer`.
