# AI CONTEXT

## Project identity

This is a private Minecraft network launcher/client ecosystem.

First target:
Windows 10/11
Minecraft 1.21.11
Fabric 0.19.2
Java 24

## Current phase

PHASE 0 — PROJECT GOVERNANCE / ARCHITECTURE

The Prism minimal-core / licensing checkpoint is RESOLVED at architecture
level (DECISION-003, DECISION-004; see `docs/PRISM_MINIMAL_CORE.md` and
`docs/PRISM_ARCHITECTURE_REVIEW.md`).

Production implementation remains blocked until:
- the owner signs off DECISION-004 (GPLv3 distribution model), and
- DECISION-005 (metadata source) is resolved

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
- Tracked: governance/state documentation only
- Untracked by design: `reference/` (DECISION-002) and local tooling state
- Architecture documentation now includes `docs/PRISM_MINIMAL_CORE.md` and
  `docs/PRISM_ARCHITECTURE_REVIEW.md` (DECISION-003/004/005)
- No production source code exists in the repository yet

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

No production implementation approved yet.
The Prism subsystem classification is complete (DECISION-003). The next
implementation tasks, once DECISION-004 is signed off and DECISION-005 is
resolved, are the launcher fork bootstrap and the launcher core design.

## Handoff rule

This file must always describe the current real repository state.

Never claim functionality that does not exist.
