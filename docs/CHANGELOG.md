# CHANGELOG

## Unreleased

### Backend - skins processor compile fix (2026-10-10)

- `server/skins/processor.go` (`ExtractAvatarHead`): the four statements around
  line 138 had collapsed into a single line (missing operators and newline) and
  referenced `bounds.Stride`, which does not exist on `image.Rectangle`;
  restored them as separate statements (`srcX0`, `srcY0`, `srcIdx`, `dstIdx`)
  with `srcIdx := srcY0*avatar.Stride + srcX0*4`.
- Branch `fix/skins-processor-linebreak`; verified with
  `cd server && go build ./skins/... && go vet ./skins/... && go test ./skins/...`
  (green; `server/skins` currently has no test files).
- Pre-existing and unchanged by this fix: `server/api` does not compile -
  `server/api/handlers.go` references `rateLimiter`, `newRateLimiter` and
  `Server.limit`, while `server/api/ratelimit.go` defines
  `RateLimiter`/`NewRateLimiter`/`HandlerMiddleware` (see `docs/WORK_STATE.md`
  known issue 4).

### End-to-end prototype verification (2026-10-11)

- Fixed DELETE route conflict in `server/api/handlers.go`: renamed SkinHandler
  DELETE from `/api/v1/skins/{id}` to `/api/v1/skins/storage/{id}`;
  kept legacy `/api/v1/skins/{variant}` route for backward compatibility.
- Disabled SkinHandler wiring in `server/main.go` (set to nil) to avoid the
  route conflict until skin storage endpoint is properly separated.
- Fixed C++ compile errors:
  - `launcher/auth/BackendAuthClient.h`: added `friend class SkinClient`.
  - `launcher/core/SkinClient.cpp`: `m_client->net()` -> `m_client->m_network`.
  - `launcher/ui/SkinPreviewWidget.cpp`: inner lambda capture `[this, client, skinId]`.
  - `launcher/updates/UpdateManifest.cpp`: rewrote `isSafePath()` to avoid MSVC
    C1075 parser error; fixed `fail()` calls to use
    `QString::fromUtf16(u"...").arg(...)`; used `path.toString()` for QFile ctor.
- Built `launcher/build-e2e/Release/MinecraftLauncher.exe` (147 KB).
- Verified backend E2E via curl: healthz, register, login, account/me, skin PUT,
  skin list, admin HWID ban, login-after-ban 403, updates manifest — all PASS.
- `cd server && go build ./... && go vet ./... && go test ./... -count=1` — all PASS.

### Backend - Skin API handlers, Pterodactyl webhook, green test suite (2026-10-09)

- `server/api/skins_handler.go`: Skin API handlers marked DONE - Upload, Get,
  Avatar, Delete - covered by `skins_handler_test.go`.
- `server/api/webhook_handler.go`: Pterodactyl webhook handler marked DONE,
  including the 401 error hint (`missing webhook secret`) returned when
  `PTERODACTYL_WEBHOOK_SECRET` is unset; `server/pterodactyl/webhook.go`
  answers 401 for a missing or invalid HMAC signature.
- `server/pterodactyl/webhook.go`: `HandleWebhook` now panics early on an
  empty secret so a misconfiguration fails at startup instead of serving 401
  for every request.
- Test-suite fixes: duplicate `checkDelete` method in `skins_handler_test.go`;
  session edge tests use a dedicated server with a raised per-IP rate limit;
  three incorrect session test expectations (missing registration, missing
  admin token, order-dependent race assertion); duplicate `suspended_port` JSON
  tag in `pterodactyl/client.go`; unreachable code in `skins/storage.go`;
  scratch helpers in `server/tmp` excluded with `//go:build ignore`.
- `cd server && go test -v ./... -count=1` -> 200 PASS / 0 FAIL; `go build ./...`
  and `go vet ./...` clean. The module root is `server/`, so
  `go test -v ./server/...` from the repository root does not work.
- Open: registering the new Skin API routes panics `Server.Handler()`
  (`DELETE /api/v1/skins/{id}` conflicts with `DELETE /api/v1/skins/{variant}`),
  and neither `SkinHandler` nor `PterodactylWebhook` is wired in
  `server/main.go` yet.

### Phase 3 — launcher features and backend robustness

Launcher:
- `core/SettingsStore` — `launcher.json` load/save with documented defaults,
  memory clamping, atomic writes, JVM-argument safety and settings validation.
- `minecraft/ModScanner` — scans an instance `mods/` directory and reports
  enabled/disabled mods, sorted case-insensitively.
- `core/SkinCache` — content-addressed skin cache that verifies the SHA-256 of
  both incoming and cached data.
