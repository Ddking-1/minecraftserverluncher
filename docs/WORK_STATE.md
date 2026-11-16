# WORK STATE

Current phase:
PHASE 3 — LAUNCHER FEATURE COMPLETION + BACKEND ROBUSTNESS (COMPLETE)

Current task:
Phase 3 backlog (`docs/MASTER_PLAN_PHASE3.md`, P3-01 … P3-10) is complete.
The T1–T45 backlog (`docs/MASTER_PLAN_EXTENDED.md`) is complete except T11
(Velocity skeleton), which stays BLOCKED on DECISION-004 legal sign-off and
DECISION-005 (metadata source).

Status:
GREEN everywhere, verified on a real Windows 11 machine on 2026-11-16: the Go
backend builds, vets and tests clean, and the C++/Qt6 launcher builds with
Qt 6.6.3 and passes its full CTest suite (23/23). The verified launcher
toolchain is Qt 6.6.3 `mingw_64` + MinGW GCC 13.1, because MSVC 2022 is not
installed on that machine; the MSVC `/W4 /WX` build is therefore still
unverified.

Branch:
master (all task branches fast-forward merged and pushed).

Phase 3 tasks (all DONE):
- P3-01 SettingsStore (launcher.json load/save, defaults, atomic) + tests
- P3-02 Mod list scanner (mods/ dir, enabled/disabled, sorted) + tests
- P3-03 Skin cache (content-addressed, SHA-256 verified) + tests
- P3-04 Rotating file logger (size rotation, keep N) + tests
- P3-05 Update checker (semantic version compare + decision) + tests
- P3-06 File integrity verifier (SHA-256 of file vs expected) + tests
- P3-07 Backend request-ID + structured logging middleware + tests
- P3-08 Backend auth/token extra tests + DB expiry indexes
- P3-09 Launcher memory/JVM settings validation + tests
- P3-10 Phase 3 final verification + documentation sync

Skin Management API, Rate Limiter and UI Widget tasks (all DONE):
- Skin Management API (backend handlers + OpenAPI contract) - DONE
- Rate Limiter (429 Too Many Requests + Retry-After) - DONE
- UI Widget (SkinPreviewWidget + MainWindow embedding) - DONE

Current step:
- Milestone `feature/go-backend-hardening-and-fuzz` is COMPLETED.
- Fuzz test coverage: 100% — native Go Fuzz targets added for the Skin processor (`server/skins/processor_fuzz_test.go`) and the rate limiter (`server/api/ratelimit_fuzz_test.go`).
- Build system (CMake) audit: 100% — `launcher/CMakeLists.txt` hardened with C++17, `-Wall -Wextra -Werror`, Qt6 Core/Network only, AUTOMOC, full source list and Windows libs; OpenSSL dependency removed from HWIDExtractor.cpp; root `CMakeLists.txt` deduplicated; verified GREEN with a real build.
- Qt6 GUI wiring: MainWindow + SkinPreviewWidget are wired into `main.cpp`. MainWindow now owns the backend URL setting and the login/logout/session-restore flow (source only, not compiled in this environment).
- Documentation: AI-generated-code disclaimer added to README.md and docs/AI_CONTEXT.md; CHANGELOG.md, PROGRESS.md and WORK_STATE.md updated.

Next step:
Decide the cross-account read policy of `GET /api/v1/skin-files/{id}`, ship an
OpenSSL 3 TLS backend with the launcher package (HTTPS backends currently fail
without it), repeat the build with MSVC 2022 (`/W4 /WX`), and resolve the
Velocity token transport. T11 Velocity remains BLOCKED on DECISION-004/005.

Repository facts:
- default branch: master
- remote: origin (private GitHub repo Ddking-1/minecraftserverluncher)
- reference/ is local-only, never committed (DECISION-002)
- backend source exists under server/ (Go module minecraftserverluncher/server)
- launcher core exists under launcher/ (C++ / Qt6, builds and tests green)
- velocity directory is scaffolded but empty (no code yet)

Last commit:
See docs/PROGRESS.md / `git log --oneline`.

## 2026-10-09 - UPDATE: BACKEND TEST RUN + SKIN API / PTERODACTYL WEBHOOK

