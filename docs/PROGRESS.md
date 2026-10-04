# PROGRESS

## Initial state

The project requirements have been collected.

The project will use:
- a private central GitHub repository for development
- separate branches for AI work
- local read-only reference repositories
- persistent project documentation
- multiple AI agents
- local AI for a substantial amount of implementation work

No production implementation should be assumed complete.

Each agent must append meaningful progress here.

Format:

## YYYY-MM-DD HH:MM — TASK

Agent/model:
Branch:
Task:
Started:
Completed:
Files:
Tests:
Current status:
Next:
Manual user action:

## 2026-09-05 — REPOSITORY / PROJECT-STATE SYNCHRONIZATION

Agent/model: project governance agent
Branch: architecture/project-state
Task: Synchronize the repository with the approved project state; governance documentation only, no production code.
Started: 2026-09-05
Completed: 2026-09-05

Files:
- .gitignore (new — excludes reference/ and local tooling state)
- AGENTS.md (permanent agent rules 21–26 + Git/GitHub ownership rule, master/main clarification, reference/ scanning ban)
- docs/AI_CONTEXT.md (approved direction + repository state)
- docs/ARCHITECTURE.md (approved direction and component ownership table)
- docs/DECISIONS.md (DECISION-001, DECISION-002)
- docs/WORK_STATE.md (current state and next step)
- docs/PROGRESS.md (this entry)
- docs/CHANGELOG.md (unreleased governance entry)

Tests:
No build/test tooling exists in the repository yet (documentation-only change).
Checks performed: git status, full git diff review, verification that reference/ and .abacusai/ are excluded from the commit.

Current status:
Governance baseline committed and pushed on branch architecture/project-state.
Phase 0 continues. No production implementation exists or is approved.

Next:
Architecture/licensing checkpoint — resolve the Prism Launcher GPL analysis,
inventory the Prism functionality to retain, define the launcher core boundary.

Manual user action:
Review and merge branch architecture/project-state into master (agents must not
work directly on the default branch).

## 2026-09-05 — PRISM MINIMAL-CORE ARCHITECTURE / LICENSING CHECKPOINT

Agent/model: Prism minimal-core architecture agent
Branch: architecture/prism-minimal-core
Task: Determine the minimum Prism functionality required for this project, resolve the Prism licensing boundaries, and define the launcher core boundary (the DECISION-001 checkpoint).
Started: 2026-09-05
Completed: 2026-09-05

Files:
- docs/PRISM_MINIMAL_CORE.md (new — subsystem inventory with KEEP / KEEP BUT ADAPT / REPLACE / NOT NEEDED classifications)
- docs/PRISM_ARCHITECTURE_REVIEW.md (new — architectural decisions combined with dependency and licensing risks)
- docs/DECISIONS.md (DECISION-003, DECISION-004, DECISION-005)
- docs/ARCHITECTURE.md (minimal-core boundary note, checkpoint status, architecture status)
- docs/LICENSE_AUDIT.md (Prism entry: analysis complete at architecture level, open audit items)
- docs/AI_CONTEXT.md (phase and repository state synchronized)
- docs/WORK_STATE.md (task state updated)
- docs/PROGRESS.md (this entry)
- docs/CHANGELOG.md (unreleased entry)

Tests:
No build/test tooling exists in the repository yet (documentation-only change).
Checks performed: targeted read-only inspection of reference/prism-launcher
(no recursive scanning, AGENTS.md rule 24); every file path cited in the new
documents was verified to exist in the snapshot; Prism license verified by
reading upstream COPYING.md; git status and the full diff reviewed before
committing.

Current status:
Checkpoint resolved at architecture level. DECISION-004 still requires owner
legal sign-off; DECISION-005 (metadata source) is PENDING. No production
code exists or is approved yet.

Next:
Owner legal sign-off on DECISION-004; resolve DECISION-005; launcher fork
bootstrap (pin upstream commit, public launcher repository, vendored license
audit, trim plan); then launcher core design.

Manual user action:
1. Legal sign-off on DECISION-004 (GPLv3 distribution / public launcher
   source repository model).
2. Review and merge branch architecture/prism-minimal-core into master.

## 2026-09-05 17:02 — BACKEND SCAFFOLD / AUTH API

Agent/model: local backend implementation agent
Branch: backend/scaffold-and-auth
Task: Create the repository scaffold and implement the core backend auth API
(Go + SQLite). No Microsoft auth, UI, Velocity plugin or LimboAuth.
Started: 2026-09-05
Completed: 2026-09-05

Files:
- Scaffold directories (tracked via .gitkeep): launcher/core, launcher/auth,
  launcher/minecraft, launcher/updates, server/api, server/auth,
  server/sessions, server/devices, server/database, velocity/plugin, tests
- server/go.mod, server/go.sum (Go module minecraftserverluncher/server)
- server/main.go (HTTP server bootstrap)
- server/config/config.go (environment-driven config; no hardcoded secrets)
- server/database/database.go, accounts.go, devices.go, sessions.go
  (SQLite schema init + account/device/session/access-token stores)
