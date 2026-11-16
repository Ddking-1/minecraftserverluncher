# Security

This document summarizes the hardening applied to the Minecraft Launcher ecosystem
and how to report a vulnerability.

## Principles

- The client is never the final security authority. Server-side authorization is
  authoritative.
- Never trust client-supplied identity, metadata, or headers (X-Forwarded-For is
  ignored unless `AUTH_TRUST_PROXY` is set).
- All admin operations require either the shared `X-Admin-Token` or a bearer
  token for an account holding the `admin` permission/role.

## Hardening in place

- Passwords: bcrypt (cost 12).
- Sessions: short-lived access tokens + refresh tokens, device binding.
- HWID/IP bans enforced server-side.
- Skin uploads: size cap, dimension allow-list, PNG re-encode (strips metadata),
  decompression-bomb guard (header decoded before full decode).
- Rate limiting: per-IP token bucket, per-username login limiter, per-account
  skin upload limit.
- Anti-bot: honeypot field `website`, per-username sliding window, sanitized
  `X-Request-Id`.
- Account & skin approval workflows (admin-gated).
- CAPTCHA: Cloudflare Turnstile (config-gated; dev mode when secret unset) OR
  ALTCHA (self-hosted alternative).
- Tripwires: persisted anomaly events + admin review endpoint.
- Security headers: X-Content-Type-Options, X-Frame-Options, Referrer-Policy,
  Cache-Control: no-store, Content-Security-Policy.
- Update manifest: HMAC-SHA256 signature verification in launcher + backend.
- SQL: parameterized queries throughout; no string concatenation in queries.
- Path traversal: skin ids validated against `^[A-Za-z0-9_-]{1,128}$`.

## Reporting

Open an issue or contact the maintainer. Do not post exploit details publicly.
Secrets must never be committed; use environment variables / Sites Secrets.
EOF
echo created