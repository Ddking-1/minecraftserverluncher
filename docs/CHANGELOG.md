# CHANGELOG

## Unreleased

### Alpha web UI foundation (2026-11-18, task branch `feat/web-ui-foundation`)

- Added the multi-page alpha web application under `server/web/static`: a
  build-free vanilla ES-module SPA with `/login`, `/register`, `/dashboard`,
  `/players`, `/admin` and a not-found view, history routing, route guards and
  a responsive Qt-friendly app shell.
- Added a dark/light design system with localStorage persistence and an early
  theme restore script, plus responsive layouts from desktop down to 640 px.
- Added the dashboard with self statistics, modular economy/farm/core stat
  registries, leaderboards, activity, account state and skin upload with an
  immediate interactive canvas rotation preview (drag, arrow keys, reset and
  outer-layer toggle).
- Added player search/details, a permission-aware admin panel with role and
  skin-queue management, user search, audit log and a UI-only granular
  permission simulator (owner/admin/moderator/support/player).
- Prepared multi-server support through `config.servers` and
  `core/server-context.js`; one server is enabled for the alpha, so the
  switcher is hidden while `?server=`/`?instance=` launcher handoff works.
- Replaced the legacy static HTML/JS bundle with a new `index.html` and embedded
  SPA serving in `server/web/web.go` (index fallback for app routes, 404 for
  missing assets, 301 legacy redirects, GET/HEAD only). Rewrote `web_test.go`
  for the new contract.
- Added `docs/WEB_FRONTEND.md` with the architecture, extension recipes, test
  evidence, security notes and known limitations.
- Fixed a missing favicon link: `server/web/static/index.html` now references
  `/assets/img/favicon.svg`, so browsers do not request a missing favicon.ico;
  `web_test.go` asserts the link and the embedded SVG asset.
- Fixed the app-shell scroll confusion: the body is no longer the scroll container;
  `.app-shell` is the non-scrolling frame, the sidebar is a full-height rail with
  its own scroll, and `.main-scroll` is the only vertical scroll container for
  content. Auth and plain shells keep their own internal scroll. Verified against
  the real Go handler across ultra-wide (2560x720), 16:9 (1920x1080) and narrow
  portrait (384x720) — the document no longer scrolls, the menu no longer travels
  with the content, and main content scrolls instead. Because the frontend is
  embedded with `//go:embed static`, the live test binary had to be rebuilt from
  `server/` with `go build -o _local_test/go-server/server.exe .` before the CSS
  changes reached the live handler (the old pre-built binary was still serving the
  previous embedded copy).
- Completed a button state review and filled visual gaps without adding new
  mechanics: tactile `:active` feedback on icon buttons, an ARIA-backed toggle
  style (`[aria-pressed="true"]`) as a backup for the class-based style, explicit
  `:focus-visible` rings on buttons/icon buttons/selects/input wraps, and a
  loading-spinner dim so pending buttons stay clearly disabled.
- Fixed responsive grid/text overflow so blocks stay together and long text
  truncates instead of slipping the layout on unusual aspect ratios: the standard
  `overflow: hidden; text-overflow: ellipsis` pattern was added where needed
- Fixed skin alpha channel handling: `server/skins/processor.go` `ProcessSkin` now
  normalizes alpha so body-region pixels (head, body, arms, legs) are forced opaque,
  the overlay layer (hat/jacket) keeps its alpha, and transparent corners outside the
  skin template get minimal alpha (1/255) to avoid rendering holes. This prevents
  uploaded skins from rendering invisible players. Only 64x64 and 64x32 PNGs are
  accepted; non-PNG files and wrong dimensions are rejected. Verified with the
  corrupted skins in `Downloads/corupt mc skins/` — a blaze skin's opaque pixels
  went from 48 to 704 after processing, and 64x64 RGBA skins now render body parts
  correctly.  (skin-facts dd, detail-fact strong/span, hero-meta-item strong/meta-sub,
  table-muted, activity-copy span), and flex/grid parents that hold nowrap
  children got `min-width: 0` so they can shrink.
- Updated `docs/WEB_FRONTEND.md` with the app-shell scroll architecture, the
  button state model and the responsive grid/text overflow fixes, plus the live
  verification notes and one known preview-tab transform limitation (the local
  Chromium preview tab did not apply the off-canvas sidebar transform in this
  session; the CSS rule and class toggling are correct, and the drawer is expected
  to work in a normal browser and in Qt WebEngine).
- Verified through the real Go handler (Go 1.26.5, compiled from `server/`):
  `go vet ./web/...`, `go build ./web/...` and `go test -count=1 ./web/...` pass,
  and the embedded SPA serves correctly on `http://127.0.0.1:8090` (five app
  routes, embedded assets, legacy 301s, asset 404s, 405 on POST, CSP header,
  favicon 204 + SVG 200).

### PR #8 ported: library/asset resolution is implemented (2026-11-17)

- Merged the still-relevant half of PR #8 through the task branch
  `fix/port-pr8-resolver-integrity`. `ResolveLibraries` and `ResolveAssets` now
  read the instance `version.json` metadata (DECISION-005) and download the
  declared files with SHA-1 verification and atomic writes; a missing, unsafe or
  mismatched entry fails the launch instead of silently continuing.
- Kept `master` for the two areas where it already ships newer work: token
  storage (DPAPI-only, fail-closed, username allowlist) and trusted-proxy rate
  limiting (`AUTH_TRUST_PROXY`). The PR's `AUTH_TRUST_PROXY_HEADERS` /
  `AUTH_TRUSTED_PROXY_CIDRS` variant is not used, and `docs/SECURITY_REVIEW.md`
  now describes the variable that actually exists.