- server/auth/auth.go (bcrypt password hashing, opaque token gen/hash)
- server/sessions/service.go (session lifecycle)
- server/devices/service.go (device identity mapping)
- server/api/handlers.go, ratelimit.go (register/login/refresh/logout/
  logout-all endpoints, per-IP rate limiting)
- server/auth/auth_test.go, server/database/database_test.go,
  server/api/handlers_test.go (minimal tests)
- server/.env.example, server/.gitignore, server/README.md
- docs/AI_CONTEXT.md, docs/WORK_STATE.md, docs/PROGRESS.md (this entry),
  docs/CHANGELOG.md

Tests:
go build ./... (pass), go vet ./... (pass), go test ./... (pass; 11 tests:
auth 2, database 5, api integration 4). No test tooling existed before this
task; the Go checks above are the checks appropriate to the changed area.

Current status:
Backend scaffold and core auth API implemented and pushed on branch
backend/scaffold-and-auth. Launcher and Velocity directories are scaffolded
but empty. No Microsoft auth, UI, Velocity plugin or LimboAuth implemented.

Next:
Owner review/merge of branch backend/scaffold-and-auth; then launcher
bootstrap and launcher core design once DECISION-004 sign-off and DECISION-005
resolution are complete.

Manual user action:
1. Review and merge branch backend/scaffold-and-auth into master.
2. Legal sign-off on DECISION-004 (GPLv3 distribution / public launcher
   source repository model) remains outstanding.

## 2026-09-05 — LAUNCHER CORE SCAFFOLD (C++ / Qt)

Agent/model: local launcher core implementation agent
Branch: launcher/core-scaffold
Task: Implement the minimal C++ instance-management and launch-task framework
(launcher/core, launcher/minecraft, launcher/auth, launcher/launch) with a
backend auth client replacing Prism's Mojang/MSA flow, and a minimal launch
sequence. No Microsoft checks.
Started: 2026-09-05
Completed: 2026-09-05

Files:
- launcher/CMakeLists.txt (Qt6 Core/Gui/Widgets/Network; core/minecraft/auth
  subdirectories)
- launcher/main.cpp (minimal entry point exercising the instance list)
- launcher/README.md (structure, compile status, notes)
- launcher/core/BaseInstance.h/.cpp, InstanceList.h/.cpp,
  InstancePersistence.h/.cpp, InstanceTypes.h, CMakeLists.txt
- launcher/minecraft/MinecraftInstance.h/.cpp, PackProfile.h/.cpp,
  CMakeLists.txt (combined minecraft + launch library)
- launcher/auth/BackendAuthClient.h/.cpp, SessionManager.h/.cpp,
  TokenStorage.h/.cpp, CMakeLists.txt
- launcher/launch/LaunchTask.h/.cpp, LaunchStep.h/.cpp,
  steps/{ClaimAuthSession,ResolveLibraries,ResolveAssets,ExtractNatives,
  VerifyJava,LauncherPartLaunch}.h/.cpp, java/{JavaInstall,JavaVersion}.h,
  java/JavaVersion.cpp
- Removed obsolete launcher/{core,auth,minecraft}/.gitkeep
- docs/AI_CONTEXT.md, docs/WORK_STATE.md, docs/PROGRESS.md (this entry),
  docs/CHANGELOG.md

Tests:
No Qt toolchain is installed in this environment, so the C++ code could not be
compiled. Checks performed: CMake configure dry-run reaches the Qt
find_package step cleanly (proving the CMake structure and MSVC toolchain
detection are valid; the only failure is the missing Qt 6); the C++ backend
auth client request/response contract was verified line-by-line against
server/api/handlers.go; all project-header #include paths were verified to
resolve; manual review for incomplete-type and upcast issues.

Current status:
Minimal launcher core scaffold implemented and pushed on branch
launcher/core-scaffold. Compilation is BLOCKED until Qt 6 is installed
(no Qt toolchain present). Library/asset download steps are stubs pending
DECISION-005. Token storage is first-pass plaintext pending OS-protected
credential storage (docs/AUTH.md).

Next:
Install Qt 6 and compile/verify the launcher core; resolve DECISION-005
(metadata source) to implement library/asset download; owner legal sign-off
on DECISION-004.

Manual user action:
1. Install Qt 6 (Core/Gui/Widgets/Network) so the launcher core can be
   compiled and verified.
2. Review and merge branch launcher/core-scaffold into master.
3. Legal sign-off on DECISION-004 remains outstanding.

## 2026-10-04 — LAUNCHER CMake / Qt6 BUILD FIX

Agent/model: launcher build agent (worker model lm-studio/deepseek-coder-v2-lite-instruct)
Branch: fix/launcher-cmake-qt6
Task: Make the launcher C++ core actually compile with the installed Qt 6.6.3
MSVC 2022 toolchain and remove the OpenSSL dependency from the HWID helper.
Started: 2026-10-04
Completed: 2026-10-04

Files:
- launcher/CMakeLists.txt (rewritten: Qt6 Core/Network, AUTOMOC, C++17, full
  source list, include dirs, Windows libs)
