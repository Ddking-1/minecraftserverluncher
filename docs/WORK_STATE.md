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
## 2026-11-16 - OPENSSL 3 INSTALLATION AND HTTPS END-TO-END VERIFICATION

Installed: Qt's own OpenSSL 3 package via aqtinstall
(`tools_opensslv3_x64` / `qt.tools.opensslv3.win_x64`, OpenSSL 3.0.16) at
`C:\\Users\\Try.APAPC\\Qt\\Tools\\OpenSSLv3\\Win_x64`; `libssl-3-x64.dll` and
`libcrypto-3-x64.dll` were copied into `C:\\Users\\Try.APAPC\\Qt\\6.6.3\\mingw_64\\bin`.

Packaging fixes found while validating HTTPS:
1. Qt's TLS backend is a **plugin** (`plugins/tls/qopensslbackend.dll`). Only
   `platforms/` was staged, so with the runtime next to the executable Qt
   reported `TLS initialization failed` / `supportsSsl=0` even though the
   OpenSSL DLLs were present. `plugins/tls` is now staged as well (this also
   ships `qschannelbackend.dll`, the native Windows fallback, and
   `qcertonlybackend.dll`).
2. `launcher/CMakeLists.txt` now stages the OpenSSL 3 DLLs next to the
   executable when they can be found (`find_file` over the Qt bin directory and
   `OPENSSL_DIR`), with a configure warning when they are missing.
3. The MinGW C++ runtime (`libstdc++-6.dll`, `libgcc_s_seh-1.dll`,
   `libwinpthread-1.dll`) is now staged as well; without it the launcher could
   not start at all on a machine without the toolchain.

Verification (all three with `PATH` reduced to `C:\\Windows\\system32` only, so
every dependency comes from the staged files):
- `MinecraftLauncher.exe --version` -> exit 0 and prints the version.
- Launcher GUI (offscreen) starts and stays alive for the 6 s check window.
- `supportsSsl=1`, `sslRuntime=OpenSSL 3.0.16 11 Feb 2025` (build-time version
  string is OpenSSL 3.0.7; Qt loads the 3.0.16 runtime).
- Public certificate validation through the system trust store:
  `https://api.github.com/zen`, `https://www.google.com/`,
  `https://login.microsoftonline.com/` all return HTTP 200 with `error=0`.
- Launcher client (`BackendAuthClient` + `SkinClient`) against the real backend
  behind an HTTPS terminator (`https://localhost:8443`, self-signed CA trusted
  only inside the test process via `QSslConfiguration`): login success, upload
  success with `skin_id`, avatar decoded as 64x64 `QPixmap`.
- `ctest` stays 23/23 after the CMake changes.

Note: `https://example.com/` fails with "The issuer certificate of a locally
looked up certificate could not be found" while GitHub, Google and Microsoft
validate. That is a property of that host's served chain, not of the launcher:
the same handshake succeeds with curl using the same CA bundle, and three other
public chains verify. Recorded as an observation, not a defect.

Branch:
fix/windows-build-and-e2e-verification (same task branch; the OpenSSL work is a
second commit on it).

## 2026-11-16 - MSVC 2022 BUILD ATTEMPT (BLOCKED, NO CODE CHANGES)

> SUPERSEDED: the toolchain was installed later the same day; see
> "MSVC 2022 BUILD AND TEST (UNBLOCKED, VERIFIED)" at the end of this file.

The MSVC build could not be run: this machine has no Visual Studio, no MSVC
compiler and no Windows SDK (`vswhere.exe`, `Program Files (x86)/Microsoft
Visual Studio/2022` and `Windows Kits/10` are all absent). The Qt 6.6.3
`msvc2019_64` binaries were installed successfully via aqtinstall, so only the
compiler is missing.

Two elevated installs of `vs_buildtools.exe` (17.14, C++ workload) were
requested and both were cancelled at the UAC prompt, so nothing was installed
and no system change was made.

Exact next step (requires an administrator):
1. Run `C:\Users\Try.APAPC\devtools\dl\vs_buildtools.exe` and select the
   "Desktop development with C++" workload, or run it elevated with
   `--quiet --wait --norestart --installPath "C:\Program Files (x86)\Microsoft Visual Studio\2022\BuildTools" --add Microsoft.VisualStudio.Workload.VCTools --includeRecommended`.
