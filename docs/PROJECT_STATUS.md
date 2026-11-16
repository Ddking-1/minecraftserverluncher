# PROJECT STATUS — Gap Analysis and Minimum Publishable Baseline

> ⚠️ **Disclaimer:** The code in this repository is heavily generated and maintained by AI assistants. As such, there is no guarantee of continuous updates, bug fixes, or long-term maintenance. Use at your own risk.

- **Date:** 2026-11-16
- **Branch:** `master` (`a9a041a`, matches `origin/master`)
- **Status:** DRAFT — subject to owner scope decisions (see "Open questions" at the bottom)
- **Purpose:** Give a single, readable picture of (a) what already works and is verified, (b) what is still missing, and (c) what the smallest publishable baseline ("v1") needs. Written for an owner who is not a programmer.

---

## 1. What already works (verified, green)

### Backend (Go + SQLite, `server/`) — VERIFIED
- Account system: register, login, refresh, logout, logout-all, profile, password change, session list/revoke.
- Security: bcrypt passwords, opaque hashed tokens, expiry + revocation, device identity, HWID bans, per-IP and per-user rate limiting, request logging, panic recovery, security headers, `/healthz` + `/readyz`, graceful shutdown.
- Skin files: upload (`multipart` field `skin`), get, avatar (64x64 head), owner-only delete/activate, disk persistence, concurrency-safe storage. Verified end-to-end over HTTP and HTTPS.
- Pterodactyl webhook: HMAC-verified, fails closed (401) without a secret.
- Tests: `go build ./...`, `go vet ./...`, `go test ./...` pass (200+ tests); fuzz targets for skin processor and rate limiter.
- Config: env-driven, no hardcoded secrets; `server/.env.example` documents variables.

### Launcher (C++ / Qt6, `launcher/`) — VERIFIED on Windows 11
- Build: Qt 6.6.3 + MinGW GCC 13.1 and MSVC 2022 (`/W4 /WX`), 23/23 CTest passing on both.
- Core: instances, settings store, mod scanner, skin cache, rotating logger, file integrity, version, Java detector, game paths, game launcher.
- Auth: backend client (15 s timeout, status-code error mapping), DPAPI token storage, session manager (login/restore/logout), HTTPS-only remote policy (HTTP allowed for localhost).
- GUI: MainWindow with backend URL edit/save, login/logout, status line, session restore; SkinPreviewWidget upload + avatar; LoginDialog; SplashScreen; EditorSettingsWidget.
- Skin upload works end-to-end against a real backend over HTTPS.
- Update system: manifest parser, checker, UpdateManager (signed manifest) — implemented but **not wired into startup**.
- CLI: filter/version options.

### Velocity plugin (`velocity/`) — BUILDS
- Builds against `velocity-api:4.2.0` with JDK 25 (`./gradlew build` → `minecraftserverlauncher-velocity-0.1.0.jar`).
- Contains: plugin main, config, BackendAuthClient, SkinInjection.
- Join-time auth reads a `minecraftserverlauncher-token` GameProfile property that the **launcher does not set yet** → end-to-end join authorization cannot succeed. This is a documented gap, not a security relaxation.

### Website / web (`server/web/`, `web/`) — PARTIAL
- `server/web/web.go`: embedded static file server, routes `/` → `skin-manager.html`, `/dashboard`, `/admin`.
- `web/static/skin-manager.html` exists.
- No dashboard/admin HTML templates are present in the repo (only `skin-manager.html`), so `/dashboard` and `/admin` currently serve nothing useful.
- No forgot-password, no admin user management, no registration page beyond skin-manager.

### CI / infrastructure — WORKING (Linux)
- Self-hosted runner `debian-ai-runner` on `192.168.1.23` (Debian 13), systemd, labels `self-hosted,linux,x64,debian-13`.
- `linux-ci.yml` runs Go suite + Velocity build on every push to `master` — green.
- Windows workflows (`launcher.yml`, `backend.yml`) still target a Windows runner that is **not registered** → they queue (with path filters) and never run. This is the main remaining CI gap.

