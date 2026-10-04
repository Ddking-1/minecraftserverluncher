# MASTER PLAN — PHASE 3

This document is the ordered task backlog for Phase 3 of the autonomous
"marathon" development runs. It builds on the completed T1–T45 backlog
(`docs/MASTER_PLAN_EXTENDED.md`) and continues the same workflow.

Role split (per AGENTS.md / docs/WORKFLOW_RULES.md):
- Architect: planning, integration, build/test execution, Git management.
  The Architect never authors product code by hand.
- Worker (`lm-studio/deepseek-coder-v2-lite-instruct`): all code generation,
  code modification and defect fixing, invoked via `opencode run`.

Error policy: after 8 failed Worker attempts on one task, record the failure in
`docs/ERROR_LOG.md`, mark the task as blocked, and continue with the next task.

Status legend: `DONE`, `BLOCKED`, `PENDING`.

## Tasks (P3-01 … P3-10)

| ID | Title | Area | Status |
|----|-------|------|--------|
| P3-01 | Launcher SettingsStore (load/save launcher.json, defaults, atomic write) + tests | launcher/core | PENDING |
| P3-02 | Mod list scanner (scan mods/ dir, enabled/disabled, sorted) + tests | launcher/minecraft | PENDING |
| P3-03 | Skin cache (id → local file, SHA-256 verified) + tests | launcher/core | PENDING |
| P3-04 | Launcher rotating file logger (size rotation, keep N) + tests | launcher/core | PENDING |
| P3-05 | Update checker (semantic version compare + update decision) + tests | launcher/updates | PENDING |
| P3-06 | File integrity verifier (SHA-256 of file vs expected) + tests | launcher/core | PENDING |
| P3-07 | Backend request-ID + structured logging middleware + tests | server/api | PENDING |
| P3-08 | Backend auth/token extra tests + DB expiry indexes | server | PENDING |
| P3-09 | Launcher memory/JVM settings validation + tests | launcher/core | PENDING |
| P3-10 | Phase 3 final verification + documentation sync | docs | PENDING |

## Task details

### P3-01 — Launcher SettingsStore
Load and save `launcher.json` through a small project-owned store built on
`InstancePersistence` primitives. Applies documented defaults for missing
fields, refuses to load malformed JSON, and writes atomically. Unit test:
round-trip, defaults, missing file, malformed file.

### P3-02 — Mod list scanner
Scan an instance `mods/` directory and return a sorted list of mod entries
(name, file name, enabled). A `.jar` file is enabled; a `.jar.disabled` file is
disabled. Missing directory yields an empty list. Unit test with a temporary
directory.

### P3-03 — Skin cache
Cache a skin asset under a content-addressed file name derived from its
SHA-256. Verify the cached file hash before reuse. Unit test: store/lookup,
hash mismatch rejection, missing entry.

### P3-04 — Rotating file logger
Append log lines to a file; when the file exceeds a size threshold, rotate to
`.1`, `.2`, … and keep at most N files. Unit test: rotation, retention cap.

### P3-05 — Update checker
Compare the running launcher version against a parsed `UpdateManifest`
component version (and `min_launcher`). Provide a semantic-version comparator
and an update decision. Unit test: equal, newer, older, prerelease-ish.

### P3-06 — File integrity verifier
Compute the SHA-256 of a file and compare it against an expected digest.
Unit test: match, mismatch, missing file.

### P3-07 — Backend request-ID middleware
Accept or generate a request id, expose it to downstream handlers, echo it in
the `X-Request-Id` response header, and include it in the request log line.
Unit test: generated id present and unique, provided id echoed.

### P3-08 — Backend auth/token tests + indexes
Add indexes on `access_tokens(expires_at)` and `sessions(refresh_expires_at)`
to the schema, and add focused tests for token digest determinism/uniqueness
and password verification failure. No behaviour change beyond the indexes.

### P3-09 — Memory/JVM settings validation
Validate min/max memory bounds and their relationship, and reject unsafe JVM
argument strings. Integrate the validator into `SettingsStore` so invalid
loaded/updated settings are clamped or rejected. Unit test: bounds, ordering,
unsafe args.

### P3-10 — Phase 3 final verification + documentation sync
Run full backend `go build`/`go vet`/`go test` and full launcher Release build
+ CTest. Synchronize AI_CONTEXT / WORK_STATE / PROGRESS / CHANGELOG and mark
this plan complete.
