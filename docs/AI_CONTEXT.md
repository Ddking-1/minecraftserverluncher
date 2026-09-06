# AI CONTEXT

## Project identity

This is a private Minecraft network launcher/client ecosystem.

First target:
Windows 10/11
Minecraft 1.21.11
Fabric 0.19.2
Java 24

## Current phase

PHASE 1 — BACKEND SCAFFOLD / AUTH API + LAUNCHER CORE SCAFFOLD (IN PROGRESS)

The Prism minimal-core / licensing checkpoint is RESOLVED at architecture
level (DECISION-003, DECISION-004; see `docs/PRISM_MINIMAL_CORE.md` and
`docs/PRISM_ARCHITECTURE_REVIEW.md`).

The backend scaffold and core auth API are implemented on branch
`backend/scaffold-and-auth` (Go + SQLite). A minimal C++ launcher core scaffold
is implemented on branch `launcher/core-scaffold` (instance management,
launch-task framework, backend auth client). Launcher compilation is BLOCKED
until Qt 6 is installed in the build environment. Velocity work remains
blocked until the owner signs off DECISION-004 (GPLv3 distribution model) and
DECISION-005 (metadata source) is resolved.

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
- Tracked: governance/state documentation plus the backend scaffold and auth
  API under `server/` (branch `backend/scaffold-and-auth`) and the launcher
  core scaffold under `launcher/` (branch `launcher/core-scaffold`)
- Untracked by design: `reference/` (DECISION-002) and local tooling state
- Architecture documentation includes `docs/PRISM_MINIMAL_CORE.md` and
  `docs/PRISM_ARCHITECTURE_REVIEW.md` (DECISION-003/004/005)
- Backend source exists under `server/` (Go module `minecraftserverluncher/server`)
- Launcher core scaffold exists under `launcher/` (C++ / Qt, not yet compiled)

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

Backend scaffold and core auth API implemented on branch
`backend/scaffold-and-auth` (Go + SQLite): account model with bcrypt password
hashing, server-side session model with expiry and revocation, device
identity mapping, and registration/login/refresh/logout/logout-all endpoints.
SQLite schema initialization and minimal tests are included. See
`server/README.md`.

A minimal C++ launcher core scaffold is implemented on branch
`launcher/core-scaffold`: `launcher/core` (BaseInstance, InstanceList as a
QAbstractListModel, InstancePersistence), `launcher/minecraft`
(MinecraftInstance, PackProfile), `launcher/auth` (BackendAuthClient matching
the server auth API, TokenStorage, SessionManager — replaces Prism's
Mojang/MSA flow), and `launcher/launch` (LaunchTask/LaunchStep framework with
ClaimAuthSession, ResolveLibraries, ResolveAssets, ExtractNatives, VerifyJava,
LauncherPartLaunch steps). See `launcher/README.md`. The C++ code is NOT yet
compiled: no Qt 6 toolchain is installed in the build environment. Library and
asset download steps are stubs pending DECISION-005; token storage is
first-pass plaintext pending OS-protected credential storage (docs/AUTH.md).

Velocity implementation remains blocked until DECISION-004 is signed off and
DECISION-005 is resolved.

## Handoff rule

This file must always describe the current real repository state.

Never claim functionality that does not exist.
