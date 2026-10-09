# AI CONTEXT

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
baseline is 200 passing tests). The backend also contains the Skin API handlers
(upload/get/avatar/delete over `server/skins` validation + storage) and the
Pterodactyl webhook handler (HMAC signature verification, 401 on a missing or
invalid signature); both are unit-tested but are not wired in `server/main.go`
yet, and the Skin API `DELETE /api/v1/skins/{id}` route still conflicts with the
legacy `DELETE /api/v1/skins/{variant}` pattern.

The C++/Qt6 launcher core is implemented on `master` and the build baseline is
GREEN: CMake configure and a Release build with Qt 6.6.3 MSVC 2022 succeed and
produce `MinecraftLauncher.exe`. CTest runs five tests (hwid, settings,
update_manifest, skin_model, backend_contract), all passing. Newer launcher
work includes the update manifest parser, skin model, backend status-code error
mapping, a request timeout, and CLI filter/version options.

Velocity work remains blocked until the owner signs off DECISION-004 (GPLv3
distribution model) and DECISION-005 (metadata source) is resolved.

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
timeouts with graceful shutdown. `go build`, `go vet` and `go test ./...` pass.
See `docs/API.md`, `docs/SECURITY_REVIEW.md` and `server/README.md`.

Launcher core (`launcher/`, C++/Qt6) is on `master`: `launcher/core`
(BaseInstance, InstanceList, InstancePersistence, SkinModel, SkinCache,
SettingsStore, RotatingFileLogger, FileIntegrity, Version), `launcher/minecraft`
(MinecraftInstance, PackProfile, ModScanner), `launcher/auth`
(BackendAuthClient with a 15s timeout and status-code error mapping,
TokenStorage using Windows DPAPI, SessionManager), `launcher/launch`
(LaunchTask/LaunchStep framework), `launcher/updates` (UpdateManifest parser,
UpdateChecker) and `launcher/game` (JavaDetector, MinecraftLauncher). CMake
configure and `cmake --build ... --config Release` succeed with Qt 6.6.3
MSVC 2022; CTest runs 11 passing tests (hwid, settings, update_manifest,
skin_model, backend_contract, settings_store, mod_scanner, skin_cache,
rotating_logger, update_checker, file_integrity). Library/asset download steps
remain stubs pending DECISION-005.

Velocity implementation remains blocked until DECISION-004 is signed off and
DECISION-005 is resolved.

## Handoff rule

This file must always describe the current real repository state.

Never claim functionality that does not exist.