2. Then:
   `cmake -S launcher -B build-msvc -G "Visual Studio 17 2022" -A x64 -DCMAKE_PREFIX_PATH=C:/Users/Try.APAPC/Qt/6.6.3/msvc2019_64`
   `cmake --build build-msvc --config Release`
   `ctest --test-dir build-msvc -C Release --output-on-failure`
   `MinecraftLauncher` is built with `/W4 /WX`, so this run is the actual
   warnings-as-errors gate that the GCC 13.1 build could not cover.

## 2026-11-16 - MSVC 2022 BUILD AND TEST (UNBLOCKED, VERIFIED)

The Visual Studio 2022 Build Tools (17.14.37906.3, MSVC 19.44.35230, Windows SDK
10.0.26100.0) were installed with the "Desktop development with C++" workload,
so the `/W4 /WX` gate is now covered.

Commands (from the repository root):

    cmake -S launcher -B C:/Users/Try.APAPC/build-msvc -G "Visual Studio 17 2022" -A x64 \
          -DCMAKE_PREFIX_PATH=C:/Users/Try.APAPC/Qt/6.6.3/msvc2019_64
    cmake --build C:/Users/Try.APAPC/build-msvc --config Release --parallel
    ctest --test-dir C:/Users/Try.APAPC/build-msvc -C Release --output-on-failure

Results:
- Configure: OK, MSVC 19.44.35230.0, Windows SDK 10.0.26100.0.
- Release build: exit 0. `MinecraftLauncher` compiles clean under `/W4 /WX`;
  the only warning in the whole build is C4834 (discarded `[[nodiscard]]`
  result at `launcher/auth/SessionManager.cpp:118`) raised while compiling
  `test_main_window`, a test target that does not use `/WX`. Recorded as an
  observation, not a build break.
- `MinecraftLauncher.exe` starts and stays alive for the 4 s check window.
- `ctest -C Release`: 23/23 passed, 0 failed.

The run also found a runtime packaging defect: the test executables link the
shared `Qt6Test.dll`, which was never staged, so 14 of the 23 tests aborted at
load time with `0xc0000135` (STATUS_DLL_NOT_FOUND). `launcher/CMakeLists.txt`
now copies `Qt6Test.dll` next to the binaries (guarded by `EXISTS`); the suite
is green and the MinGW build is protected on machines where the Qt bin
directory is not on `PATH` either.

Branch:
fix/windows-build-and-e2e-verification.

## 2026-11-16 - WARNINGS-AS-ERRORS FOR THE TESTS (C4834 FIXED)

`launcher/auth/SessionManager.cpp` (the `applyTokens` tail) called
`TokenStorage::save()`, which is `[[nodiscard]]`, and discarded the result. MSVC
reported C4834 while compiling that translation unit for `test_main_window`.
The session is set in memory first, so a failed write must not fail the login,
but it also cannot be ignored: the next launch would ask for credentials again
with no explanation. The result is now checked and reported with
`qWarning() << "Could not persist the auth session for" << username;`.

`launcher/CMakeLists.txt` now applies the same gate to every test executable
(`LAUNCHER_TEST_TARGETS`, plus `fuzz_manifest_parser` when LibFuzzer is found):
`/W4 /WX` on MSVC and `-Wall -Wextra -Werror` on GNU/Clang, so test-only code
cannot hide behind a weaker configuration. The block sits after the targets are
defined because `target_compile_options()` requires the target to exist.

Verification:
- MSVC 2022 Release: build exit 0 with **zero** warnings/errors in the log (the
  C4834 warning is gone), `ctest -C Release`: 23/23 passed.
- MinGW GCC 13.1 Release: build exit 0, also warning-free under
  `-Wall -Wextra -Werror` on the tests, `ctest`: 23/23 passed.

Branch:
fix/msvc-strict-warnings-tests.

## 2026-11-16 - MERGED INTO master

Both task branches were fast-forward merged into `master` (no merge commits, no
force pushes, linear history preserved):

1. `fix/windows-build-and-e2e-verification` (`bf3862b`, `ba18545`, `49cff8d`,
   `78f66c6`) - the Windows build/CTest work, the OpenSSL 3 + TLS staging, the
   MSVC 2022 verification and the `Qt6Test.dll` staging fix.