### Documentation — EXTENSIVE
- `docs/` has ~30 files: architecture, decisions, API, security review, threat model, code wiki, requirements, update spec, etc.
- `docs/DECISIONS.md` records DECISION-001…006.
- `docs/TASKS_40.md` is the real roadmap (~65 items across 7 phases + addendum).

---

## 2. What is still missing (open, grouped)

### A. Blocking decisions (owner action, no code yet)
- **DECISION-004 — GPLv3 distribution / public launcher source repository.** The launcher is derived from Prism (GPLv3). Publishing it requires a public source repository and license compliance. Not resolved.
- **DECISION-005 — metadata source.** Where Minecraft/Fabric version metadata comes from. Blocks `ResolveLibraries` / `ResolveAssets` launch steps and library/asset downloads.
- Cross-account read policy for `GET /api/v1/skin-files/{id}` (currently any authenticated account can read; only delete/activate are owner-scoped). Needs a DECISION.

### B. Backend features not implemented (from `TASKS_40.md`)
- Forgot-password flow (`POST /api/v1/auth/forgot`, `POST /api/v1/auth/reset`, `password_resets` table, token hash + expiry, revoke all sessions) — **N2–N4**.
- Admin panel backend: list users, ban/unban, delete (cascade) — **N5–N7**.
- Dashboard HTML routes + forms wired to real endpoints — **N1, N8, N9**.
- Per-user rate limiting for auth endpoints (currently per-IP) — **N21, N23**.
- Background pruning of expired rate-limit buckets and password resets — **N22**.
- Body-size middleware test + content-type sniffing on all uploads — **N24**.
- Skin upload per-user rate limit (2 per 3 h) — **T-A**.
- Strict skin format validation (dimensions 64x64 or 64x32) — **T-B** (partially done server-side).
- Magic-byte MIME sniffing — **T-C**.
- Refresh token rotation — **T26**.
- Login: reject before bcrypt when HWID banned; avoid enumeration — **T27**.
- Password strength checks — **T28**.
- Config validation (TTLs, ports, bcrypt cost) — **T29**.
- Limit refresh/logout by user, not just IP — **T30**.
- Device-ownership enforcement on session create/revoke — **T25**.
- Uniqueness index tests — **T31**.

### C. Launcher features not implemented
- `ResolveLibraries` / `ResolveAssets` launch steps are stubs (blocked on DECISION-005) — **T33–T38 area**.
- Auto-update implemented but not wired into startup — **TASKS BATCH 11/12**.
- Skin client UI (upload + display) in the launcher — **T-D** (partially: SkinPreviewWidget exists).
- SplashScreen wired into main flow — **N14**.
- Editor window + field widgets bound to EditorConfig — **N10–N12**.
- Skin preview widget in launcher loading by username — **N13**.
- InstancePersistence round-trip tests, RotatingFileLogger retention edge, UpdateChecker incompatible-decision test, SettingsStore JVM-arg injection rejection, JavaDetector plausible-dirs test, BackendAuthClient non-regression — **T33–T38**.

### D. Velocity plugin
- Backend `POST /api/v1/proxy/auth` endpoint (uuid, access_token hash, server name → ok/deny), rate-limited, admin-token gated — **N16**.
- `BackendAuthClient.java` HTTP client — **N15**.
- `LoginEvent` listener + `ServerPreConnectEvent` guard — **N18–N19**.
- proxy_sessions helper or stateless HMAC — **N17**.
- Gradle build green on master — **N20** (builds, but end-to-end join auth unverified).

### E. Web / website
- Only `skin-manager.html` exists; `/dashboard` and `/admin` have no templates — **N8**.
- No registration page, no forgot-password page, no admin table UI.
- CSP headers, static asset serving — **T17**.

### F. CI / infrastructure
- No Windows self-hosted runner registered → `launcher.yml` and `backend.yml` queue forever on `launcher/**` or `server/**` changes.
- Go `-race` + `-cover` not in backend workflow — **N25 / T39**.
- Launcher stays on `windows-2022` (GitHub-hosted) per T39, but currently uses self-hosted Windows.

### G. Packaging / distribution (not started)
- No installer (NSIS exists under `packaging/nsis/` but only `.nsi`? — needs verification).
- No code signing, no public release artifacts, no update channel/hosting.
- OpenSSL 3 DLLs + MinGW runtime staging is implemented in CMake (good), but the release process is undocumented.