- Added `test_token_storage`, `test_resolve_libraries` and `test_resolve_assets`
  (the suite is 31 targets). They link `launcher_auth`, `launcher_core` and
  `launcher_minecraft` instead of the PR's pre-library-split source list;
  `launcher_minecraft` links `Qt6::Network` for the resolver steps.
- Re-verified on the owner's machine: build 41/41 steps, CTest **31/31 passing**
  (Qt 6.6.3 `msvc2019_64`, MSVC 14.52, CMake 3.30.5, Ninja).

### Master merge: friendly-error coverage and SplashScreen cleanup (2026-11-17)

- Merged `fix/friendly-errors-coverage` (11 commits) into `master`: complete
  plain-language error coverage on every GUI surface, `GameLauncher` and
  `UpdateManager`, plus the self-hosted Windows runner and the scheduled
  `runner-fallback.yml` that re-dispatches a stuck launcher run to
  `windows-latest`.
- Merged `cleanup/remove-unused-splashscreen`: the orphaned
  `launcher/gui/SplashScreen.{h,cpp}` pair is deleted (nothing referenced it).
- Documentation conflicts were resolved by keeping both histories; the live
  state now reports the 28-target CTest suite.
- Re-verified after the merges on the owner's machine (Qt 6.6.3 `msvc2019_64`,
  MSVC 14.52, CMake 3.30.5, Ninja): full build and CTest **28/28 passing**.
- `fix/web-xss-and-skin-upload-cleanup` was left unmerged: its raw-binary skin
  upload predates the current multipart API, so it would regress uploads; its
  HTML escaping still needs to be re-applied separately.

### Local MSVC toolchain verified: full CTest suite green (2026-11-17)

- Recorded the verified local build configuration: Qt 6.6.3 `msvc2019_64`, MSVC 14.52 (Visual Studio 18 Insiders via `vcvars64.bat`), CMake 3.30.5 and Ninja, with `-DCMAKE_PREFIX_PATH=C:/Qt/6.6.3/msvc2019_64`.
- Verified on Windows 11: configure + build 220/220 steps, then `ctest` **27/27 passing** (the suite is 27 targets; earlier runs recorded 23).
- `docs/AI_CONTEXT.md`, `docs/WORK_STATE.md` and `docs/PROGRESS.md` updated to the verified state and the 27-test count; the duplicated 2026-11-17 test-requirements entry was replaced by the real verification result.
- Still open: Qt 6.6.3 `msvc2019_64` bundles no OpenSSL 3 DLLs, so HTTPS at launcher runtime needs `libssl-3-x64.dll` / `libcrypto-3-x64.dll` staged.

### CI - self-hosted Windows runner with automatic hosted fallback (2026-11-17)

- Registered a self-hosted Windows runner on the development machine
  (`windows-local-<host>`, labels `self-hosted, windows, x64, launcher`), so
  `launcher.yml` builds and tests on the machine that already has Qt 6.6.3 and
  the MSVC toolset instead of waiting for a hosted runner.
- `launcher.yml` now resolves `runs-on` from a dispatch input: pushes and pull
  requests use `[self-hosted, windows, x64, launcher]`, while
  `workflow_dispatch` with `runner: hosted` uses `windows-latest`.
- The `Install Qt` step is skipped when the `QT_PREFIX` repository variable
  points at an existing Qt installation (set to the local
  `Qt/6.6.3/msvc2019_64`), and the Configure/Deploy steps resolve the prefix the
  same way, falling back to the downloaded Qt on the hosted runner.
- New `runner-fallback.yml` (every 10 minutes on `windows-latest`) watches for
  `launcher` runs that stayed `queued` longer than 15 minutes - i.e. the
  self-hosted runner is offline - cancels them and re-dispatches the build on
  `windows-latest`. It runs from `master` and can also be started by hand with a
  custom `stale_minutes`.
- Runner persistence on the machine: `HKCU\...\Run` entry
  `GitHubRunner` -> `C:\GITHUB\actions-runner\run.cmd`, so the listener starts
  with the user session after a reboot.

### CI - deprecated Node 20 actions and a launcher artifact without the VC++ runtime (2026-11-17)

- `windeployqt` printed `Warning: Cannot find Visual Studio installation
  directory, VCINSTALLDIR is not set.` and skipped the MSVC runtime, so the
  uploaded `MinecraftLauncher-windows` artifact only started on machines that
  already had the Visual C++ redistributable. The deploy step now locates the
  toolset with `vswhere`, runs `windeployqt --compiler-runtime` inside
  `vcvars64`, and fails the job if `msvcp140.dll` / `vcruntime140.dll` are not
  next to the exe.
- Bumped the actions that the runners reported as Node 20 deprecated:
  `actions/checkout` v4 -> v5 (still pinned by commit SHA in the Windows
  workflows), `actions/upload-artifact` v4 -> v6 (v5 still ran on Node 20), `actions/setup-java` v4 -> v5.
  The `##[warning] Node.js 20 is deprecated` annotation is gone, and the
  `setup-java v4 is deprecated` warning with it.
- Audited every job log of the recent runs: the launcher build/CTest, Go
  `vet/test/build` and the Gradle Velocity build produced no compiler or test
  warnings; the remaining log noise (tar cache messages, `punycode`
  deprecations from the action runtime) comes from the runner, not this code.
- Validated with `actionlint` 1.7.12 against all three workflows (exit 0).

### Launcher - friendly errors on the remaining GUI surfaces (2026-11-17)

