# BACKEND SECURITY REVIEW

Scope: the project-owned Go + SQLite auth backend under `server/`. This is a
point-in-time review of the implemented surface, not a penetration test.

## Threat model summary

The client is never the final security authority. The backend enforces all
authorization, session lifetime, revocation and ban decisions. Client-supplied
values (username, device id, tokens) are treated as untrusted input.

## Controls in place

| Area | Control |
|------|---------|
| Password storage | bcrypt, cost configurable (4-31), default 12 |
| Tokens | opaque random values; only SHA-256 digests stored |
| Access token lifetime | short (default 15 min), server-enforced expiry |
| Session lifetime | capped at 30 days (`config.maxRefreshTTL`) |
| Revocation | per-session logout and logout-all; checked on every access validation |
| Rate limiting | per-client-IP token bucket on all mutating auth endpoints |
| Login enumeration | identical "invalid credentials" for unknown user and bad password |
| Admin endpoints | disabled (404) unless `AUTH_ADMIN_TOKEN` set; constant-time token compare |
| Server-start endpoint | no shell, fixed configured script path, 30s timeout, output never returned to client |
| Request body size | capped at 1 MiB via `http.MaxBytesReader` |
| SQL | parameterized statements throughout; no string concatenation |
| Secrets | environment-only; `.env` ignored by Git |

## Input validation

- `username`: 3-32 chars, `[A-Za-z0-9_-]`.
- `password`: minimum 8 chars.
- `device_id`: trimmed; admin ban endpoint caps at 256 chars.
- JSON bodies are size-limited and must decode as objects.

## Known limitations / follow-ups

1. Rate limiter state is in-memory and per-process; it resets on restart and is
   not shared across instances. Acceptable for the single-instance target
   (docs/DECISIONS.md resource constraint). A shared store is explicitly out of
   scope.
2. Rate limiter is keyed on `RemoteAddr`; behind a reverse proxy this is the
   proxy address unless the proxy sets a trusted client IP header. Deployments
   behind a proxy must terminate rate limiting at the proxy or add trusted
   header handling (not yet implemented).
3. Token storage on the client is DPAPI-protected on Windows (see
   `launcher/auth/TokenStorage.cpp`) with a documented plaintext fallback.
4. No account lockout after repeated failures beyond the per-IP rate limit.
5. TLS termination is expected to be handled by the deployment (reverse proxy);
   the server does not implement TLS itself.

## Endpoints

See `docs/API.md` for the endpoint reference.

## Change policy

Any change to authentication, authorization, token handling or the admin
surface must update this document and add or extend the corresponding tests in
`server/`.
