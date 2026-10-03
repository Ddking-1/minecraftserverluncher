# MASTER PLAN EXTENDED

This document is the extended, ordered task backlog for the autonomous
"marathon" development runs. It extends the original 15-task plan with 30
additional medium-sized tasks. Tasks are executed strictly in order; each task
uses its own branch, is committed, pushed, and merged into `master` before the
next begins.

Role split (per AGENTS.md / docs/WORKFLOW_RULES.md):
- Architect: planning, integration, build/test execution, Git management.
  The Architect never authors product code by hand.
- Worker (`lm-studio/deepseek-coder-v2-lite-instruct`): all code generation,
  code modification and defect fixing, invoked via `opencode run`.

Error policy: after 8 failed Worker attempts on one task, record the failure in
`docs/ERROR_LOG.md`, mark the task as blocked, and continue with the next task.

Status legend: `DONE`, `BLOCKED`, `PENDING`.

## Original plan (T1–T15)

| ID | Title | Status |
|----|-------|--------|
| T1 | Launcher CMake/Qt6 build baseline | DONE (ab6c38e) |
| T2 | HWID via Windows CNG (no OpenSSL) | DONE (ab6c38e) |
| T3 | Remove duplicate launcher/src, repo hygiene | DONE (1223b54) |
| T4 | Secure /api/server/start, admin authz, HWID bans | DONE (658fa72) |
| T5 | TokenStorage Windows DPAPI | DONE (77b5483) |
| T6 | Java 24 detection | DONE (2801a18) |
| T7 | launchGame uuid + accessToken | DONE (d32f92a) |
| T8 | HWIDExtractor verification test | PENDING |
| T9 | .gitignore hygiene (instances, build) | DONE (1223b54) |
| T10 | docs/ERROR_LOG.md + WORKFLOW_RULES committed | PENDING |
| T11 | Velocity plugin skeleton | PENDING (DECISION-004) |
| T12 | Backend test coverage (config/sessions/devices) | PENDING |
| T13 | BackendAuthClient contract test | PENDING |
| T14 | CI workflow (GitHub Actions) | PENDING |
| T15 | config/.env.example documentation | DONE (658fa72) |

## Extended plan (T16–T45)

| ID | Title | Notes | Status |
|----|-------|-------|--------|
| T16 | Backend security review document | rate limiting, input validation, threat notes | PENDING |
| T17 | Launcher settings storage (JSON) | project-owned config layer | PENDING |
| T18 | Launcher settings tests | round-trip persistence | PENDING |
| T19 | Update manifest model + parser | docs/UPDATE_SPEC.md alignment, stub | PENDING |
| T20 | Update manifest tests | parser unit tests | PENDING |
| T21 | Skin system model (skin id, hash, URL) | docs/SKIN_SYSTEM.md stub | PENDING |
| T22 | Backend: account profile endpoint | GET /api/v1/account/me | PENDING |
| T23 | Backend: password change endpoint | POST /api/v1/account/password | PENDING |
| T24 | Backend: session list endpoint | GET /api/v1/account/sessions | PENDING |
| T25 | Backend: session revoke endpoint | POST /api/v1/account/sessions/revoke | PENDING |
| T26 | Backend tests for profile/password endpoints | | PENDING |
| T27 | Backend tests for session management endpoints | | PENDING |
| T28 | Rate limiter unit tests | window, limit, reset | PENDING |
| T29 | Config loader unit tests | env parsing + validation | PENDING |
| T30 | Device service unit tests | register/upsert semantics | PENDING |
| T31 | Session service unit tests | create/refresh/revoke lifecycle | PENDING |
| T32 | Server graceful shutdown hardening | timeouts, error paths | PENDING |
| T33 | Structured request logging middleware | method, path, status, latency | PENDING |
| T34 | Panic-recovery middleware | return 500, log stack | PENDING |
| T35 | Security headers middleware | nosniff, no-store, frame deny | PENDING |
| T36 | Launcher HTTP client timeouts + TLS pinning notes | QNetworkAccessManager hardening | PENDING |
| T37 | Launcher error mapping for backend status codes | 401/403/409/429 UX messages | PENDING |
| T38 | Instance list CLI filter/search | main.cpp option | PENDING |
| T39 | Launcher version + build info output | --version | PENDING |
| T40 | Backend health/readiness split | /healthz + /readyz | PENDING |
| T41 | Backend OpenAPI-ish endpoint doc | docs/API.md | PENDING |
| T42 | CI: launcher build matrix on windows-latest | GitHub Actions | PENDING |
| T43 | CI: go vet + test on ubuntu-latest | GitHub Actions | PENDING |
| T44 | Documentation sync pass | AI_CONTEXT/WORK_STATE/PROGRESS/CHANGELOG | PENDING |
| T45 | Final verification, merge, push, report | full build + go test | PENDING |

## Non-negotiables

- Server-side authorization is authoritative; the client is never the final
  security authority.
- No secrets in source; configuration is environment-driven.
- No premature distributed systems, queues, Redis or heavyweight frameworks.
- Every task updates `docs/PROGRESS.md`, `docs/WORK_STATE.md`,
  `docs/CHANGELOG.md` (when appropriate) and `docs/AI_CONTEXT.md` (when project
  state changes).
