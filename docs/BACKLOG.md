# Backlog — Minecraft Launcher Ecosystem

Last updated: 2026-11-16. Created for review while the user is away.

The user should delete or annotate lines they do NOT want. Everything not
rejected will be implemented in order, one item per commit, with tests and a
push. Items are grouped by risk; security/quality items are applied first.

## A. Security (applying now — universally desired)

A1. XSS hardening in `web/static/admin.html` — replace `innerHTML` table
    rendering with safe DOM construction / escaping for API-derived fields
    (usernames, skin ids, tripwire details).
A2. Decompression-bomb guard in `server/skins/processor.go` — call
    `image.DecodeConfig` and reject oversized dimensions BEFORE full decode in
    both `ProcessSkin` and `ExtractAvatarHead`.
A3. Ensure admin endpoints enforce `RequirePermission("admin")` in addition to
    `adminAuthorized` (finer-grained, per-user RBAC), while keeping the shared
    token path working.
A4. Add a Content-Security-Policy + X-Frame-Options + X-Content-Type-Options
    header middleware for served web assets.
A5. Sanitize `X-Request-Id` (already done) — verify no other header is
    trusted for identity (X-Forwarded-For gated behind AUTH_TRUST_PROXY).
A6. Skin file route: add explicit traversal test + ensure 404 (not 500) on
    missing id.
A7. Rate-limit verification: confirm login + register + upload + tripwire
    endpoints are all covered; add missing ones.
A8. CSRF: admin web UI uses bearer/token in headers (not cookies) → no CSRF
    needed, but document it. Add SameSite/secure notes for any future cookies.
A9. Secret hygiene: ensure `.env`/`.dev.vars`/secret files are gitignored;
    verify no secrets in repo history (scan).

## B. Stability / correctness (applying now)

B1. Image decode error paths return 4xx not 5xx.
B2. Idempotent shutdown of rate-limiter cleanup goroutine (already done) —
    verify in tests.
B3. Async queue: bounded result retention (already done) — add metric.
B4. Launcher: null-pointer guards in SkinPreviewWidget (already done).
B5. Update manifest HMAC verification end-to-end test.

## C. Feature candidates (PENDING USER APPROVAL — NOT applied yet)

These come from research on what users value in Minecraft launchers
(Prism/MultiMC/GDLauncher/SKLauncher comparisons). Filtered against the
project philosophy: secure, closed, fast, easy for any user.

C1. Auto-detect and assign Java per instance (already partially: JavaDetector).
    Finish: per-instance Java path override + version mismatch warning.
C2. Crash log viewer inside the launcher (color-coded, tail last N lines) —
    easy for users, no security impact. HIGH value.
C3. Instance isolation (already) + "duplicate instance" action.
C4. Modpack import from a TRUSTED/curated source only (not arbitrary
    internet downloads) — to keep "closed". Could be a server-hosted modpack
    manifest endpoint. MEDIUM value, MEDIUM risk.
C5. One-click skin apply from launcher (already wired) — polish UX.
C6. Delta updates for launcher (UpdateManager exists) — wire into startup.
C7. RAM preset profiles (Low/Medium/High) instead of raw MB input.
C8. Offline mode: launch last-played version without network (if cached).
C9. News/changelog panel from backend (low priority, adds network).
C10. Multi-language UI (i18n) — only if requested.

## D. Documentation / ops (applying now)

D1. Update docs/AI_CONTEXT, WORK_STATE, PROGRESS, CHANGELOG per batch.
D2. Add a `docs/SECURITY.md` summarizing hardening + how to report.
D3. Add a simple health/readiness endpoint doc.

## E. Explicitly NOT doing (conflicts with philosophy)

- Cracked / offline-account login (violates "secure, closed").
- Arbitrary mod downloads from untrusted CDNs (violates "closed").
- Telemetry / analytics (violates privacy + "fast").
- Heavy web browser embedded in launcher (violates "fast/lightweight").

## Manual user actions still required

- CAPTCHA keys (Turnstile or ALTCHA swap) — see prior report.
- Admin role grant for at least one account.
- Qt toolchain to build/test the C++ launcher (blocked here).
EOF
echo created