### H. Legal / licensing
- DECISION-004 unresolved → blocks public distribution of the launcher.
- `docs/LICENSE_AUDIT.md` open items.

---

## 3. What is NOT needed for a minimum publishable v1

To keep v1 small, these can be deferred (they exist in TASKS_40 but are not blocking a first release):
- Admin panel (N5–N7, N8 admin table) — can be a CLI/DB operation for v1.
- Forgot-password (N2–N4) — can be manual reset for v1, but **strongly recommended** before public users.
- Editor window + field widgets (N10–N12) — the editor settings widget already exists; a full editor window is polish.
- Per-user rate limits beyond IP (N21–N24) — IP limits exist; per-user is hardening.
- Skin per-user rate limit (T-A) and magic-byte sniffing (T-C) — recommended before public, but uploads already validate dimensions.
- CI `-race`/`-cover` (N25/T39) — quality, not blocking.
- Full web dashboard — `skin-manager.html` + backend API is enough for a small test team.

---

## 4. Proposed minimum publishable baseline (v1)

This is a proposal for discussion, not a decision:

**v1 = "one Minecraft server, small test team, beta-quality"**
1. Launcher (Windows) that logs in, downloads/updates the modpack, launches Minecraft 1.21.11 + Fabric 0.19.2.
2. Backend (Go + SQLite) with auth, sessions, skin upload/avatar, HWID bans, readyz.
3. Velocity plugin that validates joins via the backend (needs N15–N19).
4. Auto-update wired into launcher startup (needs wiring).
5. Library/asset download steps implemented (needs DECISION-005).
6. A simple skin-manager page (already exists).
7. Documentation: how to install, run the backend, configure env vars, and the "beta, not always available" notice.

**Explicitly out of v1:** admin web UI, forgot-password (manual reset), editor window, per-user rate limits, full web dashboard, code signing, multi-server UI.

---

## 5. Owner decisions (2026-11-16 discussion) — ANSWERED

1. **Scope of v1 — Velocity included.** v1 needs the Velocity plugin WITH join authorization/permission management. (DECISION-level: proxy auth is in scope.)
2. **Auto-update hosting — GitHub Releases + backend redirect.** Update files live on GitHub Releases of the public launcher repo (free, versioned, CDN). The client checks BOTH the backend and GitHub; the backend can tell the client the update host moved (DNS/URL change), and the client refreshes its stored location. This gives a primary + backup path. Recorded as **DECISION-007**.
3. **Metadata source (DECISION-005) — RESOLVED.** Online fetch from Mojang/Fabric meta with a local cache and a bundled fallback list (cache-first, network refresh, offline fallback). Recorded as **DECISION-009**.
4. **Distribution/legal — publish only what is legally required.** Keep private by default; publish publicly only what a license (Prism GPLv3) requires, and prefer publishing slightly more than the bare minimum when that reduces legal risk. Recorded as **DECISION-008**. (DECISION-004 still needs counsel confirmation of exact GPLv3 obligations.)
5. **Windows CI — owner asked "why change it?"** See §9. Recommendation stands: move to GitHub-hosted `windows-latest` (no maintenance) OR register a Windows self-hosted runner. Pending owner choice.
6. **Beta & access control.** Beta/alpha labels everywhere. Registration is **invite-only by email allowlist** (admin pre-approves tester emails; registration with an unlisted email is rejected with "not a beta tester"). Target capacity: ~200 users at beta launch. Full admin permission system with granular, per-user permissions (e.g. Minecraft ban but not system ban; approve username-change requests; approve skin requests). User data isolation: the server must only return data belonging to the requesting account; admin panel must not even be referenced in HTML for non-admins (even in dev). Skin uploads must be format-validated (magic bytes + dimensions + size) to prevent malware. Anti-bot + ban handling. Modular, independently testable builds with emulated/mock backends.
7. **Modpack delivery — mrpack-like.** An instruction file lists mods to download (from Modrinth/Forge/etc.) plus the project's own configs. Must be tamper-protected (signed manifest + hash verification, already in UpdateManager) and any tampering attempt logged + silently reported (what was tried, by whom).

---

## 6. New requirements captured from the discussion (not in TASKS_40 yet)