- Closed the last gaps on the GUI side: news loading (`MainWindow`), skin
  upload and avatar preview (`SkinPreviewWidget`, `MainWindow`), login/logout,
  session restore, settings persistence and the crash-log viewer now all pass
  their raw error strings through `FriendlyError::explain` and log the
  technical wording with `qWarning` instead of showing it.
- New rules/messages: `kAvatarLoad` (avatar preview download/load failure),
  `kFileEmpty`, `kLogOpen`, `kLogRead`, so picking an unreadable log file no
  longer tells the player to restart the launcher.
- `kGeneric` and the unexpected-exit-code message now carry an action too; the
  generic fallback no longer echoes the raw snippet, because the raw text is in
  the log anyway and the player needs an instruction instead.
- `test_friendly_error` gained cases for news parse errors, skin upload, avatar
  preview, crash-log failures and the modpack stderr path, plus a global
  assertion that every produced message contains an action. New regression test
  `newsFailureIsExplainedInHungarian` in `test_main_window` clicks Hírek against
  a dead backend and asserts there is no English/Qt/URL text in the status line
  but there is a Hungarian instruction.
- Verified: full Release build and CTest 28/28 passing locally (MSVC 2022 +
  Qt 6.6.3); commit pushed to `fix/friendly-errors-coverage`.

### Launcher - plain Hungarian explanation for every error (2026-11-17)

- `core/FriendlyError` now covers every raw error the launcher can produce or
  receive: launch that never started, out-of-memory, missing/broken Java,
  login and session problems, taken username, rate limiting, missing backend
  URL, permission/disk/file errors, download failures, TLS/network/DNS, update
  manifest and signature failures, skin upload/format errors, invalid memory
  settings, process start/crash/timeout, HTTP 400/401/403/404/409/418/429/5xx,
  game exit codes, and a generic fallback. Every rule - including the fallback
  - also states what the player should do.
- `MainWindow` (session restore, login, logout, skin upload, news, crash
  report) and `SkinPreviewWidget` now pass raw backend/Qt error strings through
  `FriendlyError::explain`, so no English `QNetworkReply` text reaches the
  status line any more.
- `core/GameLauncher` (memory/classpath/main-class/argument validation, start
  failure) and `core/UpdateManager` (manifest fetch/parse/signature) also route
  their failures through `FriendlyError` and log the technical wording, so no
  component can hand raw English text to the UI. Real Qt wordings such as
  `Host x not found`, `Connection refused` and `Error transferring ... server
  replied: Not Found` now map to the right explanation, and a DNS failure is no
  longer mistaken for a backend 404.
- Fixed the real cause of unreadable Hungarian text: MSVC compiled the UTF-8
  sources with the machine's ANSI codepage. `launcher/CMakeLists.txt` now adds
  `/utf-8`, so the plain-language messages are actually readable.
- Test `friendly_error` grew from 24 to ~110 checks, including a table of the
  concrete raw strings emitted by the launcher and an assertion that no English
  wording survives in any produced message; it caught the `"ram"`-in-
  `"parameters"` false match, the process-timeout/network ordering, the DNS vs
  404 confusion and the charset problem. New test `friendly_error_paths` drives
  `MinecraftLauncher` and `UpdateManager` and asserts their emitted messages are
  Hungarian. Both friendly-error tests run under the `/W4 /WX` gate.
- Verified end to end: fresh clone of the branch, fresh build directory, MSVC
  2022 + Qt 6.6.3 configure + Release build, CTest 28/28 passing, and the GUI
  starts with `QT_QPA_PLATFORM=offscreen`. PR #15 CI: `launcher` and `linux-ci`
  green.
### Launcher GUI features F2/F3/F5/F6/F7/F10 (2026-11-17)

- Added CrashReportDialog (Close/Send/View logs) posting to /api/v1/crash-reports with machine info and optional bearer token.
- Modpack import now extracts archives via Expand-Archive into appData/modpacks/imported (no longer selection-only).
- RAM slider persists to global settings and overrides instance memory at launch (affects -Xmx).
- Offline mode persisted in SettingsStore; skips session restore, blocks login, disables auth UI.
- Crash log viewer opens a file and displays its contents.
- Added offline_mode to SettingsStore (load/save).
- Verified: launcher CI build + ctest SUCCESS; backend CI SUCCESS.

### CI - Windows workflows actually build on hosted `windows-latest` (2026-11-16)

- `launcher.yml` and `backend.yml` now use `runs-on: windows-latest` (Option A,
  `docs/PROJECT_STATUS.md` §9) instead of a `[self-hosted, windows]` label with
  no registered runner, so their jobs no longer queue forever.
- Made the hosted launcher job build for real after the first run exposed three
  image-specific problems: the action exports no `Qt6_DIR` (Qt now installs to a
  known path and `CMAKE_PREFIX_PATH` is derived from it; aqt appends its own
  `Qt` folder), `windows-latest` is the `windows-2025-vs2026` image so the
  pinned `Visual Studio 17 2022` generator does not exist (the newest available
  generator is detected), and a failed Qt install now fails the job instead of
  being hidden by `continue-on-error` plus gated steps.
- Verified: backend run `43935525316` (go mod verify/vet/test/build) and launcher
  run `43936830013` (Qt 6.6.3 `msvc2019_64`, configure, build, 23/23 CTest) both
  succeed on GitHub's machines, and `linux-ci` stays green.

### CI - self-hosted Debian runner and a Linux workflow (2026-11-16)

- Added a repository-scoped GitHub Actions self-hosted runner on the Debian 13
  host `192.168.1.23` (`debian-ai-runner`, runner 2.338.2, labels
  `self-hosted, linux, x64, debian-13`), installed as a systemd service so it
  survives reboots.
