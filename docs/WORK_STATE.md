# WORK STATE

Current phase:
PHASE 2 — BACKEND HARDENING + LAUNCHER CORE FEATURES (IN PROGRESS)

Current task:
Continue the ordered marathon backlog in `docs/MASTER_PLAN_EXTENDED.md`
(T1-T45). Backend and launcher tasks through T41 are complete except the
Velocity skeleton (T11, blocked on DECISION-004/005) and the documentation
sync (T44) / final verification (T45).

Status:
Green. Backend: `go build`, `go vet`, `go test ./...` all pass. Launcher:
CMake configure + Release build with Qt 6.6.3 MSVC 2022 succeed; CTest runs
5 tests (hwid, settings, update_manifest, skin_model, backend_contract), all
pass.

Branch:
master (all task branches fast-forward merged and pushed).

Last completed steps:
- T12/T29 config loader tests, T31 session service tests, T30 device service
  tests, T28 rate limiter tests.
- T16 backend security review (`docs/SECURITY_REVIEW.md`).
- T22-T27 account profile, password change, session list/revoke endpoints and
  their tests.
- T33-T35 request logging, panic recovery and security header middleware.
- T40 `/readyz` readiness endpoint.
- T32 hardened HTTP server timeouts + graceful shutdown.
- T36-T39 launcher HTTP client timeout, backend status-code error mapping, CLI
  instance filter, `--version`/`--help`.
- T19/T20 update manifest parser + tests; T21 skin model + tests.
- T13 BackendAuthClient contract test (local fake HTTP server).
- T41 `docs/API.md` endpoint reference.
- T14/T42/T43 CI workflows (backend on ubuntu-latest, launcher on
  windows-latest).

Current step:
Documentation sync (T44) and final verification (T45).

Next step:
1. T44: synchronize AI_CONTEXT / WORK_STATE / PROGRESS / CHANGELOG.
2. T45: full backend `go test ./...` + full launcher Release build + CTest,
   then final report.
3. T11 Velocity plugin remains BLOCKED on DECISION-004 legal sign-off and
   DECISION-005 (metadata source).

Repository facts:
- default branch: master
- remote: origin (private GitHub repo Ddking-1/minecraftserverluncher)
- reference/ is local-only, never committed (DECISION-002)
- backend source exists under server/ (Go module minecraftserverluncher/server)
- launcher core exists under launcher/ (C++ / Qt6, builds and tests green)
- velocity directory is scaffolded but empty (no code yet)

Last commit:
See docs/PROGRESS.md / `git log --oneline`.
