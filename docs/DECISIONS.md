# DECISIONS

This file records architectural decisions.

## Decision format

### DECISION-NNN

Date:
Status:
Decision:
Reason:
Alternatives:
Consequences:
Affected files:
Reversible:
Legal review required:

## Recorded decisions

### DECISION-001 — Launcher foundation direction

Date: 2026-09-05

Status: APPROVED (direction) / LEGAL REVIEW REQUIRED (licensing)

Decision:
The launcher is built as a Prism Launcher fork, but the approved direction is
**required Prism core + own launcher core + own backend**.

This explicitly means:
- only the Prism functionality actually required by this project should remain in the distributed launcher
- unused Prism subsystems are to be removed/disabled over time, not carried along
- all project-specific launcher logic (accounts, sessions, devices, trusted launch authorization, server profiles, updater, integrity) lives in the project's own launcher core layer
- the backend is project-owned (Go + SQLite)
- the Velocity plugin is project-owned
- Geyser, Floodgate and the Limbo components are external integrations/dependencies, not project-owned code

The architecture must NOT be described as "full Prism Launcher + custom layer".

Reason:
A trimmed Prism core reduces attack surface, maintenance cost, binary size and the scope of the GPL source-distribution obligation, and keeps project-specific security logic in code the project fully controls.

Alternatives:
- launcher written from scratch (rejected for now: high cost, mature Minecraft instance/launch logic already exists in Prism)
- unmodified full Prism plus a plugin/custom layer (rejected: carries unused subsystems and blurs the ownership boundary)

Consequences:
- a component-level inventory of "required Prism functionality" must be produced before implementation
- the Prism/GPL licensing analysis in `docs/LICENSE_AUDIT.md` remains a blocking checkpoint
- the launcher core boundary must be defined architecturally before code is written

Affected files:
`docs/ARCHITECTURE.md`, `docs/AI_CONTEXT.md`, `docs/LICENSE_AUDIT.md`, `docs/THIRD_PARTY_COMPONENTS.md`

Reversible: yes, before implementation starts

Legal review required: YES (Prism Launcher GPL obligations, distribution and public source availability)

### DECISION-002 — `reference/` is not tracked in Git

Date: 2026-09-05

Status: APPROVED

Decision:
The local third-party reference sources under `reference/` are excluded from version control via `.gitignore` and are not committed to the private repository.

Reason:
They are read-only research material (~67 MB, ~4300 files across 8 upstream projects), not product source. Committing them would mix third-party code into the project history before any license decision exists, bloat the repository, and make later license/provenance auditing harder.

Alternatives:
- commit them directly (rejected: history pollution, licensing/provenance risk)
- Git submodules pinned to exact upstream commits (deferred: preferred future option because it records the exact inspected commit, which the license audit needs)

Consequences:
- each agent/developer must obtain the reference sources locally
- exact upstream commit identification for the license audit is currently NOT captured by Git and must be recorded manually in `docs/LICENSE_AUDIT.md`

Affected files:
`.gitignore`, `AGENTS.md`, `docs/THIRD_PARTY_COMPONENTS.md`

Reversible: yes (may be replaced by pinned submodules)

Legal review required: no

## Rules

Do not silently overwrite decisions.

If a decision changes:
1. create a new decision
2. reference the old decision
3. explain why it changed

Important unresolved decisions must be explicitly marked:

STATUS: PENDING

Legal uncertainty must be marked:

LEGAL REVIEW REQUIRED