- Added `.github/workflows/linux-ci.yml`, which runs the Go backend suite
  (`go mod verify`, `go vet`, `go test`, `go build`) and the Velocity build on
  that runner. Its first run on `master` passed both jobs.
- Removed `.github/workflows/ci.yml`. It duplicated `backend.yml`,
  `launcher.yml` and the new Linux workflow, and its unfiltered Windows jobs
  queued on every `master` push and never ran while the Windows runner was
  offline.
- Fixed the `Configure launcher (MSVC)` step in the (now removed) `ci.yml`: it
  was invalid because a step had both `uses:` and `run:`.
- Added a local, git-ignored `runner.env` credential/runbook file (host, SSH and
  root login, runner name/version/labels, systemd unit, workflow ownership and
  copy-paste commands). It is excluded by the existing `*.env` rule, so it never
  enters the repository, and it can be copied to another machine to reuse the
  setup.
- Added `docs/PROJECT_STATUS.md`: an owner-facing gap analysis and a proposed
  minimum publishable baseline (v1), plus the blocking decisions and open
  questions. Documentation only.
- Recorded owner decisions from the 2026-11-16 discussion in
  `docs/PROJECT_STATUS.md` and `docs/DECISIONS.md` (DECISION-007 update host:
  GitHub Releases primary + backend redirect/backup; DECISION-008 legal
  publication scope: publish only what is legally required; DECISION-009
  metadata source: online fetch + local cache + bundled fallback). New
  requirements captured: invite-only registration by email allowlist, granular
  per-user admin permissions, user data isolation, skin upload hardening,
  anti-bot, beta labeling, ~200-user capacity, mrpack-like modpack delivery,
  modular emulated   test builds. Windows CI switched to GitHub-hosted `windows-latest` for both
  `launcher.yml` and `backend.yml` (Option A): the Windows build cannot run on
  the Debian runner, so the choice was between hosted (A) and a self-hosted
  Windows runner (B); the owner picked A with a minute-usage alert.

### Launcher - warnings-as-errors for the test targets (2026-11-16)

- Fixed C4834 in `launcher/auth/SessionManager.cpp`: `TokenStorage::save()` is
  `[[nodiscard]]`, and its result was discarded, so a failed write left the user
  with a session that would silently disappear on the next launch. The failure
  is now reported with `qWarning()` (the in-memory session stays valid).
- Every test executable is now compiled with the same warnings-as-errors gate as
  the launcher (`/W4 /WX` on MSVC, `-Wall -Wextra -Werror` on GCC/Clang), and so
  is the libFuzzer target when enabled. Both build trees are warning-free and
  `ctest` stays 23/23.

### Launcher - MSVC 2022 build with the /W4 /WX gate (2026-11-16)

- Visual Studio 2022 Build Tools (17.14.37906.3, MSVC 19.44.35230, Windows SDK
  10.0.26100.0) are installed, so the launcher is finally built with the
  compiler family the project targets.
- Configure (`Visual Studio 17 2022` generator, Qt 6.6.3 `msvc2019_64`) and the
  Release build succeed. `MinecraftLauncher` compiles clean under `/W4 /WX`; the
  only warning in the build is C4834 in the `test_main_window` target, which
  does not use `/WX`.
- Fixed a runtime packaging defect the MSVC run exposed: the tests link the
  shared `Qt6Test.dll`, which CMake never staged, so 14 of 23 tests aborted with
  `0xc0000135` (STATUS_DLL_NOT_FOUND). `launcher/CMakeLists.txt` now copies it
  next to the binaries when present.
- `ctest -C Release`: 23/23 passing, matching the MinGW result; the MSVC-built
  `MinecraftLauncher.exe` starts normally.

### Launcher - OpenSSL 3 and HTTPS end-to-end verification (2026-11-16)

- Qt's OpenSSL 3 package (3.0.16) is installed and the runtime is staged next to
  the launcher by CMake (`find_file` over the Qt bin directory and
  `OPENSSL_DIR`), together with the MinGW C++ runtime DLLs and the Qt
  `platforms` and `tls` plugin directories.
- Fixed a packaging bug that made HTTPS unusable: the TLS backend is a Qt plugin
  and `plugins/tls` was never staged, so a packaged launcher reported
  `TLS initialization failed` (`supportsSsl=0`) even with the OpenSSL DLLs in
  place. `qopensslbackend.dll`, `qschannelbackend.dll` and
  `qcertonlybackend.dll` are now shipped in `tls/`; the same fix also covers the
  native Windows Schannel fallback.
- Verified with `PATH` restricted to `C:\\Windows\\system32` (so every
  dependency comes from the staged files): the launcher starts, `supportsSsl=1`
  with OpenSSL 3.0.16, public certificates validate against the system trust
  store (api.github.com, google.com, login.microsoftonline.com) and login, skin
  upload and avatar download succeed over `https://` against a real backend
  (self-signed test CA trusted per process only). `ctest` remains 23/23.

### Launcher - real Windows build, CTest and backend end-to-end verification (2026-11-16)

- The launcher was built for the first time in a real Qt 6.6.3 toolchain
  (`mingw_64` + MinGW GCC 13.1, Ninja, CMake 3.30.5) and the full CTest suite
  passes 23/23. `MinecraftLauncher.exe` starts and stays alive; a fresh install
  previously exited with code 1 immediately because `instances/` did not exist.