These must be added to the roadmap (TASKS_40 or a new phase) before implementation:

- **Invite-only registration (allowlist by email).** New table `allowed_emails` (or `invites`); registration checks the allowlist; unlisted emails rejected with a clear "not a beta tester" message; admin can add/remove invites. (Backend + web + admin.)
- **Granular admin permission system.** Per-user permission flags (e.g. `ban.minecraft`, `ban.system`, `approve.username_change`, `approve.skin`, `manage.invites`, `manage.updates`), enforced server-side on every admin endpoint; admin UI only rendered for accounts holding the relevant permission (hidden even in dev builds).
- **User data isolation audit.** Server-side review that every endpoint only returns data owned by the requesting account (or covered by a permission). Admin endpoints gated by permission, not by "is admin".
- **Skin upload hardening.** Magic-byte MIME sniffing, dimension check (64x64 / 64x32), size cap, image-decode validation (no malformed/crafted images), rate limit per user. (T-A, T-B, T-C extended.)
- **Anti-bot.** Rate limiting (exists) + CAPTCHA or proof-of-work on registration/login (new).
- **Beta labeling.** "BETA/ALPHA" banner in launcher, web and admin; version string includes beta tag.
- **Capacity for ~200 concurrent users.** SQLite + Go is fine for 200; verify connection pool, indexes, and rate-limit memory; load test before beta.
- **Modpack delivery (mrpack-like).** Signed instruction manifest listing mods (Modrinth/Forge URLs + hashes) + project configs; client downloads/verifies each; tamper attempts logged + silently reported (what/who). Reuses UpdateManager's signed-manifest machinery.
- **Modular, emulated test builds.** Each subsystem runnable in isolation with mocked backends (e.g. launcher with a fake auth that accepts any registration; server with an in-memory DB) for fast, debuggable testing.
- **Update host design (DECISION-007).** Client checks backend + GitHub; backend can redirect update host.

## 7. Suggested next steps (after your answers)

1. Clarify question 3 (metadata source) — see §8.
2. Decide Windows CI approach — see §9.
3. Resolve DECISION-004 with counsel (legal).
4. Add the new requirements above to `docs/TASKS_40.md` as a new phase.
5. Implement in order: (a) Velocity proxy auth (N15–N20), (b) invite-only registration + admin permissions, (c) modpack/update wiring, (d) skin hardening, (e) beta labeling + capacity test.

## 8. Metadata source — RESOLVED (DECISION-009)

Decision: **online fetch + local cache + bundled fallback list** (cache-first, network refresh, offline fallback). The owner's exact words: "online + helyi cache", i.e. online with a local cache, so it keeps working if Mojang/Fabric is unreachable.

Recorded as DECISION-009 in `docs/DECISIONS.md`.

## 9. Windows CI — RESOLVED: Option A (GitHub-hosted `windows-latest`)

**Decision (owner, 2026-11-16):** both `launcher.yml` and `backend.yml` now use
`runs-on: windows-latest` (GitHub's own Windows machines). Free minutes
(2000/month for private repos), zero maintenance, Qt installable there.

Why not the Debian runner: the launcher is a Windows binary (Qt6 + MSVC/MinGW
Windows toolchain) and Linux cannot build it without a heavy cross-compilation
toolchain. So the Windows build cannot run on Debian. The backend (Go) could,
but keeping both on `windows-latest` is simpler and still within the free tier.

Note: Windows builds are also verified locally by the agent (MSVC + MinGW,
23/23 tests), so even if CI minutes run out the builds are known to work.

**Manual action for the owner:** set a billing/usage alert in GitHub →
Settings → Billing → Actions spending limit, so you are notified before the
2000 free minutes run out. If they do run out, switch to Option B (register
your Windows PC as a self-hosted runner) — see git history / §9 of earlier
versions of this file for the steps.

---

## 7. Notes

- This document is a draft. It will be updated as scope decisions are made.
- `docs/TASKS_40.md` remains the granular roadmap; this file is the high-level view.
- Verified facts come from `docs/AI_CONTEXT.md`, `docs/WORK_STATE.md`, `docs/PROGRESS.md` and the repository tree as of 2026-11-16. Anything not verified is marked as such.