- CMakeLists.txt (root: removed duplicate add_subdirectory(launcher) and
  Qt5-first lookup)
- launcher/core/HWIDExtractor.cpp (OpenSSL sha.h -> Windows CNG bcrypt SHA-256)
- launcher/game/MinecraftLauncher.cpp (Qt6 QProcess::start void-return fix)
- docs/WORK_STATE.md, docs/PROGRESS.md (this entry), docs/CHANGELOG.md,
  docs/AI_CONTEXT.md

Tests:
Real build performed (no test framework for the launcher yet):
cmake -S launcher -B launcher/build-test -G "Visual Studio 17 2022" -A x64
  -DCMAKE_PREFIX_PATH="C:/Users/koppa_9my1ub2/6.6.3/msvc2019_64"  -> OK
cmake --build launcher/build-test --config Release                    -> OK
Artifact: launcher/build-test/Release/MinecraftLauncher.exe.
Build directory removed after verification.

Current status:
Launcher build baseline is GREEN on branch fix/launcher-cmake-qt6. All 18
translation units compile and link. No launcher unit tests exist yet.

Next:
Owner review/merge of branch fix/launcher-cmake-qt6; continue the approved
15-task plan (HWID verification, DPAPI token storage, launch steps); resolve
DECISION-005.

Manual user action:
1. Review and merge branch fix/launcher-cmake-qt6 into master.
2. Legal sign-off on DECISION-004 remains outstanding.

## 2026-10-04 - MARATHON TASKS T12-T41 (BACKEND + LAUNCHER + CI)

Agent/model: Architect (integration) + Worker (lm-studio/deepseek-coder-v2-lite-instruct, code generation)
Branch: master (per-task branches fast-forward merged and pushed)
Task: Execute the ordered marathon backlog docs/MASTER_PLAN_EXTENDED.md from T12 through T41.
Started: 2026-10-04
Completed: 2026-10-04

Files:
- server: config/config_test.go; sessions/service_test.go; devices/service_test.go;
  api/ratelimit_test.go, account.go, account_test.go, account_sessions_test.go,
  middleware.go, middleware_test.go, health.go, health_test.go, handlers.go;
  database/accounts.go, sessions.go, database.go; main.go
- launcher: auth/BackendAuthClient.cpp; core/Version.h, SkinModel.h/.cpp;
  updates/UpdateManifest.h/.cpp; tests/test_settings.cpp, test_update_manifest.cpp,
  test_skin_model.cpp, test_backend_contract.cpp; main.cpp; CMakeLists.txt
- docs: SECURITY_REVIEW.md, API.md, ERROR_LOG.md, MASTER_PLAN_EXTENDED.md
- .github/workflows/backend.yml, launcher.yml

Tests:
- Backend: go build ./..., go vet ./..., go test ./... -> all pass.
- Launcher: cmake configure + Release build (Qt 6.6.3 MSVC 2022) -> OK;
  ctest -> 5/5 pass (hwid, settings, update_manifest, skin_model, backend_contract).
- YAML workflows parsed with a YAML parser.

Current status:
T12-T41 complete except T11 (Velocity, blocked on DECISION-004/005). Remaining:
T44 documentation sync (this entry) and T45 final verification.

Next:
T45 final verification (full go test + full launcher build + ctest) and report.

Manual user action:
Legal sign-off on DECISION-004 and resolution of DECISION-005 remain outstanding
(blocks T11 Velocity work).

## 2026-10-04 - PHASE 3 (P3-01..P3-10)

Agent/model: Architect (integration) + Worker (lm-studio/deepseek-coder-v2-lite-instruct, code generation)
Branch: master (per-task branches fast-forward merged and pushed)
Task: Design and execute 10 new Phase 3 tasks (docs/MASTER_PLAN_PHASE3.md).
Completed: 2026-10-04

Files:
- launcher/core: SettingsStore.{h,cpp}, SkinCache.{h,cpp}, RotatingFileLogger.{h,cpp}, FileIntegrity.{h,cpp}
- launcher/minecraft: ModScanner.{h,cpp}
- launcher/updates: UpdateChecker.{h,cpp}
- launcher/tests: test_settings_store.cpp, test_mod_scanner.cpp, test_skin_cache.cpp,
  test_rotating_logger.cpp, test_update_checker.cpp, test_file_integrity.cpp
- launcher/CMakeLists.txt
- server/api: middleware.go, handlers.go, middleware_test.go
- server/auth/auth_extra_test.go
- server/database: database.go, database_test.go
- docs: MASTER_PLAN_PHASE3.md, WORK_STATE.md, CHANGELOG.md, AI_CONTEXT.md

Tests:
- Backend: go build ./..., go vet ./..., go test ./... -> all pass.
- Launcher: cmake configure + Release build (Qt 6.6.3 MSVC 2022) -> OK;
  ctest -> 11/11 pass.

Current status:
Phase 3 complete. T11 Velocity remains BLOCKED on DECISION-004/005.

Next:
No unblocked tasks. Await new instructions.

Manual user action:
Legal sign-off on DECISION-004 and resolution of DECISION-005 remain outstanding.