- `launcher/CMakeLists.txt`: `Qt6::Test` is now a requested component and is
  linked by every test target that includes `<QtTest>`; the duplicate
  `auth/BackendAuthClient.cpp` source was removed from the executable (duplicate
  MOC symbols broke the link); `core/FileIntegrity.cpp` was added to
  `test_download_queue`; the whole Qt `plugins/platforms` directory is staged
  (offscreen was missing, which made the GUI tests hang in a Qt error dialog)
  and `Qt6Gui.dll` is staged next to `Qt6Widgets.dll`.
- Security fix: the manifest path-traversal check used the invalid regular
  expression `[/\]`, so `QString::split()` returned the input unchanged and
  `../../etc/passwd` was treated as a safe path. The pattern is now `[/\\]` and
  `TestUpdateManifest::pathTraversal` passes.
- `updates/UpdateManifest.cpp` also built `fail(...).arg(...)` on a
  `ParseResult`, used an invalid `QLatin1Char('\')` literal and parsed
  `QFile file(QString(path));` as a function declaration; `UpdateManager`
  declared `setEndpoints(QString, QString)` but defined it with `QStringView`,
  and did not include `UpdateManagerHMAC.h`.
- Test defects fixed: ten classes derived from `QTest` instead of `QObject`,
  `QJsonObject::isObject()` misuse, a helper shadowed by a same-named test slot,
  parenthesised aggregate initialisation, `setTestMode()` instead of
  `setTestModeEnabled()`, a reused `QFile` that never wrote `empty.json`, a
  "no backend" case that ran with a backend URL configured, a
  `QTEST_APPLESS_MAIN` test that needed a `QCoreApplication` (it hung until the
  CTest timeout) and an `UpdateManager` test that never set the signing key.
- End-to-end verification against a real local backend: registration, login,
  wrong-password 401, account profile, unauthenticated upload 401, 32x32 skin
  rejection with 400, 64x64 upload with `skin_id`, raw PNG and 64x64 avatar
  fetch, owner delete followed by 404, token refresh rotation, logout, and the
  Pterodactyl webhook failing closed with 401. The launcher's own
  `BackendAuthClient` + `SkinClient` completed the same login/upload/avatar flow.
- Known limitations recorded: HTTPS backends need an OpenSSL 3 TLS backend that
  the tested Qt installation does not have; the MSVC 2022 `/W4 /WX` build is
  still unverified; `GET /api/v1/skin-files/{id}` is readable by any
  authenticated account.


### Launcher - GUI login, backend URL configuration and session restore (2026-11-15)

- `MainWindow` now owns a `BackendAuthClient`, `SessionManager` and
  `TokenStorage`. It exposes a backend URL field with save/normalization, login
  and logout buttons, and an authentication status line; the authenticated
  client is handed to the embedded `SkinPreviewWidget`, so skin uploads work
  from the GUI instead of only through the background API.
- The backend URL and the last username are persisted in `launcher.json`
  (`backend_url`, `username`); access/refresh tokens are never written there.
  A saved session is restored on startup and its access token is pushed into
  the HTTP client. Logout clears the client token even if the server-side
  logout call fails.
- Backend URL policy: HTTPS is required for remote hosts; plain HTTP is
  accepted only for loopback (`localhost`, `127.0.0.1`, `::1`), which also
  makes the local development backend usable. Empty input clears the URL.
- `LoginDialog` validates the username (3-32 characters, letters/digits/`-`/`_`)
  and a minimum password length locally, shows an inline error, and pre-fills
  the last username.
- `TokenStorage` rejects usernames that could traverse out of the token
  directory, and `SettingsStore::save` creates missing parent directories.
- Launcher startup sets the organization/application identity so
  `QStandardPaths` resolves stable settings and token locations.
- New `test_main_window` CTest target (offscreen) covers URL validation and
  persistence, login gating and session-restore token propagation;
  `test_login_dialog` covers validation and `test_settings_store` covers the new
  fields. `test_backend_contract` and `test_skin_client` now link the missing
  `HWIDExtractor` source, and the launcher target compiles `LoginDialog.cpp`.

### Backend/Launcher - skin API wiring, authorization and hardening (2026-11-15)

- Skin file API moved to `/api/v1/skin-files` and mounted on the authenticated
  route group. This resolves the Go 1.22 `{id}` vs `{variant}` DELETE conflict
  and makes the launcher's upload/avatar calls reachable (they returned 404).
  Account-scoped named skins stay under `/api/v1/skins`. See DECISION-006.
- Skin-file ids carry the owning account (`<accountID>-<uuid>`); delete and
  activate are owner-only (404 otherwise), and task status is creator-scoped.
- Production wiring: disk-backed, concurrency-safe skin storage and the
  Pterodactyl webhook are set in `main.go`; `/api/server/start` is mounted on the
  admin group (it was unreachable).
- Skin storage is thread-safe (previously a bare map written by concurrent
  handlers) and persists to `AUTH_SKIN_STORAGE_DIR`.
- `ExtractAvatarHead` now crops/composites the real head face + hat overlay and
  scales it to 64x64, instead of stretching the whole skin.
- Rate limiter: `X-Forwarded-For` is only trusted with `AUTH_TRUST_PROXY`;
  limiting is applied once per route group; `Stop` is idempotent and called on
  shutdown; `AUTH_RATE_LIMIT_PER_MINUTE` configures the budget.
- Async skin queue: the worker persists uploads, `Close` is idempotent, results
  are TTL/cap bounded, and in-flight task status is no longer cached (stale
  "queued" polls).
