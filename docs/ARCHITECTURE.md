# ARCHITECTURE

This file is the authoritative architecture document.

IMPORTANT:
No implementation agent may silently replace architecture decisions.

## Approved direction (DECISION-001)

The system is:

**required Prism core + own launcher core + own backend**

It is NOT "full Prism Launcher + custom layer".

Component ownership:

| Component | Ownership | Note |
|---|---|---|
| Launcher shell/instance/launch core | Prism Launcher fork, trimmed | only the Prism functionality actually required by this project remains in the distributed launcher |
| Launcher core (accounts, sessions, devices, trusted launch authorization, server profiles, updater, integrity) | project-owned | lives in the project's own layer, not in forked Prism code |
| Auth/API backend | project-owned | Go + SQLite |
| Velocity plugin/integration | project-owned | |
| Geyser / Floodgate | external | integration/dependency, not project code |
| LimboAPI / LimboAuth / LimboFilter | external | integration/dependency, not project code |

Minimal core boundary (DECISION-003):
The component-level KEEP / KEEP BUT ADAPT / REPLACE / NOT NEEDED inventory of
Prism subsystems is defined in `docs/PRISM_MINIMAL_CORE.md` and adopted by
DECISION-003. Architectural decisions and dependency/licensing risks are
combined in `docs/PRISM_ARCHITECTURE_REVIEW.md`.

Scope of the first release:
- first launcher target: Windows 10/11
- first release exposes one server
- the internal model must already support additional servers

Blocking checkpoint:
RESOLVED at architecture level (2026-09-05). The Prism/GPL analysis is
recorded in `docs/PRISM_ARCHITECTURE_REVIEW.md` and DECISION-004 (public
launcher source repository; private central repository). Production launcher
implementation may start only after: owner legal sign-off on DECISION-004,
and resolution of DECISION-005 (metadata source).

## Core layers

The system is conceptually composed of:

1. Launcher
2. Authentication/API backend
3. Velocity integration
4. Java fallback authentication
5. Geyser/Floodgate Bedrock integration
6. Minecraft servers
7. Skin system
8. Update/artifact infrastructure
9. Admin/web system

## Trust boundaries

### Trusted launcher Java path

Launcher
→ Auth backend
→ device/session verification
→ short-lived launch authorization
→ Minecraft
→ Velocity
→ server

### Normal Java path

Minecraft
→ Velocity
→ trusted authorization check
→ if absent/invalid:
→ LimboFilter
→ CAPTCHA
→ LimboAuth
→ server

### Bedrock path

Bedrock
→ Geyser/Floodgate
→ Velocity
→ server

The Bedrock path is not part of Java LimboAuth.

## Account architecture

A common logical account system is required.

Conceptual entities include:
- Account
- JavaIdentity
- BedrockIdentity
- Device
- Session
- LaunchAuthorization
- Skin

The exact schema is determined by the architecture phase and must be documented before implementation.

## Backend

Preferred:
Go + SQLite.

The backend should remain lightweight.

The backend should primarily handle:
- accounts
- sessions
- device registration
- authorization
- metadata
- manifests
- admin
- skin metadata
- account linking

Large artifact downloads should normally be externalized.

## Server profiles

The launcher must not hardcode the first server into source code.

A server profile should be able to define:
- server identity
- Minecraft version
- loader
- Java requirement
- modpack
- JVM profile
- update profile
- connection target

The first release may expose only one server in the UI.

## External components

Existing open-source projects must be evaluated before recreating their functionality.

Reference repositories are under `reference/`.

Prism Launcher: evaluated — fork (GPLv3), trimmed per DECISION-003, distributed
with public source per DECISION-004.

The architecture must clearly state whether each dependency is:
- external runtime dependency
- library dependency
- plugin dependency
- fork
- wrapper
- reference only
- rewritten independently

## Security principle

The client must not be the final authority.

The server must validate trusted-launcher authorization.

The client may perform integrity checks, but server-side authorization and gameplay protection remain authoritative.

## Performance principle

Use local cache and manifests whenever possible.

Avoid unnecessary:
- polling
- downloads
- hashing
- backend traffic
- background CPU use

## Multi-server principle

The first UI may expose one server.

The internal model must not assume there will ever be only one.

## Architecture status

Filled in so far:
- component ownership and the Prism minimal-core boundary
  (DECISION-003, `docs/PRISM_MINIMAL_CORE.md`)
- launcher legal/licensing boundaries
  (DECISION-004, `docs/PRISM_ARCHITECTURE_REVIEW.md`)

Still to be filled in before production implementation:
- launcher core design (accounts/sessions/devices, launch authorization,
  server profiles, updater/integrity, automatic RAM algorithm)
- API boundaries and protocols (launcher-backend, Velocity plugin)
- deployment topology
- data flow
- DECISION-005 resolution (metadata source)