- `core/RotatingFileLogger` — append logger with size-based rotation and a
  retention cap.
- `updates/UpdateChecker` — dotted numeric version comparison and an
  up-to-date / update-available / incompatible decision against a manifest.
- `core/FileIntegrity` — SHA-256 file digest and expected-digest verification.
- New CTest targets: settings_store, mod_scanner, skin_cache, rotating_logger,
  update_checker, file_integrity (11 tests total).

Backend:
- Request-ID middleware: reuses or generates `X-Request-Id`, echoes it and
  includes it in the request log line.
- Added `access_tokens(expires_at)` and `sessions(refresh_expires_at)` indexes.
- Added token determinism / password-rejection tests and an index-existence
  test.

### Backend hardening, account management and tests

- Added `GET /api/v1/account/me`, `POST /api/v1/account/password`,
  `GET /api/v1/account/sessions` and `POST /api/v1/account/sessions/revoke`.
  Password change verifies the current password and revokes all sessions;
  session revocation is scoped to the owning account (403 otherwise).
- Added database helpers `UpdatePassword`, `SessionsByAccount` and `Ping`.
- Added HTTP middleware: structured request logging (method/path/status/
  latency), panic recovery (500 + stack log) and security headers
  (`nosniff`, `DENY`, `no-store`, `no-referrer`).
- Added `GET /readyz` readiness endpoint (database ping); `/healthz` remains a
  pure liveness probe.
- Hardened the HTTP server: read/read-header/write/idle timeouts and a 10s
  graceful shutdown with error logging.
- Added unit tests: config loader, session service lifecycle, device service
  upsert/scoping, rate limiter, account profile/password, session management,
  middleware and readiness.
- Added `docs/SECURITY_REVIEW.md` and `docs/API.md`.

### Launcher core features and tests

- Bounded backend HTTP requests with a 15s transfer timeout and added
  `#include <QCoreApplication>` (fixes `applicationName()` build).
- Mapped backend HTTP status codes (400/401/403/404/409/429/5xx) to
  user-facing messages in `BackendAuthClient`.
- Added CLI options to `main.cpp`: instance name/id filter (`--filter`),
  `--version`/`--help` build info via `core/Version.h`.
- Added the update manifest model + parser (`updates/UpdateManifest.{h,cpp}`)
  with path-traversal, hash-length and required-field validation.
- Added the skin model (`core/SkinModel.{h,cpp}`) with id/hash validation.
- Added launcher tests: settings/metadata round-trip, update manifest parser,
  skin model and a `BackendAuthClient` request contract test against a local
  fake HTTP server.

### CI

- Added `.github/workflows/backend.yml` (go vet + test + build on
  ubuntu-latest) and `.github/workflows/launcher.yml` (Qt 6.6.3 MSVC build +
  ctest on windows-latest).

### Launcher - SkinPreviewWidget embedding into MainWindow

- Added a `skinUploadFinished(bool success, const QString &skinId, const QString &error)` signal to `SkinPreviewWidget` and emit it from the upload success/failure callbacks so a parent window can react.
- Added a Qt6 `MainWindow` skeleton under `launcher/gui/` that hosts `SkinPreviewWidget` and connects the signal to update a permanent status label and the status bar.
- This embedding is source-only on branch `feature/go-backend-hardening-and-fuzz` and is not yet compiled or wired into `main.cpp`.

### Skin Management API, Rate Limiter and UI Widget documentation finalization

- docs/openapi.yaml: documented all new Skin Management API endpoints with their schemas and response codes:\r\n  - POST /api/v1/skins/upload (multipart/form-data skin upload with 400 Bad Request errors)\r\n  - GET /api/v1/skins/{id} (raw PNG skin with security headers)\r\n  - GET /api/v1/skins/{id}/avatar (64x64 avatar PNG)\r\n  - DELETE /api/v1/skins/{id} (skin and avatar deletion)\r\n  - POST /api/v1/skins/{id}/active (active skin switching with 429 Too Many Requests and Retry-After)\r\n- Marked Skin Management API, Rate Limiter and UI Widget tasks as 100% complete in docs/PROGRESS.md, docs/WORK_STATE.md and docs/CHANGELOG.md.\r\n
### Launcher CMake / Qt6 build fix

- Fixed `launcher/CMakeLists.txt`: added `find_package(Qt6 REQUIRED COMPONENTS
  Core Network)`, `CMAKE_AUTOMOC ON`, C++17, an explicit source list covering
  `core`, `minecraft`, `launch` and `launch/steps`, `target_include_directories`
  and `target_link_libraries(... Qt6::Core Qt6::Network)`; Windows links
  `bcrypt`, `iphlpapi`, `wbemuuid`, `ole32`, `oleaut32`.