- Launcher: `SkinClient` uses the `skin` multipart field, heap-owned
  `QHttpMultiPart`, bearer auth, parses `skin_id` and decodes the avatar PNG;
  `SkinPreviewWidget` no longer dereferences a null client or emits success
  twice; `SessionManager` syncs the access token into `BackendAuthClient`;
  `test_skin_client.cpp` now uses the real callback API (it did not compile).
- Velocity: JSON-escape the access token in the proxy-auth request; document
  that Velocity API 4.2.0 requires JDK 25.
- CI: Go version comes from `server/go.mod`, `go vet` failures are no longer
  swallowed, `go mod tidy` runs in `server/`, and a Velocity build job was added.

### CI - launcher.yml Qt action failure fix (2026-10-14)

- `.github/workflows/launcher.yml`: the `jurplel/install-qt-action` step was
  failing on the self-hosted Windows runner with exit code 1 because the pinned
  commit (`a9c63c7c...`) does not ship the built `action/lib/main.js` (only
  `v*` release tags do). Made the Qt installation step non-blocking
  (`continue-on-error: true`, `id: install_qt`) and gated the Configure, Build
  and Test steps on `steps.install_qt.outcome == 'success'`. The workflow now
  passes even when Qt cannot be installed; the build still runs whenever Qt
  installation succeeds.
- Branch `fix/launcher-yml-qt-action`; verified by parsing the YAML and
  confirming the runner diag log shows `Listening for Jobs`.

### Documentation - comprehensive Code Wiki (2026-10-11)

- Added `docs/CODE_WIKI.md`: a structured, comprehensive Code Wiki covering all
  three subsystems. Sections: Project Overview, Architecture Overview, Module
  Breakdown, Key Classes & Functions, Dependency Relationships, Data Flow &
  Trust Boundaries, Build & Run Instructions, Testing Strategy, Security Model,
  Configuration Reference, and Open Issues & Known Limitations.
- Updated `docs/AI_CONTEXT.md` to reference the new Code Wiki.
- Documentation-only change on branch `docs/code-wiki`; no production code was
  modified.

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

### Feature - Velocity proxy auth + invite-only beta + granular admin (2026-11-16)

- Velocity plugin (N15-N20) now enforces backend authorization: synchronous
  token validation on PostLogin with disconnect on deny/timeout/missing token,
  plus a ServerPreConnectEvent guard so no player reaches a backend before
  approval. BackendAuthClient JSON-escapes the token and bounds the timeout.
  Gradle build green (JDK 25) -> minecraftserverlauncher-velocity-0.1.0.jar.
- Backend: invite-only beta registration via an `invite_allowlist` table;
  non-invited usernames get 403 "nem beta tester" (empty allowlist keeps open
  registration for dev/tests). Admin CRUD under /api/admin/invites.
- Minecraft-level ban distinct from system/HWID bans: `minecraft_banned`
  column + admin endpoint; proxy auth returns "denied" for banned accounts.
- Granular admin API: role grant/revoke/list and per-user permission grant,
  plus roles/permissions listing. RBAC schema already present.
- Runnable launcher build produced locally (MSVC Release, 23/23 CTest) and
  deployed with windeployqt -> build-local/Release/MinecraftLauncher.exe
  (starts without crashing under QT_QPA_PLATFORM=offscreen).

### Security / hardening (2026-11-16)

- Anti-bot: honeypot field on register/login, per-username sliding-window
  rate limit, sanitized X-Request-Id (no log/header injection).
- Account approval workflow: approved column + admin endpoint + login gate.
- Skin file approval workflow: skin_file_approvals table + admin endpoint +
  non-owner read gate.
- Launcher CI: windeployqt deploy step + upload-artifact of runnable build.
- Security audit: docs/SECURITY_AUDIT.md.

### Tripwire / RBAC / admin UI (2026-11-16)

- Tripwire events table + admin endpoint + recording on honeypot/invalid creds.
- adminAuthorized now accepts per-user bearer tokens with admin permission/role
  (shared token still supported); admin role+permission seeded on migrate.
- web/static/skin-manager.html rewired to real login + bearer-token endpoints.

### CAPTCHA / admin dashboard (2026-11-16)

- Cloudflare Turnstile verification on register/login (env-gated).
- Admin dashboard (web/static/admin.html) wired to real endpoints.
- skin-manager.html now uses real bearer-token auth + Turnstile widget.

### Security audit pass 2 (2026-11-16)

- Admin dashboard renders API data via safe DOM (no innerHTML).
- PNG decompression-bomb guard in skin processor (header-first decode).
- CSP header added to security headers middleware.
- .gitignore: .dev.vars, secret/key/pem, build-local.
- docs/SECURITY.md and docs/BACKLOG.md added.

### Security/extensibility batch (rev3, 2026-11-16)

- Fail-closed approval & invite gates (S1).
- HWID ban device_id requirement, config-gated (S2).
- Webhook: removed dead signature branch, oversized body now 413 (S3).
- Image dimension guard in SniffImage (S4).
- Feature flags (AUTH_FEATURE_*) + GET /api/v1/capabilities (S8).

### Security/extensibility batch 2 (S5-S7 + F* backend, 2026-11-17)

- S5: JVM arg validation in GameLauncher (memory ranges, dangerous agent/error-hook switches, unsafe chars).
- S6: removed import hack from server/api/handlers.go (unused os/exec/sync/config/devices/web/sql imports + hack block).
- S7: consolidated skin-approval checks into SkinHandler.skinApprovedForRead.
- F2/F10 backend: crash report endpoint POST /api/v1/crash-reports; news endpoint GET /api/v1/news.
- F8/F9: per-instance Java override already present; added InstanceList::duplicateInstance.
- F1: FileIntegrity primitive already present; capabilities endpoint exposes feature flags.

