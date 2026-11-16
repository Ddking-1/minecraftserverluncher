# Security audit (2026-11-16)

Scope: full backend (server/), launcher (launcher/), web static assets, CI workflows.

## Fixed in this batch

- **Log/header injection via X-Request-Id**: client-provided request id is now
  sanitized (control chars stripped, capped at 64 bytes) in `withRequestID`.
- **Registration/login spam**: honeypot field (`website`) rejects automated
  submissions; per-username sliding-window rate limit (20/min) added on top
  of the existing per-IP limiter.
- **Account takeover via unapproved accounts**: account approval workflow
  (`approved` column + admin endpoint + login gate).
- **Skin abuse**: skin-file approval workflow; non-owners can only read
  approved skins.
- **Update manifest integrity**: HMAC-SHA256 over version|url|sha256|mandatory,
  verified by the launcher before trusting the manifest (constant-time compare).

## Findings and recommendations (not yet fixed)

1. **Single shared admin token.** All `/api/admin/*` routes share one
   `X-Admin-Token`. The RBAC schema + per-user roles/permissions exist and
   are manageable, but enforcement on admin endpoints still uses the shared
   token. Recommendation: migrate high-sensitivity admin operations to bearer
   tokens checked with `RequirePermission` (the middleware already exists).
2. **No CAPTCHA / proof-of-work.** The honeypot is a weak anti-bot. For
   production, add an external CAPTCHA or a lightweight proof-of-work
   challenge on register/login.
3. **Multipart memory pressure.** `ParseMultipartForm(32 << 20)` allows up
   to 32 MiB of multipart overhead before `ProcessSkin` rejects >100 KiB.
   Recommendation: lower the parse limit to ~2 MiB.
4. **Password policy is length-only** (`validPassword`: >= 8 chars).
   Recommendation: add a minimum complexity check (mixed case / digit) and
   a breached-password check (k-anonymity) for production.
5. **TrustProxy defaults to false** (good). Document that it must only be
   enabled behind a trusted reverse proxy that overwrites
   X-Forwarded-For / X-Real-IP; otherwise rate limiting is bypassable.
6. **HTTPS not enforced in code.** TLS termination is expected at the proxy;
   document and verify HSTS at the edge.
7. **Skin read endpoints are authenticated but not ownership-scoped.** Any
   authenticated user can fetch any skin by id. Acceptable for a shared skin
   gallery, but revisit if skins become private.
8. **No per-user rate limit on skin uploads beyond the existing limiter.**
   The `skin_upload_limits` table exists; verify it is enforced in the
   upload path.

## Confirmed safe

- Passwords hashed with bcrypt (cost 12); tokens stored as SHA-256 digests.
- SQL queries use bound parameters; no `fmt.Sprintf` SQL construction.
- Skin storage validates id pattern (`storageIDPattern`) and caps file reads.
- Admin token compared with `crypto/subtle.ConstantTimeCompare`.
- Rate limiter uses the right-most XFF entry only when TrustProxy is on.
- Security headers (nosniff, frame deny, no-store, referrer no-referrer) set
  globally.

## Update (2026-11-16, second pass)

- Tripwire anomaly tracking added (DB + admin endpoint + recording).
- Admin authorization now accepts per-user bearer tokens with the "admin"
  permission/role, in addition to the shared X-Admin-Token (backward compat).
- Cloudflare Turnstile CAPTCHA integration (config-gated, dev mode when unset).
- Admin dashboard and skin-manager rewired to real endpoints.
- Remaining: migrate individual admin handlers to RequirePermission for
  finer-grained control; add CAPTCHA to frontend with a real site key.
