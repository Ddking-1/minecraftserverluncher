# api

HTTP layer of the Go backend: routing, handlers, middleware and request/response
types. See the parent `server/README.md` for the full endpoint reference.

## Contents

- `handlers.go` — server wiring, route registration, auth handlers and shared
  helpers (`authenticate`, `bearerToken`, `writeJSON`/`writeError`, `decodeBody`).
- `account.go` — account profile, password change, session list/revoke.
- `admin.go` — admin token gate and HWID ban endpoints.
- `health.go` — `/healthz` and `/readyz`.
- `middleware.go` — request-ID, request log, panic recovery, security headers.
- `ratelimit.go` — per-IP token-bucket rate limiter with expiry pruning.
- `*_test.go` — unit tests for handlers, middleware, rate limiter, account,
  sessions and admin.

## Security properties

- Login runs a dummy bcrypt compare for unknown accounts (timing defense).
- Admin token gate uses `subtle.ConstantTimeCompare`.
- Request bodies capped via `MaxBytesReader`.
- `decodeBody` is strict; all queries use parameterized SQL.