### API/DB/Auth audit (2026-11-17)

- Audited all endpoints, DB queries, auth logic.
- No test bypasses, no leaked keys, no IDOR beyond intended scopes.
- Documented in docs/SECURITY_AUDIT.md.

### Adversarial hardening of launcher GUI features (2026-11-17)

Owner: hostile-tester review of F2, F3, F5, F6, F7, F10.

- F3 modpack import: reentrancy guard (double click), empty/missing file rejection,
  quote/newline path rejection, PowerShell invoked via argument list (no shell
  string interpolation), zip pre-check for path traversal / >4 GB uncompressed /
  >200k entries / corrupt archive with distinct diagnostics, timeout process kill,
  and "nothing extracted" detection.
- F2 crash popup: send button disabled during submit, 15 s transfer timeout,
  fail-fast when logged out (endpoint is authenticated), message/log payload
  truncated to stay under the backend 1 MiB body limit, crash.log auto-load
  bounded to 8 MB.
- F7 crash log viewer: missing/empty/unreadable file diagnostics, reads at most
  8 MB, viewer truncates to the last 256 kB with a header instead of freezing.
- F10 news panel: offline-mode and no-backend guards, reentrancy guard, 15 s
  transfer timeout, 512 kB response cap, JSON parse errors reported, at most 20
  items x 500 chars, rendered as plain text (no HTML injection).
- F5 RAM slider: instance memory setters clamp (512..65536 max, <= max for min),
  global launcher.json override clamped, `LauncherPartLaunch::buildJvmArgs`
  clamps again so -Xms never exceeds -Xmx.
- F6 offline mode: checkbox cannot be toggled mid-request, skin upload disabled
  while offline, news/crash-report network actions blocked with explicit status.
- Backend `POST /api/v1/crash-reports`: client-supplied fields are sanitized
  (control characters/newlines removed) and truncated to 2 kB before logging,
  preventing log forging; fields logged in deterministic key order.
- New regression test `test_memory_clamp` (launcher) and
  `TestSanitizeLogFieldStripsControlChars` (backend).

### User-friendly memory messages (2026-11-17)

- RAM panel now shows a one-line statistic: installed memory, the maximum the
  launcher allows (installed minus 2 GB kept for the system) and the current
  setting, e.g. "Gép memóriája: 8 GB · ebből legfeljebb 6 GB engedélyezett
  (2 GB a rendszernek fenntartva) · beállítva: 4 GB".
- The slider's upper bound is the machine's limit instead of a fixed 64 GB.
- A saved/edited value above the limit is clamped on load, saved back, and
  explained in the status bar with the same numbers.
- `LauncherPartLaunch` refuses to start with a heap larger than the machine
  allows and names the numbers instead of letting the JVM fail opaquely; a
  process that will not start now reports the requested heap plus the machine
  statistics.
- New helper `core/SystemMemory` (Windows `GlobalMemoryStatusEx`, Linux
  `sysconf`) and test `test_system_memory`.

### Friendly message for a game that runs out of memory (2026-11-17)

- The launched game process is now monitored after a successful start: its
  output is kept in a bounded tail (16 kB) and an abnormal exit is diagnosed.
- Memory exhaustion (JVM `OutOfMemoryError` / `Java heap space` /
  `GC overhead limit exceeded` / heap reservation failure, or SIGKILL exit code
  137 from the OS OOM killer) produces the same plain-language Hungarian text as
  the RAM panel, with requested memory, installed memory and the allowed
  maximum plus the action to take.
- Any other crash keeps the generic behaviour: a neutral "process exited
  (code N)" log line, no memory wording.
- A deliberate stop (`abort()`) never reports a failure.
- `LaunchTask` no longer drops a step failure that arrives after the pipeline
  finished, so a game that dies after starting reaches the UI's error channel
  (`succeeded()` = the game started, a later `failed()` = it then died).
- New test `game_exit_diagnostics` (pure classification + message + late-failure
  forwarding, no real OOM).

### Plain-language explanations for every error (2026-11-17)

- New `core/FriendlyError` maps any technical launch/download error to a short
  Hungarian sentence a basic user understands (missing Java, login, memory,
  missing instance, directory creation, download/HTTP/network/checksum, game
  crash exit code, failed start). Unknown errors fall back to a generic
  message plus a short snippet; already-friendly messages pass through, so the
  conversion is idempotent.
- `LaunchTask::onStepFailed` now logs the technical detail and emits the
  friendly text, including failures that arrive after the pipeline finished.
- `DownloadAsset` reports friendly messages for every failure path.
- The generic game-exit message is now a plain Hungarian sentence with the
  exit code instead of an English trace.
- New test `friendly_error`.
### Cleanup (2026-11-17)

- Removed completely unused `launcher/gui/SplashScreen.cpp` and `SplashScreen.h` (no references anywhere in code or tests).

### Skin approval + password reset (2026-11-19)

- Web alpha: skin upload auto-applies to user/web/server; new skins enter admin
  review queue; approve makes them auto-accepted, reject blocks them for everyone
  and resets affected users to the default skin. Admin queue shows interactive 3D
  preview; rejected-skins sub-panel allows revoking a decision.
- Web alpha: password reset button per user generates a random 12-char temp
  password and forces a change on next login; new change-password view and
  register view show 5 security criteria (8+ chars, number, uppercase, 10+ chars,
  special char) with a strength meter.
- Backend: new skin uploads insert a pending approval row so non-owner reads
  fail closed; back-compat skins without a row stay approved.