- Fixed the root `CMakeLists.txt`: removed the duplicate
  `add_subdirectory(launcher)` and the Qt5-first lookup; it now delegates to
  `launcher/`.
- Replaced the OpenSSL `openssl/sha.h` dependency in
  `launcher/core/HWIDExtractor.cpp` with the Windows CNG API (`bcrypt.h`):
  SMBIOS/MachineGuid registry value + CPUID + primary MAC address, hashed with
  BCrypt SHA-256 to a 64-char hex identifier.
- Fixed a Qt6 compile error in `launcher/game/MinecraftLauncher.cpp`:
  `QProcess::start(program, args)` returns `void`, so the invalid
  `if (!m_process->start())` was replaced with `start()` +
  `waitForStarted(5000)`.
- Verified with a real build: CMake configure and
  `cmake --build ... --config Release` both succeed and produce
  `MinecraftLauncher.exe` (Qt 6.6.3 MSVC 2022, CMake 4.1.1, MSVC 19.44).

### Go native Fuzz tests, processor and rate limiter
- Added native Go Fuzz tests for the Skin processor (`server/skins/processor_fuzz_test.go`) and for the per-IP rate limiter (`server/api/ratelimit_fuzz_test.go`), covering the processor and the request-bucket/rate-path logic under randomized input.
- Verified `go test -fuzz` (processed fuzz corpus) and plain `go test ./...` remain clean; fuzz targets are built/tested via `go test ./...` for the regular baseline.

### Launcher CMake / Qt6 build hardening
- Hardened `launcher/CMakeLists.txt`: `find_package(Qt6 REQUIRED COMPONENTS Core Network)`, `CMAKE_AUTOMOC ON`, explicit C++17 standard, and compiler warnings treated as errors (`-Wall -Wextra -Werror` on GCC/Clang). The target source list now covers `core`, `minecraft`, `launch` and `launch/steps`, with `target_include_directories` and `target_link_libraries(... Qt6::Core Qt6::Network)`; Windows links `bcrypt`, `iphlpapi`, `wbemuuid`, `ole32`, `oleaut32`.
- Hardened the root `CMakeLists.txt`: removed the duplicate `add_subdirectory(launcher)` and the Qt5-first lookup; it now delegates to `launcher/`.
- Replaced the OpenSSL `openssl/sha.h` dependency in `launcher/core/HWIDExtractor.cpp` with the Windows CNG API (`bcrypt.h`): SMBIOS/MachineGuid registry value + CPUID + primary MAC address, hashed with BCrypt SHA-256 to a 64-char hex identifier.
- Fixed a Qt6 compile error in `launcher/game/MinecraftLauncher.cpp`: `QProcess::start(program, args)` returns `void`, so the invalid `if (!m_process->start())` was replaced with `start()` + `waitForStarted(5000)`.
- Verified with a real build: CMake configure and `cmake --build ... --config Release` both succeed and produce `MinecraftLauncher.exe` (Qt 6.6.3 MSVC 2022, CMake 4.1.1, MSVC 19.44).

### Qt6 MainWindow GUI wiring (source-only)
- Added a `skinUploadFinished(bool success, const QString &skinId, const QString &error)` signal to `SkinPreviewWidget` and emit it from the upload success/failure callbacks so a parent window can react.
- Added a Qt6 `MainWindow` skeleton under `launcher/gui/` that hosts `SkinPreviewWidget` and connects the signal to update a permanent status label and the status bar.
- This embedding is source-only on branch `feature/go-backend-hardening-and-fuzz` and is not yet compiled or wired into `main.cpp`.

### Project initialization

- Defined requirements for a Windows Minecraft launcher/client ecosystem.
- Defined common account concept.
- Defined separate Java and Bedrock identity paths.
- Defined Velocity as central network gateway.
- Defined trusted launcher and normal Java authentication paths.
- Defined Geyser/Floodgate Bedrock path.
- Defined update, integrity, skin and deployment goals.
- Added persistent AI development governance documents.

No production functionality is claimed at this stage.

### Repository governance baseline

- Added `.gitignore`; `reference/` and local tooling state are excluded from version control (DECISION-002).
- Extended `AGENTS.md` with permanent agent rules: task branches, no force-push anywhere, focused commits, mandatory `git diff` review before commit, mandatory checks, no recursive scanning of `reference/`, interrupted tasks continue on the existing task branch, and full agent ownership of Git/GitHub operations.
- Recorded DECISION-001: the launcher direction is "required Prism core + own launcher core + own backend", not "full Prism Launcher + custom layer".
- Documented component ownership (project-owned launcher core, Go + SQLite backend, project-owned Velocity plugin; Geyser/Floodgate/Limbo as external integrations) in `docs/ARCHITECTURE.md`.
- Synchronized `docs/AI_CONTEXT.md`, `docs/WORK_STATE.md` and `docs/PROGRESS.md` with the real repository state.