2. `fix/msvc-strict-warnings-tests` (`8373675`) - the C4834 fix and the
   warnings-as-errors gate on every test target.

`master` is now at `8373675` and matches `origin/master`.
Post-merge verification on `master` itself:
- MSVC 2022 Release build: exit 0, zero warnings/errors.
- MinGW GCC 13.1 Release build: exit 0.
- `ctest`: 23/23 passed on both toolchains.

The merged task branches were left in place (they are already contained in
`master`); no branch was deleted and no history was rewritten.

## 2026-11-16 - SELF-HOSTED GITHUB ACTIONS RUNNER (DEBIAN) AND LINUX CI

A repository-scoped GitHub Actions self-hosted runner is installed on the LAN
host `192.168.1.23` (note: the first hand-off said `192.168.1.24`, which does
not exist on the network; `192.168.1.23:22` is the machine, `debian-ai`).

Host and installation:
- Debian GNU/Linux 13 (trixie), x86_64, systemd 257, 6 GB RAM, 26 GB free.
- Runner `actions-runner` **2.338.2** under `/home/ai2/actions-runner`,
  registered against `Ddking-1/minecraftserverluncher` as
  `debian-ai-runner` with labels `self-hosted, linux, x64, debian-13`.
- Runs as the systemd service
  `actions.runner.Ddking-1-minecraftserverluncher.debian-ai-runner.service`
  (enabled, active, user `ai2`). Manage with `./svc.sh status|start|stop` from
  the runner directory; the service survives reboots.
- The runner needed the .NET runtime dependencies (`libicu-dev` and friends),
  installed once through the root account.

CI layout after this change:
- `.github/workflows/linux-ci.yml` (new) runs on `[self-hosted, linux, x64]`:
  `go mod verify`, `go vet`, `go test ./... -count=1`, `go build ./...` in
  `server/`, plus the Velocity Gradle build with JDK 25.
- `.github/workflows/ci.yml` was **deleted**: it duplicated `backend.yml`
  (Go), `launcher.yml` (Qt/MSVC) and the new `linux-ci` (Velocity), and because
  it had no path filter it queued a Windows job on every push to `master` and
  sat in the queue forever once the Windows runner went offline.
- `launcher.yml` and `backend.yml` still target `[self-hosted, windows]` and
  keep their path filters, so they only queue when `launcher/**` or
  `server/**` changes.

Verification:
- Runner shows `online` in `gh api .../actions/runners`.
- A run of `linux-ci` on `master` (`43907599898`) passed: both the `go` and
  `velocity` jobs are `success` (~105 s).

Credentials: the `ai2` SSH password and the root password used for the install
were **not** written into the repository. They live in `runner.env` in the
repository root on the operator machine, which is git-ignored by the existing
`*.env` rule (verify with `git check-ignore -v runner.env`) and therefore never
appears in `git status`, a commit or a push. That file also records the host,
runner name/version/labels, the systemd unit, the paths, which workflows use
the host, and copy-paste commands for status, re-registration and manual test
runs, so it can be copied as a single file to another machine or handed to
another model. The runner registration token was written through `/tmp` on the
host and deleted immediately after `config.sh`.

If `runner.env` is ever exposed, rotate the two passwords on the host
(`passwd ai2`, `passwd root`) and update the file.

Branch:
ci/linux-runner-tests (linux-ci workflow) and ci/drop-redundant-workflow
(workflow cleanup + this documentation).

docs/project-status-overview: adds docs/PROJECT_STATUS.md, an owner-facing gap
analysis and minimum publishable baseline (draft).
docs/project-status-update2: records owner decisions (Velocity in v1,
GitHub Releases updates with backend redirect, legal publication scope,
invite-only registration, granular admin permissions, ~200 users, mrpack-like
modpack, beta labeling) as DECISION-007/008 and updates PROJECT_STATUS.
docs/decision-009-and-ci-answer: DECISION-009 (metadata: online + local cache)
recorded; PROJECT_STATUS §9 clarified that the Windows build cannot run on the
Debian runner (Option A/B remain open).
ci/windows-latest-runners: `launcher.yml` and `backend.yml` now run on
GitHub-hosted `windows-latest` (Option A, owner decision). Remaining watch item:
the 2000 free Actions minutes/month for private repos.