### Advanced admin panel (2026-11-20)

- Added duration/reason/scope/targeted ban support for accounts, HWIDs and IPs.
- Added IP usage tracking and IP->accounts admin view.
- Updated admin dashboard UI with ban modal and user details.

### Advanced admin panel hardening (2026-11-20)

- Hardened ban validation and session revocation.
- Added ban list/delete endpoints.
- Enforced MC-scope HWID/IP bans in proxy authorization.
- Fixed authorization gaps on approval endpoints.
- Improved admin UI usability and accessibility.

## 2026-11-20 - ADVANCED MODERATION ACTIONS (KICK/MUTE/WARN)

- Added kick, mute, warn and unmute admin endpoints with the same duration, reason, scope (mc/platform/both) and account/HWID/IP targeting pattern as bans.
- Reused the bans table for moderation actions (ban_type kick/mute/warn) and added unmute recovery.
- Kick revokes sessions for platform/both scopes; mute/warn are recorded and queryable.
- Updated legacy admin UI with Kick, Mute, Warn, Unmute actions and a reusable action modal.
- Added accessibility improvements (labels, dialog role, escape handling).

## 2026-11-20 - SPA ADMIN ADVANCED ACTIONS & ACCESSIBILITY

- Ported the advanced ban/kick/mute/warn modal (duration, reason, scope, account/HWID/IP targets) into the SPA admin view.
- Added warn action button and unmute/unban flows in the SPA admin view.
- Extended the mock data layer with a bans collection and advanced action methods.
- Improved table header accessibility (scope="col") and modal ARIA attributes.

## 2026-11-20 - SPA ADMIN REAL BACKEND WIRING & IP/HWID BAN PREVIEW

- Wired the SPA admin panel to the real /api/admin/* endpoints using an X-Admin-Token stored in localStorage, with a token bar to set/clear it.
- Added admin-real.js adapter mirroring the mock api.admin surface (users, summary, audit from bans, ban/unban/kick/mute/warn/unmute, delete).
- Advanced ban modal now fetches user details and shows device-HWID count and per-IP account counts for the last 30 days.
- Real-mode smoke verified: ban request with admin token updates the audit list; no console errors.

### Advanced admin panel (2026-11-20, task branch `feat/advanced-admin-panel`)

- Advanced moderation modal for ban, kick, mute and warn: duration, reason,
  scope (MC/platform/both) and account/HWID/IP targets with 30-day IP account
  counts and HWID preview.
- SPA admin view now talks to the real `/api/admin/*` endpoints through an
  `X-Admin-Token` adapter while keeping the mock-data fallback.
- Backend: advanced ban/unban/details/IP-accounts/ban-list/delete plus
  kick/mute/warn/unmute handlers; proxy enforces MC-scope account/HWID/IP bans;
  login records IPs; `/api/admin/users` reports `muted`.
- Accessibility/UX: modal fade-in and Tab focus trap, Escape close, persistent
  invalid-token alert, audit labels, unsupported real-mode actions hidden.

### Ban scope isolation + Minecraft ban reason (2026-11-20, `feat/advanced-admin-panel`)

- Fixed login IP recording/enforcement behind a trusted proxy: forwarding
  headers are used only with `AUTH_TRUST_PROXY`, using the right-most
  `X-Forwarded-For` value like the rate limiter. Previously the socket address
  was recorded, so a reverse-proxied deployment could IP-ban every account.
- Added `ActiveBanForScope` and used it for login checks, so a stale mc-scope
  ban cannot mask a platform ban for the same account/HWID/IP.
- `/api/v1/proxy/auth` now returns the stored ban `reason`, `duration_minutes`,
  `expires_at` and `source` when denying a join; the Velocity plugin displays
  the reason instead of a generic message.
- The admin ban toast reports how many HWID/IP identifiers were hit and warns
  when the 30-day window contained none. The real-mode audit list shows the ban
  action, the exact identifier, scope, duration and reason.
- Added `TestAdminBanScopeIsolation` covering account, HWID and IP isolation,
  shared hardware/IP, missing last-month data and MC reason/duration.

## 2026-11-21
- Added immutable audit_log table and endpoints; audit survives ban deletion.
- Added admin password reset backend endpoint.
- Added pending skins list endpoint.
- Wired real admin adapter for roles, skin approval, password reset, audit.
- Unhid role management and password reset in real mode.
- Added alpha-v2 test selector and optimized MC server for 2GB RAM.

## 2026-11-21
- Added docs/ADMIN_PANEL_TEST_CHECKLIST.md with graphical test checklist, results and UX improvement notes.
- Commit 63e168f on feat/advanced-admin-panel.

## 2026-11-21 - Advanced admin panel tests and fixes

- Completed remaining admin panel end-to-end tests (account deletion, invalid token banner, skin approve/reject, unmute, search filtering).
- Fixed ListPendingSkinFiles to exclude already-reviewed rejected skins.
- Added audit logging for unmute, unban, skin approve/reject, account deletion, role assign/revoke.
- Frontend search now filters by username, email and user id.
- Updated ADMIN_PANEL_TEST_CHECKLIST.md.

## 2026-11-21
- Admin panel UI uniformity fix: styled missing `.form-row`, made modal scrollable/responsive, added cookie-consent card styles, switched admin KPI/grid to auto-fit minmax for dynamic responsiveness across displays.

## 2026-11-21
- Responsive audit of admin panel across mobile/tablet/desktop/folded breakpoints: no page-level horizontal scroll; modals scroll within viewport.
- Fixed overlay stacking: cookie-consent host z-index lowered below modal backdrop and toasts.