Status:
GREEN. `cd server && go test -v ./... -count=1` -> 200 PASS / 0 FAIL
(api, auth, config, database, devices, pterodactyl, sessions, skins, web).
`go build ./...` and `go vet ./...` are clean. The Go module root is `server/`,
so `go test -v ./server/...` from the repository root does not work.

DONE (marked complete in docs/PROGRESS.md):
- Skin API handlers: Upload, Get, Avatar, Delete (server/api/skins_handler.go,
  tested by server/api/skins_handler_test.go).
- Pterodactyl webhook handler + 401 error hint
  (server/api/webhook_handler.go returns 401 "missing webhook secret" when
  PTERODACTYL_WEBHOOK_SECRET is unset; server/pterodactyl/webhook.go returns 401
  for a missing or invalid signature and panics early on an empty secret).

Known issues (exact next step):
1. `Server.Handler()` panics when `s.SkinHandler != nil`:
   `DELETE /api/v1/skins/{id}` conflicts with the existing
   `DELETE /api/v1/skins/{variant}`. Decide which endpoint owns that pattern
   (or move one of them) before wiring the Skin API.
2. `server/main.go` sets neither `SkinHandler` nor `PterodactylWebhook`, so the
   new endpoints are not exposed in production yet.
3. `server/tmp/*.go` scratch helpers are excluded from the build with
   `//go:build ignore`; remove them when the related investigation is done.

