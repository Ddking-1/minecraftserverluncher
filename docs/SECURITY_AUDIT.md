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

## API / DB / Auth Audit (2026-11-17)

Scope: all API endpoints in server/api, all DB queries, auth/session logic.

Findings:
- Test bypasses in production code: NONE. No debug endpoints, no skip-auth flags, no hardcoded credentials in non-test code. Dev-mode CAPTCHA skip is config-gated (TurnstileSecret empty) and documented.
- Leaked API keys / secrets: NONE in tracked files. Secrets are read from env (AUTH_ADMIN_TOKEN, AUTH_TURNSTILE_SECRET, AUTH_UPDATE_SIGNING_KEY, PTERODACTYL_WEBHOOK_SECRET). Web/static and launcher contain no keys. .env family is gitignored.
- IDOR / cross-user data exposure:
  - /api/v1/skins (upload/list/delete) scoped to authenticated accountID.
  - /api/v1/skin-files delete/set-active require ownsSkinFile (ownership parsed from id prefix).
  - /api/v1/skin-files/{id} read enforces skin-approval for non-owners (fail-closed).
  - /api/v1/tasks/{id} checks res.AccountID == accountID.
  - /api/v1/account/sessions/revoke checks srow.AccountID == accountID.
  - /api/v1/account/* scoped to authenticated account.
  - Admin endpoints require adminAuthorized (shared token or bearer with admin permission/role); intended to operate across users.
  - /api/v1/skins/{account}/{variant} GET is intentionally public (any authenticated user may fetch a public skin).
- SQL injection: all queries parameterized; no string concatenation in SQL found.
- Auth: bearer token validation via sessions.ValidateAccess; admin uses constant-time compare for shared token; RBAC permission check for bearer admins.
- Webhook: HMAC signature verified; oversized body rejected with 413 (no silent truncation).

Residual recommendations:
- Admin bearer path ignores DB errors in HasPermissionFast/AccountRoles (fail-closed would be stricter); low risk since errors default to deny.
- Crash report endpoint logs full payload (may include user-provided data); acceptable for debugging but consider redaction.
- No CORS configured (documented); IP-based rate limiting only (documented).
