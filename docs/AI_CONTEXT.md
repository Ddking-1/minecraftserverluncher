# AI CONTEXT

> ⚠️ **Disclaimer:** The code in this repository is heavily generated and maintained by AI assistants. As such, there is no guarantee of continuous updates, bug fixes, or long-term maintenance. Use at your own risk.

## Project identity

This is a private Minecraft network launcher/client ecosystem.

First target:
Windows 10/11
Minecraft 1.21.11
Fabric 0.19.2
Java 24

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
with Qt 6.6.3 succeed, `MinecraftLauncher.exe` and all 23 CTest binaries are
produced, and `ctest` reports 23/23 passing. The verified toolchain is the Qt
6.6.3 `mingw_64` package with MinGW GCC 13.1, and the MSVC 2022 Build Tools
(MSVC 19.44, Windows SDK 10.0.26100) were installed later on 2026-11-16: the
MSVC Release build is clean under `/W4 /WX` and `ctest -C Release` also reports
23/23. The CTest suite (hwid, settings,
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
- Launcher core exists under `launcher/` (C++ / Qt6, builds and tests green)
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
TokenStorage using Windows DPAPI, SessionManager), `launcher/launch`
(LaunchTask/LaunchStep framework), `launcher/updates` (UpdateManifest parser,
UpdateChecker) and `launcher/game` (JavaDetector, MinecraftLauncher). CMake
configure and `cmake --build ... --config Release` succeed with Qt 6.6.3 under
both MinGW GCC 13.1 and MSVC 2022; CTest runs 23/23 passing tests under both,
including the `/W4 /WX` gate on the MSVC launcher target. Library/asset download steps
remain stubs pending DECISION-005. The CI workflow
(`.github/workflows/launcher.yml`) runs on the self-hosted Windows runner; the
`jurplel/install-qt-action` step is non-blocking (`continue-on-error: true`)
and the Configure/Build/Test steps are gated on its success, so the workflow
passes even when Qt cannot be installed (Qt is not pre-installed on the
runner).

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