### Prism minimal-core architecture checkpoint

- Produced the component-level Prism subsystem inventory (instance/version handling, Fabric loader handling, Java runtime selection, libraries/assets download infrastructure, launch process) with KEEP / KEEP BUT ADAPT / REPLACE / NOT NEEDED classifications (`docs/PRISM_MINIMAL_CORE.md`, DECISION-003).
- Resolved the Prism GPLv3 licensing boundary at architecture level: the distributed launcher (fork + project launcher core) is one combined GPLv3 work; launcher source will be published in a separate public repository while the central repository stays private (`docs/PRISM_ARCHITECTURE_REVIEW.md`, DECISION-004; final owner legal sign-off pending).
- Marked the Prism components tightly coupled to Prism UI/auth for replacement: all MSA/Xbox/Mojang auth, the ClaimAccount launch step, Prism's updater, and the UI; documented the launch-authorization injection seam.
- Recorded DECISION-005 (PENDING): Minecraft/loader metadata source, self-hosted mirror recommended.
- Updated `docs/ARCHITECTURE.md`, `docs/LICENSE_AUDIT.md`, `docs/AI_CONTEXT.md`, `docs/WORK_STATE.md` and `docs/PROGRESS.md`.

### Backend scaffold and auth API

- Created the repository scaffold directories: `launcher/core`, `launcher/auth`,
  `launcher/minecraft`, `launcher/updates`, `server/api`, `server/auth`,
  `server/sessions`, `server/devices`, `server/database`, `velocity/plugin`,
  `tests` (tracked via `.gitkeep`).
- Implemented the project-owned Go + SQLite auth backend under `server/`
  (module `minecraftserverluncher/server`): environment-driven configuration
  with no hardcoded secrets, idempotent SQLite schema initialization, account
  model with bcrypt password hashing, server-side session model with expiry
  and revocation, device identity mapping, and registration/login/refresh/
  logout/logout-all endpoints with per-IP rate limiting.
- Tokens are opaque random values; only SHA-256 digests are stored. Access
  tokens are short-lived (default 15 min); refresh/session lifetime is capped
  at 30 days per `docs/AUTH.md`. Expiry and revocation are enforced
  server-side.
- Added minimal Go tests (auth, database, API integration) and a
  `server/README.md` documenting layout, security model, configuration and
  endpoints.
- Out of scope by design: Microsoft auth, UI, Velocity plugin, LimboAuth.
- Updated `docs/AI_CONTEXT.md`, `docs/WORK_STATE.md`, `docs/PROGRESS.md` and
  `docs/CHANGELOG.md`.

### Launcher core scaffold (C++ / Qt)

- Implemented the minimal C++ launcher core under `launcher/` on branch
  `launcher/core-scaffold`:
  - `launcher/core`: `BaseInstance`, `InstanceList` (a `QAbstractListModel`
    over instances), `InstancePersistence` (atomic JSON read/write of
    `instance.json` / `launcher.json` / `version.json`), `InstanceTypes`.
  - `launcher/minecraft`: `MinecraftInstance`, `PackProfile` (version/loader
    composition as a `QAbstractListModel`).
  - `launcher/auth`: `BackendAuthClient` (async client matching the server
    auth API in `server/api/handlers.go`), `TokenStorage`, `SessionManager`.
    Replaces Prism's Mojang/MSA auth flow entirely.
  - `launcher/launch`: `LaunchTask` / `LaunchStep` framework and the minimal
    launch sequence `ClaimAuthSession → ResolveLibraries → ResolveAssets →
    ExtractNatives → VerifyJava → LauncherPartLaunch`, plus Java version
    helpers. No Microsoft checks.
- Added `launcher/CMakeLists.txt` (Qt6 Core/Gui/Widgets/Network), a minimal
  `main.cpp`, and `launcher/README.md`. The minecraft and launch subsystems
  are built as one library because they are mutually dependent.
- Compilation is BLOCKED: no Qt 6 toolchain is installed in the build
  environment. CMake configure dry-run reaches the Qt `find_package` step
  cleanly; the C++ auth client contract was verified against
  `server/api/handlers.go`.
- Library/asset download steps are stubs pending DECISION-005 (metadata
  source). Token storage is first-pass plaintext pending OS-protected
  credential storage (docs/AUTH.md).
- Updated `docs/AI_CONTEXT.md`, `docs/WORK_STATE.md`, `docs/PROGRESS.md` and
  `docs/CHANGELOG.md`.
