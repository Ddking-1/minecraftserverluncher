# AI CONTEXT

> ⚠️ **Disclaimer:** The code in this repository is heavily generated and maintained by AI assistants. As such, there is no guarantee of continuous updates, bug fixes, or long-term maintenance. Use at your own risk.

## Project identity

This is a private Minecraft network launcher/client ecosystem.

First target:
Windows 10/11
Minecraft 1.21.11
Fabric 0.19.2
Java 24

## Web frontend foundation (task branch, 2026-11-18)

An alpha web UI now lives in `server/web/static` and is embedded by
`server/web/web.go`. It is a build-free vanilla ES-module SPA with dedicated
routes `/login`, `/register`, `/dashboard`, `/players` and `/admin`, a dark/light
theme persisted in `localStorage`, a custom interactive canvas skin viewer, a
multi-server context (one enabled server for now, launcher handoff via
`?server=`/`?instance=`) and a registry for plugin statistic modules. The admin
screen has a granular role/permission model and a UI-only simulator; the client
is not an authorization boundary.

This work is on `feat/web-ui-foundation` and is **not merged into `master`** yet.
It uses a local mock data/session provider (`data/api.js`) for the alpha; the
next integration step is replacing it with the authenticated Go API. See
`docs/WEB_FRONTEND.md` for the architecture, test evidence and limitations.
The Go toolchain was available on the development machine for this task, so the
Go web package was compiled and its tests passed (`gofmt`, `go vet`,
`go build ./web/...`, `go test -count=1 ./web/...`), and the embedded SPA was
verified through the real Go handler on `http://127.0.0.1:8090` including the
favicon, CSP and legacy-redirect behavior. Browser tests used the compiled Go
server, not only the static PowerShell server.

## Current phase

PHASE 3 — LAUNCHER FEATURE COMPLETION + BACKEND ROBUSTNESS (COMPLETE)

The Prism minimal-core / licensing checkpoint is RESOLVED at architecture
level (DECISION-003, DECISION-004; see `docs/PRISM_MINIMAL_CORE.md` and
`docs/PRISM_ARCHITECTURE_REVIEW.md`).

The Go + SQLite backend is implemented, hardened and fully tested on `master`:
auth, account management (profile, password change, session list/revoke),
middleware (request logging, panic recovery, security headers), readiness
probe, rate limiting, and admin HWID bans. `go build`, `go vet` and
`go test ./...` pass (run them from `server/`, the Go module root; the current
baseline is 200+ passing tests). The backend exposes the processed skin-file
API under `/api/v1/skin-files` (multipart upload with a `skin` field, raw PNG,
head avatar, owner-only delete/activate) and the Pterodactyl webhook
(`POST /api/v1/webhooks/pterodactyl`, HMAC-verified, fails closed with 401 when
no secret is configured). Both are wired in `server/main.go`. Skin files are
stored in a concurrency-safe store that persists to `AUTH_SKIN_STORAGE_DIR`.
The account-scoped named-skin API stays under `/api/v1/skins`.
See `docs/DECISIONS.md` DECISION-006 and `docs/WORK_STATE.md`.

