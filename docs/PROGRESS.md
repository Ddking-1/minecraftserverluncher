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