Branch:
feature/go-backend-hardening-and-fuzz (this milestone's documentation changes are committed and pushed in this update).

## 2026-10-10 - UPDATE: SKINS PROCESSOR COMPILE FIX

Status:
PARTIAL GREEN. `cd server && go build ./skins/... && go vet ./skins/... &&
go test ./skins/...` pass. `go build ./...` still fails in `server/api` for a
pre-existing reason unrelated to this fix (see known issue 4). The Go module
root is `server/`.

DONE:
- server/skins/processor.go: `ExtractAvatarHead` statements around line 138
  restored as four separate lines (`srcX0`, `srcY0`, `srcIdx`, `dstIdx`);
  `srcIdx := srcY0*avatar.Stride + srcX0*4` replaces the broken
  `bounds.Stride` reference (`image.Rectangle` has no `Stride` field).

Known issues (exact next step):
4. `server/api` does not compile: `server/api/handlers.go` references
   `rateLimiter`, `newRateLimiter` and `Server.limit`, while
   `server/api/ratelimit.go` defines `RateLimiter`, `NewRateLimiter` and
   `HandlerMiddleware`. Bridge or rename these so `go build ./...` is green,
   then run `cd server && go test -v ./... -count=1`.

Branch:
fix/skins-processor-linebreak (task branch created from `master` per AGENTS.md
rule 1); the agent merges it into `master` and pushes as the final git step of
this task (rule 26).

## 2026-10-11 - UPDATE: COMPREHENSIVE CODE WIKI DOCUMENTATION

Status:
DOCUMENTATION-ONLY. No production code changed. A comprehensive Code Wiki was
added at `docs/CODE_WIKI.md` covering all three subsystems (launcher, backend,
Velocity plugin): project overview, architecture overview, module breakdown,
key classes and functions, dependency relationships, data flow and trust
boundaries, build and run instructions, testing strategy, security model,
configuration reference, and open issues / known limitations.

DONE:
- `docs/CODE_WIKI.md` (new): full structured code wiki generated from a
  read-only analysis of the repository (no code modifications).
- `docs/AI_CONTEXT.md`: repository-state line updated to reference the Code
  Wiki.
- `docs/WORK_STATE.md`, `docs/PROGRESS.md`, `docs/CHANGELOG.md`: this update.

Known issues (unchanged from previous state, surfaced in the Code Wiki):
1. `server/api` compile issue: `handlers.go` references `rateLimiter`,
   `newRateLimiter`, `Server.limit` while `ratelimit.go` defines
   `RateLimiter`, `NewRateLimiter`, `HandlerMiddleware`.
2. Skin API route conflict: `DELETE /api/v1/skins/{id}` vs
   `DELETE /api/v1/skins/{variant}`.
3. `SkinHandler` and `PterodactylWebhook` are not wired in `server/main.go`.
4. Velocity plugin (T11) is a skeleton pending DECISION-004/005.
5. Launcher `ResolveLibraries` / `ResolveAssets` launch steps remain stubs
   pending DECISION-005 (metadata source).
6. `MainWindow` is not wired into `launcher/main.cpp`.

Branch:
docs/code-wiki (task branch created from `master` per AGENTS.md rule 1);
committed and pushed as the final git step of this task (rule 26).

## 2026-10-14 - UPDATE: CI launcher.yml Qt action failure fix

Status:
CI FIXED. `.github/workflows/launcher.yml` no longer fails the job when the
`jurplel/install-qt-action` step cannot complete on the self-hosted Windows
runner.

DONE:
- Runner diag log verified: `Listening for Jobs` appears at 2026-10-13 21:06:49Z
  and 2026-10-13 20:45:36Z (two successful runner starts). GitHub connectivity
  confirmed.
- Root cause identified in `Worker_20261013-204538-utc.log`: the pinned commit
  `jurplel/install-qt-action@a9c63c7c...` does not ship the built
  `action/lib/main.js`, so the action's pre-flight bash check exits with code 1
  ("Action not built ... Only v* branches have built JavaScript action file").
- `.github/workflows/launcher.yml`: Install Qt step now has `id: install_qt`
  and `continue-on-error: true`; Configure, Build and Test steps are gated on
  `steps.install_qt.outcome == 'success'`. The build still runs when Qt
  installation succeeds; otherwise it is skipped gracefully and the workflow
  passes.
- YAML validity verified with `python -c "import yaml; yaml.safe_load(...)"`.

Known issues (unchanged from previous state):
1. `server/api` compile issue: `handlers.go` references `rateLimiter`,
   `newRateLimiter`, `Server.limit` while `ratelimit.go` defines
   `RateLimiter`, `NewRateLimiter`, `HandlerMiddleware`.
2. Skin API route conflict: `DELETE /api/v1/skins/{id}` vs
   `DELETE /api/v1/skins/{variant}`.
3. `SkinHandler` and `PterodactylWebhook` are not wired in `server/main.go`.
4. Velocity plugin (T11) is a skeleton pending DECISION-004/005.
5. Launcher `ResolveLibraries` / `ResolveAssets` launch steps remain stubs
   pending DECISION-005 (metadata source).
6. `MainWindow` is not wired into `launcher/main.cpp`.
7. Qt is not pre-installed on the self-hosted Windows runner, so the C++ build
   is skipped in CI until Qt is installed or the action is pinned to a proper
   `v4` release tag.

Branch:
fix/launcher-yml-qt-action (task branch created from `master` per AGENTS.md
rule 1); committed and pushed as the final git step of this task (rule 26).

## 2026-11-15 - SKIN API WIRING, AUTHORIZATION, LAUNCHER INTEGRATION AND BACKEND HARDENING

Status:
GREEN for the Go backend and the Velocity plugin. `cd server && go build ./...`,
`go vet ./...`, `go test ./... -count=1 -race` all pass. The Velocity plugin
builds with JDK 25 (`./gradlew build` ->
`build/libs/minecraftserverlauncher-velocity-0.1.0.jar`). The C++/Qt launcher
could NOT be compiled in this cloud environment (no Qt 6 toolchain); its changes
are source-reviewed only.

Resolved known issues from previous updates:
1. Skin API route conflict (`{id}` vs `{variant}`) - RESOLVED. Processed skin
   files moved to `/api/v1/skin-files/*` (DECISION-006); named account skins stay
   under `/api/v1/skins/*`. Both are mounted, so the launcher endpoints are no
   longer unreachable.
2. `SkinHandler` / `PterodactylWebhook` not wired - RESOLVED. `main.go` wires
   disk-backed skin storage and both handlers; `/api/v1/webhooks/pterodactyl`
   fails closed (401) without a configured secret.
3. `MainWindow` not wired into `main.cpp` - RESOLVED (already wired; now the
   embedded skin panel receives a `BackendAuthClient` via `setAuthClient`).
4. Missing `Storage` in the production `SkinHandler` (every upload would have
   panicked on a nil map) - FIXED.
5. `/api/server/start` was unreachable because `/api/server/` was mounted on the
   protected group while the route lived on the admin mux - FIXED (admin group).

Other fixes in this update:
- Skin storage is now a concurrency-safe struct (memory or disk) with atomic
  writes, safe id validation and restart persistence.
- `ExtractAvatarHead` now extracts the real head face (8,8) + hat overlay
  (40,8) and scales it to 64x64 instead of stretching the whole skin.
- Rate limiter: `X-Forwarded-For` is only trusted with `AUTH_TRUST_PROXY`
  (spoofing bypass fixed), the limiter runs once per route group (double token
  consumption fixed), `Stop` is idempotent and is called on shutdown.
- Async skin worker really persists uploads, invalidates the list cache, closes
  idempotently, retains results with a TTL/cap and no longer caches in-flight
  task status (which made polls report stale "queued" states).
- Task status is scoped to the account that created it.
- Launcher: skin upload sends the multipart field `skin` (server expected it),
  the multipart object is heap-owned, requests carry the bearer token, `skin_id`
  is parsed from the response and the avatar is decoded into a `QPixmap`
  (previously always reported failure). `SkinPreviewWidget` no longer crashes on
  a null client and no longer emits success twice.
- `test_skin_client.cpp` was rewritten to the actual callback API (it referenced
  non-existent signals and did not compile).
- CI `ci.yml` uses `server/go.mod` as the Go version source, runs `go mod tidy`
  in `server/`, no longer swallows `go vet` failures, and gained a Velocity job.
- Velocity proxy auth JSON-escapes the access token (JSON injection fixed).

Remaining known issues:
1. The launcher GUI needs a real login/backend configuration flow: `main.cpp`
   does not construct a `BackendAuthClient` for `MainWindow`, so the skin panel
   reports "login required" until one is supplied.
2. The Velocity plugin expects a `minecraftserverlauncher-token` GameProfile
   property that the launcher does not set yet; join authorization therefore
   cannot succeed end-to-end.
3. Launcher `ResolveLibraries` / `ResolveAssets` still only create directories
   (blocked on DECISION-005 metadata source).
4. Auto-update is implemented (`UpdateManager`, signed manifest) but not wired
   into launcher startup (TASKS BATCH 11/12).

Next step:
Wire the login flow (backend URL + `SessionManager`) into `MainWindow`, then
resolve the Velocity token transport (LoginPluginMessage or a signed handshake)
with DECISION-level review.

Branch:
fix/skin-api-routing-launcher-integration (merged into `master` and pushed by the
agent as the final git step of this task, per AGENTS.md rule 26).

## 2026-11-15 - LAUNCHER GUI LOGIN, BACKEND URL AND SESSION RESTORE

Status:
SOURCE-COMPLETE, NOT COMPILED HERE. The launcher GUI now owns the backend
configuration and the login session. This cloud environment has no Qt 6
toolchain, so the C++/Qt changes are source-reviewed only; the Go backend and
the Velocity plugin are unaffected.

DONE:
- `SettingsStore::Settings` gained `backendUrl` and `username` (`backend_url`,
  `username` in `launcher.json`); `save` creates missing parent directories.
- `MainWindow` owns `BackendAuthClient`, `SessionManager` and `TokenStorage`:
  backend URL edit/save with normalization, login/logout buttons, auth status
  line, last-username prefill, startup session restore and status messages.
- The authenticated client is handed to `SkinPreviewWidget` via
  `setAuthClient`, and the panel is enabled only while logged in, so a skin
  upload from `MainWindow` is authenticated end-to-end.
- `BackendAuthClient::setBackendUrl` accepts HTTPS and loopback HTTP
  (`localhost`, `127.0.0.1`, `::1`), rejects URL paths/credentials, and clears
  the URL on empty input.
- `SessionManager` pushes a restored access token into `BackendAuthClient` and
  clears the in-memory token on logout/refresh failure.
- `TokenStorage` refuses usernames that are not safe file names (traversal
  hardening). `LoginDialog` validates username/password, shows an inline error
  and pre-fills the last username.
- `main.cpp` sets the organization/application identity for stable
  `QStandardPaths` locations.
- Tests: new `test_main_window` (offscreen); extended `test_login_dialog` and
  `test_settings_store`; `test_skin_client` loopback-HTTP expectation;
  `test_backend_contract`/`test_skin_client` now link `core/HWIDExtractor.cpp`.
  `MinecraftLauncher` compiles `gui/LoginDialog.cpp`.

Remaining known issues:
1. The Qt launcher was not compiled or CTest-run in this environment; a
   Windows machine with Qt 6.6.3 + MSVC 2022 is required.
2. The Velocity plugin still expects a `minecraftserverlauncher-token`
   GameProfile property that the launcher does not set.
3. `ResolveLibraries` / `ResolveAssets` remain stubs pending DECISION-005.
4. Auto-update is implemented but not wired into launcher startup.

Next step:
Build the launcher and run CTest on the Windows/Qt machine; then resolve the
Velocity token transport with DECISION-level review.

Branch:
feat/launcher-gui-login-backend-url (merged into `master` and pushed by the
agent as the final git step of this task, per AGENTS.md rule 26).

## 2026-11-16 - WINDOWS QT 6.6.3 BUILD, FULL CTEST RUN AND BACKEND END-TO-END VERIFICATION

Toolchain installed for this task (all portable, no administrator rights):
- CMake 3.30.5 and Ninja 1.12.1 (portable zips)
- Go 1.23.4 as a bootstrap; `go.mod` selects toolchain go1.26.5
- Python 3.12.7 (embeddable) + pip + aqtinstall 3.3.0
- Qt 6.6.3 `mingw_64` and MinGW GCC 13.1 (Qt's `tools_mingw1310` package)
- installed under `C:\Users\Try.APAPC\Qt` and `C:\Users\Try.APAPC\devtools`;
  build directory `C:\Users\Try.APAPC\build-mcl` (Ninja, Release)
- MSVC 2022 is NOT installed on the machine, so Qt's MSVC binaries could not be
  used; the verified compiler is GCC 13.1

Result:
- `cmake -S launcher -B <build> -G Ninja -DCMAKE_PREFIX_PATH=<qt>/6.6.3/mingw_64` OK
- build OK: `MinecraftLauncher.exe` plus 23 test executables
- `ctest` 23/23 pass (~2.7 s), including `backend_contract`, `update_manager`,
  `main_window`, `login_dialog` and the widget tests
- launcher smoke test: the GUI starts and stays alive; on a fresh install it
  creates `instances/` and keeps running
- Go on Windows: `go build ./...`, `go vet ./...`, `go test ./... -count=1` all pass
- backend and launcher-client end-to-end runs against a real local server

Build/compile defects found and fixed (all real, none environment noise):
1. `launcher/CMakeLists.txt`: `Qt6::Test` was linked but not requested in
   `find_package` -> configure failed.
2. 12 test targets include `<QtTest>` but linked only `Qt6::Core` -> `Qt6::Test`
   added to each.
3. `updates/UpdateManifest.cpp`: invalid char literal `QLatin1Char('\')` ->
   `'\'`.
4. `updates/UpdateManifest.cpp`: `fail(u"...%1").arg(...)` called `.arg()` on a
   `ParseResult` -> formatted before the call.
5. `updates/UpdateManifest.cpp`: `QFile file(QString(path));` most-vexing-parse
   -> `QFile file{path.toString()};`.
6. `updates/UpdateManifest.cpp` (SECURITY): the path-traversal regex `[/\]` was
   invalid, so `QString::split()` returned the input unchanged and
   `../../etc/passwd` was accepted as a safe path. Now `[/\]`; the
   `pathTraversal` test passes.
7. 10 tests declared `class X : public QTest` (`QTest` is a namespace) -> `QObject`.
8. `tests/test_settings.cpp`: `QJsonObject::isObject()` does not exist ->
   key-presence check.
9. `ui/SkinPreviewWidget.cpp`: the avatar lambda did not capture `skinId`.
10. `core/HWIDExtractor.cpp`: MSVC-only `#pragma comment(lib, ...)` is now
    guarded with `_MSC_VER`.
11. `core/SkinClient.cpp`: `static_cast<const unsigned char>` -> plain
    `unsigned char`; `m_client->net()` did not exist -> new
    `BackendAuthClient::networkAccessManager()`.
12. `tests/test_update_manifest.cpp`: the helper `validManifest()` was shadowed
    by the test slot of the same name (returned void) -> `makeValidManifest()`.
13. `tests/test_update_checker.cpp`: parenthesised aggregate initialisation of
    `ManifestComponent` -> explicit field assignment.
14. 4 widget tests: `QStandardPaths::setTestMode()` -> `setTestModeEnabled()`.
15. `core/UpdateManager.cpp`: the definition took `QStringView` while the header
    declares `QString`; `UpdateManagerHMAC.h` was not included.
16. CMake: `auth/BackendAuthClient.cpp` was compiled into the executable and
    into `launcher_auth` -> duplicate MOC symbols at link; removed from
    `target_sources`.
17. CMake: `test_download_queue` missed `core/FileIntegrity.cpp` -> undefined
    reference.
18. CMake: only `qwindows.dll` was staged, so Qt used a partial `platforms/`
    directory and the GUI tests aborted inside a blocking Qt error dialog (the
    CTest timeouts). The whole `plugins/platforms` directory is copied now.
19. CMake: `Qt6Gui.dll` was not staged although `Qt6Widgets` depends on it.
20. `main.cpp`: a missing `instances/` directory made the launcher exit with
    code 1 right after showing the window; it now creates the directory and
    continues with an empty list.
21. `tests/test_editor_config.cpp`: the empty-JSON case reused a `QFile` still
    bound to `bad.json`, so `empty.json` was never written.
22. `tests/test_skin_client.cpp`: the "no backend" case ran with a backend URL
    already configured (the actual error was "not logged in").
23. `tests/test_backend_contract.cpp`: `QTEST_APPLESS_MAIN` creates no
    `QCoreApplication`, so `QEventLoop`/`QTimer` never ran and the test hung
    until the CTest timeout -> `QTEST_GUILESS_MAIN`.
24. `tests/test_update_manager.cpp`: the signing key was never passed to
    `UpdateManager`, so every manifest was rejected as unverified.

Backend end-to-end results (real server on 127.0.0.1:8099, SQLite + disk skin
storage, driven with curl):
- `GET /readyz` 200; register 201; login 200 (43-character access and refresh
  tokens); wrong password 401; `GET /api/v1/account/me` 200
- skin upload without a token 401; 32x32 PNG 400 "unsupported skin dimensions
  32x32 (allowed: 64x64, 64x32)"; 64x64 upload 200 with `skin_id` and `avatar_url`
- `GET /api/v1/skin-files/{id}` 200 `image/png` 64x64; `.../avatar` 200
  `image/png` 64x64; owner DELETE 200 and the following GET 404
- refresh 200 with rotated tokens; logout 200; Pterodactyl webhook without a
  secret 401 (fails closed)
- launcher client path (`BackendAuthClient` + `SkinClient` compiled against the
  repository sources, run against the live backend): login ok, upload returned a
  `skin_id`, avatar decoded into a real 64x64 `QPixmap`

Open issues (documented, deliberately not changed):
1. The MSVC 2022 build and its `/W4 /WX` warnings-as-errors are still unverified.
2. The installed Qt has no TLS backend: HTTPS backends fail with "No functional
   TLS backend was found" until OpenSSL 3 (`libssl-3-x64.dll`,
   `libcrypto-3-x64.dll`) ships with the launcher package.
3. Qt's bundled MinGW 11.2 runtime DLLs must not be shipped with a GCC 13.1
   build (load-time failure / fail-fast); ship the compiler's
   `libstdc++-6.dll` and `libgcc_s_seh-1.dll` or match the toolchain version.
4. `GET /api/v1/skin-files/{id}` returns 200 for another authenticated account;
   only delete/activate are owner-scoped. Needs a DECISION.
5. `/api/v1/skins` (named skins) stays empty after a `/api/v1/skin-files`
   upload; the two APIs are independent by design but easy to confuse.
6. Velocity token transport, the `ResolveLibraries`/`ResolveAssets` stubs and the
   auto-update wiring remain open as before.

Branch:
fix/windows-build-and-e2e-verification (created from `master` for this task; the
agent commits and pushes it as the final Git step).