The C++/Qt6 launcher core is implemented on `master` and the build baseline is
GREEN on Windows 11, verified 2026-11-16: CMake configure and a Release build
with Qt 6.6.3 succeed, `MinecraftLauncher.exe` and every CTest binary are
produced, and `ctest` reports all tests passing (23/23 at that time). The
verified toolchain is the Qt 6.6.3 `mingw_64` package with MinGW GCC 13.1, and
the MSVC 2022 Build Tools (MSVC 19.44, Windows SDK 10.0.26100) were installed
later on 2026-11-16: the MSVC Release build is clean under `/W4 /WX` and
`ctest -C Release` also passed. Re-verified 2026-11-17 on the owner's Windows
machine with Qt 6.6.3 `msvc2019_64`, MSVC 14.52 (Visual Studio 18 Insiders),
CMake 3.30.5 and Ninja: configure and build succeed and `ctest` reports
**31/31 passing** — the suite has grown to 31 targets (memory_clamp,
system_memory, game_exit_diagnostics, friendly_error and friendly_error_paths
were added after the earlier 23-test runs, and token_storage,
resolve_libraries and resolve_assets came with the PR #8 resolver port). The warnings-as-errors gate applies to every test executable as well
(`/W4 /WX` on MSVC, `-Wall -Wextra -Werror` on GCC/Clang), so test-only code
cannot hide warnings.

Continuous integration: a self-hosted GitHub Actions runner (`debian-ai-runner`,
runner 2.338.2) runs on the LAN Debian 13 host `192.168.1.23` as a systemd
service, with labels `self-hosted, linux, x64, debian-13`.
`.github/workflows/linux-ci.yml` uses it for the Go backend suite and the
Velocity build and is the only workflow that currently produces runs.
`launcher.yml` and `backend.yml` target GitHub-hosted `windows-latest` (Option A);
both really run there since 2026-11-16 (backend: go mod verify/vet/test/build;
launcher: Qt 6.6.3 `msvc2019_64`, Configure, Build, 23/23 CTest tests). That image
is `windows-2025-vs2026`, so the Configure step detects the newest Visual Studio
generator instead of pinning VS 17, `install-qt-action` is given an explicit
`dir` (aqt installs into `<workspace>/Qt/<version>/<arch>` because it appends its
own `Qt` folder), and a failed Qt install fails the job (no `continue-on-error`
or gated steps that could skip the build). The old `ci.yml` was deleted because
it duplicated the other workflows and queued an unfiltered Windows job on every
push to `master`.

Host access and runner operational details are kept out of the repository in
`runner.env` in the repository root (git-ignored through the existing `*.env`
rule: host, SSH and root login, runner name/version/labels, systemd unit, paths
and commands). Read that file for access; never commit it or paste its values
into issues, logs or commits.

A high-level, owner-facing gap analysis and minimum publishable baseline live
in `docs/PROJECT_STATUS.md`. Owner decisions from 2026-11-16: Velocity + join
authorization is in v1; updates are hosted on GitHub Releases with a backend
redirect/backup (DECISION-007); legal publication is limited to what is legally
required, preferring slightly more (DECISION-008); metadata source is online
fetch + local cache + bundled fallback (DECISION-009); registration is
invite-only by email allowlist with a granular per-user admin permission
system; target ~200 beta users; modpack delivery is mrpack-like with signed,
tamper-checked manifests; beta labels everywhere. Windows CI now uses GitHub-hosted `windows-latest` for both `launcher.yml` and
`backend.yml` (Option A, PROJECT_STATUS §9); the owner will be alerted if the
2000 free minutes/month run out, and can switch to a self-hosted Windows runner
(Option B) at that point. The CTest suite (hwid, settings,
update_manifest, skin_model, backend_contract, settings_store, mod_scanner,
skin_cache, rotating_logger, update_checker, file_integrity, editor_config,
skin_client, download_asset, download_queue, java_detector, game_paths,
game_launcher, editor_settings_widget, login_dialog, main_window,
skin_upload_widget, update_manager) is defined in `launcher/CMakeLists.txt`.
Newer launcher work includes the update manifest parser, skin model, backend
status-code error mapping, a request timeout, and CLI filter/version options.

The Qt6 `MainWindow` is wired into `launcher/main.cpp` and embeds
`SkinPreviewWidget`, whose `skinUploadFinished` signal updates the status bar.
`MainWindow` owns the backend session: it edits and persists the backend URL
and the last username in `launcher.json`, opens `LoginDialog`, calls
`SessionManager::login/restore/logout`, and hands the resulting
`BackendAuthClient` to the skin panel. Tokens stay in the DPAPI-protected
`TokenStorage` (never in `launcher.json`), a valid restored session pushes its
access token into the client, and logout clears it. The remote-backend policy
is HTTPS-only; plain HTTP is accepted just for `localhost`/`127.0.0.1`/`::1`
so the local development backend can be used. `SkinClient` then reaches the
authenticated `/api/v1/skin-files` endpoints. That path was verified end to end
on 2026-11-16 against a locally running Go backend: login returned an access
token, the upload returned a `skin_id`, and the avatar decoded into a real
64x64 `QPixmap`. HTTPS was verified end to end on 2026-11-16: the Qt OpenSSL 3 package
(3.0.16), the Qt `tls` plugin directory and the MinGW C++ runtime are staged
next to the executable by CMake, public certificates validate through the
system trust store, and login/upload/avatar work over `https://` against a
real backend with only the packaged files present. The remaining limitation
in this area is that `GET /api/v1/skin-files/{id}` is readable by any
authenticated account (only delete/activate are owner-scoped).

The Velocity plugin builds against `velocity-api:4.2.0` with JDK 25
(`./gradlew build` produces `build/libs/minecraftserverlauncher-velocity-0.1.0.jar`).
Its join-time authorization currently reads a `minecraftserverlauncher-token`
GameProfile property, which the launcher does not populate yet; this transport
gap is documented as an open issue rather than a security relaxation.
DECISION-004 (GPLv3 distribution) and DECISION-005 (metadata source) remain
owner decisions for the Prism-derived launcher work.

## Approved architecture direction

**required Prism core + own launcher core + own backend**

NOT "full Prism Launcher + custom layer". See `docs/DECISIONS.md` DECISION-001.

- launcher foundation: Prism Launcher fork, trimmed to the functionality this project actually needs
- project-specific launcher logic stays in the project's own launcher core layer
- backend: project-owned, Go + SQLite
- Velocity plugin: project-owned
- Geyser / Floodgate / Limbo components: external integrations/dependencies
- first launcher target: Windows 10/11
- first release: one server, but the internal model supports more
- server-side verification is the security boundary
- updater and file integrity are required later

## Repository state

- Git default branch: `master`
- Remote: `origin` — private GitHub repository `Ddking-1/minecraftserverluncher`
- Tracked: governance/state documentation plus the backend under `server/` and
  the launcher core under `launcher/`, all on `master`
- Untracked by design: `reference/` (DECISION-002) and local tooling state
- Architecture documentation includes `docs/PRISM_MINIMAL_CORE.md` and
  `docs/PRISM_ARCHITECTURE_REVIEW.md` (DECISION-003/004/005)
- Backend source exists under `server/` (Go module `minecraftserverluncher/server`)
- Launcher core exists under `launcher/` (C++ / Qt6, builds and tests green **on a Windows machine with the verified Qt 6.6.3 + MSVC or MinGW GCC toolchain**). Latest local verification: 2026-11-17, Qt 6.6.3 `msvc2019_64` + MSVC 14.52 + CMake 3.30.5 + Ninja, `ctest` 31/31. The CMake configure and Release build are portable, but `ctest` requires the Qt 6.6.3 binaries and platform plugins staged next to the executables (the `stage_qt_runtime` target does this).
- Comprehensive Code Wiki: `docs/CODE_WIKI.md` covers architecture, modules,
  key classes/functions, dependencies, and run instructions.

## Main goals

- custom launcher
- common authentication
- trusted launcher path
- normal Java LimboAuth path
- Bedrock Geyser/Floodgate path
- Velocity gateway
- update system
- integrity
- skins
- web/admin
- lightweight deployment

## Important invariants

- One logical common account system
- Java and Bedrock identities are distinct
- Username alone never grants trust
- Bedrock does not use Java LimboAuth
- Trusted launch authorization must be server-verifiable
- Maximum launcher session lifetime: 30 days
- TPM is optional
- First UI exposes one server
- Internal architecture supports multiple server profiles
- RAM automatic range: 1.5–12 GB
- 512 MB safety reserve
- Protected files are verified
- Updates are atomic and rollback-capable
- Reference repositories are read-only
- Licensing decisions must be documented

## Reference sources

Located in:
reference/

They are research/reference sources, not automatically product code.

## Current implementation status

Backend (Go + SQLite, `server/`, module `minecraftserverluncher/server`) is on
`master`: account model with bcrypt password hashing, server-side session model
with expiry and revocation, device identity mapping, registration/login/
refresh/logout/logout-all, account profile/password/session management,
HWID ban administration, rate limiting, middleware (request logging, panic
recovery, security headers), `/healthz` + `/readyz`, and hardened HTTP server
timeouts with graceful shutdown. As of 2026-10-10, `server/skins` builds,
vets and tests clean (`ExtractAvatarHead` merged-line compile error fixed on
branch `fix/skins-processor-linebreak`), but `go build ./...` still fails in
`server/api`: `handlers.go` references `rateLimiter`, `newRateLimiter` and
`Server.limit`, which are defined nowhere (`ratelimit.go` exports
`RateLimiter`/`NewRateLimiter`/`HandlerMiddleware`). See `docs/WORK_STATE.md`.
Further reading: `docs/API.md`, `docs/SECURITY_REVIEW.md` and `server/README.md`.

Launcher core (`launcher/`, C++/Qt6) is on `master`: `launcher/core`
(BaseInstance, InstanceList, InstancePersistence, SkinModel, SkinCache,
SettingsStore, RotatingFileLogger, FileIntegrity, Version), `launcher/minecraft`
(MinecraftInstance, PackProfile, ModScanner), `launcher/auth`
(BackendAuthClient with a 15s timeout and status-code error mapping,
TokenStorage using Windows DPAPI with no plaintext fallback, SessionManager), `launcher/launch`
(LaunchTask/LaunchStep framework), `launcher/updates` (UpdateManifest parser,
UpdateChecker), `launcher/game` (JavaDetector, MinecraftLauncher) and
`launcher/core/FriendlyError`, the single source of plain-language Hungarian
error messages: every user-facing failure goes through
`FriendlyError::explain`, and the build passes `/utf-8` on MSVC because all UI
strings are UTF-8 Hungarian (without it the text is garbled on a machine whose
ANSI codepage is not UTF-8). Coverage is complete on the GUI side as well:
news loading, skin upload/preview, login/logout, session restore, settings
persistence and the crash-log viewer all pass their raw error strings through
`explain` and log the technical wording with `qWarning`, so the player only
ever sees Hungarian text plus a concrete action (the generic fallback no
longer echoes the raw snippet). CMake configure and `cmake --build ... --config
Release` succeed with Qt 6.6.3 under both MinGW GCC 13.1 and MSVC (2022, and
MSVC 14.52 + CMake 3.30.5 + Ninja re-verified 2026-11-17 with the full CTest
suite green),
including the `/W4 /WX` gate on the MSVC launcher target. The library and
asset steps (`ResolveLibraries`, `ResolveAssets`) download from the instance
`version.json` metadata (DECISION-005: the self-hosted backend is the metadata
source) and SHA-1 verify every file before an atomic write; any
metadata/path/hash mismatch fails the launch instead of proceeding. The CI workflow
(`.github/workflows/launcher.yml`) runs on the self-hosted Windows runner
(labels `self-hosted, windows, x64, launcher`) when one is available and on
GitHub-hosted `windows-latest` otherwise: `runner-fallback.yml` re-dispatches a
run that stays queued (`runner: hosted`) and the job skips `install-qt-action`
when the `QT_PREFIX` repository variable points at a local Qt; the artifact deploy step
locates the toolset with `vswhere` and runs
`windeployqt --compiler-runtime` inside `vcvars64` (without that the uploaded
launcher lacked the MSVC runtime), and the actions are on their Node 24
releases (`checkout` v5, `upload-artifact` v6, `setup-java` v5); a failing install now fails
the job instead of skipping Configure/Build/Test.

The Velocity plugin builds against `velocity-api:4.2.0` with JDK 25
(`./gradlew build` produces `build/libs/minecraftserverlauncher-velocity-0.1.0.jar`).
Its join-time authorization currently reads a `minecraftserverlauncher-token`
GameProfile property, which the launcher does not populate yet; this transport
gap is documented as an open issue rather than a security relaxation.
DECISION-004 (GPLv3 distribution) and DECISION-005 (metadata source) remain
owner decisions for the Prism-derived launcher work.

## Handoff rule

This file must always describe the current real repository state.

Never claim functionality that does not exist.

Batch update (2026-11-16): Velocity proxy auth (N15-N20) enforces backend
authorization synchronously on PostLogin plus a ServerPreConnectEvent guard;
Gradle build green. Invite-only beta registration via invite_allowlist
(403 "nem beta tester" when configured); Minecraft-level ban distinct from
system/HWID bans; granular admin API for roles/permissions/invites. Runnable
launcher build: build-local/Release/MinecraftLauncher.exe (MSVC Release,
windeployqt-deployed, starts under offscreen).

Batch (2026-11-16): anti-bot hardening (honeypot + per-username limiter +
sanitized X-Request-Id), account/skin approval workflows, launcher CI artifact
(windeployqt + upload-artifact), security audit doc. linux-ci green.

Batch (2026-11-16): tripwire anomaly tracking (DB + admin endpoint + recording
in honeypot/invalid-creds), per-user RBAC admin auth (adminAuthorized now
accepts bearer with admin permission/role; admin role+permission seeded),
admin web UI rewired to real bearer-token endpoints, security audit follow-up.
linux-ci green.

Batch (2026-11-16): Cloudflare Turnstile CAPTCHA (config-gated, dev mode when
unset), admin dashboard HTML (web/static/admin.html), skin-manager rewired with
Turnstile widget. linux-ci green.

Batch (2026-11-16b): security audit pass 2 — XSS-safe admin UI, PNG
decompression-bomb guard, CSP header, .gitignore secret/build artifacts,
docs/SECURITY.md, docs/BACKLOG.md (feature candidates pending approval).
linux-ci green.

Batch (rev3 build): S1 fail-closed approval/invite gates, S2 HWID device_id requirement (config-gated), S3 webhook cleanup (dead branch removed, oversized body 413), S4 image dimension guard in SniffImage, S8 feature flags + GET /api/v1/capabilities endpoint. linux-ci pending.

Audit (2026-11-17): API/DB/auth audit completed; no test bypasses, no leaked keys, no IDOR beyond intended public skins/admin scope. See SECURITY_AUDIT.md.

Batch (2026-11-17): launcher GUI features F2/F3/F5/F6/F7/F10 implemented and verified.
- F2 crash popup posts to /api/v1/crash-reports (bearer when session available).
- F3 modpack import extracts zip via Expand-Archive.
- F5 RAM slider persists global settings and overrides instance memory at launch.
- F6 offline mode persisted; affects session restore/login/auth UI.
- F7 crash log viewer opens and displays a log file.
- F10 news panel fetches and displays /api/v1/news.
- launcher CI (build + ctest) SUCCESS; backend CI SUCCESS.

- Adversarial hardening pass on F2/F3/F5/F6/F7/F10: input validation, timeouts,
  reentrancy guards, payload caps, zip pre-checks, memory clamping, and
  crash-report log sanitization on the backend.

- Memory UX: `core/SystemMemory` reports installed RAM and the allowed heap
  (installed - 2 GB). The GUI shows it as a one-line statistic and clamps any
  over-limit value with a plain-language message; the launch step refuses an
  impossible heap instead of letting the JVM fail opaquely.

- Game exit diagnostics: `core/GameExitDiagnostics` classifies a dead game
  process from its output tail and exit code; only genuine memory exhaustion
  gets the numbered Hungarian message, everything else keeps the generic exit
  log line. `LaunchTask::failed()` may follow `succeeded()` for a game that died
  after starting.

- User-facing errors: `core/FriendlyError::explain` converts technical
  launch/download errors to plain Hungarian; `LaunchTask` logs the raw reason
  and emits the friendly text.

- Cleanup: removed unused SplashScreen component (no references); build/tests unaffected.
