# CHANGELOG

## Unreleased

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
