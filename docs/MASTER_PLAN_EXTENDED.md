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
| T8 | HWIDExtractor verification test | DONE (f57327b) |
| T9 | .gitignore hygiene (instances, build) | DONE (1223b54) |
| T10 | docs/ERROR_LOG.md + WORKFLOW_RULES committed | DONE (39e4b97) |
| T11 | Velocity plugin skeleton | PENDING (DECISION-004) |
| T12 | Backend test coverage (config/sessions/devices) | DONE (2758d90, cf334b6, b314647) |
| T13 | BackendAuthClient contract test | DONE (079939f) |
| T14 | CI workflow (GitHub Actions) | DONE (e7190de) |
| T15 | config/.env.example documentation | DONE (658fa72) |

## Extended plan (T16–T45)

| ID | Title | Notes | Status |
|----|-------|-------|--------|
| T16 | Backend security review document | rate limiting, input validation, threat notes | DONE (fdd7a4a) |
| T17 | Launcher settings storage (JSON) | project-owned config layer | DONE (pre-existing InstancePersistence) |
| T18 | Launcher settings tests | round-trip persistence | DONE (31c3372) |
| T19 | Update manifest model + parser | docs/UPDATE_SPEC.md alignment, stub | DONE (f81ca5d) |
| T20 | Update manifest tests | parser unit tests | DONE (f81ca5d) |
| T21 | Skin system model (skin id, hash, URL) | docs/SKIN_SYSTEM.md stub | DONE (fd520e0) |
| T22 | Backend: account profile endpoint | GET /api/v1/account/me | DONE (56bd666) |
| T23 | Backend: password change endpoint | POST /api/v1/account/password | DONE (b984470) |
| T24 | Backend: session list endpoint | GET /api/v1/account/sessions | DONE (b984470) |
| T25 | Backend: session revoke endpoint | POST /api/v1/account/sessions/revoke | DONE (b984470) |
| T26 | Backend tests for profile/password endpoints | | DONE (b984470) |
| T27 | Backend tests for session management endpoints | | DONE (b984470) |
| T28 | Rate limiter unit tests | window, limit, reset | DONE (0820f80) |
| T29 | Config loader unit tests | env parsing + validation | DONE (2758d90) |
| T30 | Device service unit tests | register/upsert semantics | DONE (b314647) |
| T31 | Session service unit tests | create/refresh/revoke lifecycle | DONE (cf334b6) |
| T32 | Server graceful shutdown hardening | timeouts, error paths | DONE (1c20557) |
| T33 | Structured request logging middleware | method, path, status, latency | DONE (cf15ea5) |
| T34 | Panic-recovery middleware | return 500, log stack | DONE (cf15ea5) |
| T35 | Security headers middleware | nosniff, no-store, frame deny | DONE (cf15ea5) |
| T36 | Launcher HTTP client timeouts + TLS pinning notes | QNetworkAccessManager hardening | DONE (10c5eed) |
| T37 | Launcher error mapping for backend status codes | 401/403/409/429 UX messages | DONE (2645e09) |
| T38 | Instance list CLI filter/search | main.cpp option | DONE (e220547) |
| T39 | Launcher version + build info output | --version | DONE (e220547) |
| T40 | Backend health/readiness split | /healthz + /readyz | DONE (7ae316d) |
| T41 | Backend OpenAPI-ish endpoint doc | docs/API.md | DONE (ad2b67c) |
| T42 | CI: launcher build matrix on windows-latest | GitHub Actions | DONE (e7190de) |
| T43 | CI: go vet + test on ubuntu-latest | GitHub Actions | DONE (e7190de) |
| T44 | Documentation sync pass | AI_CONTEXT/WORK_STATE/PROGRESS/CHANGELOG | DONE (014f117) |
| T45 | Final verification, merge, push, report | full build + go test | DONE (014f117) |

## Non-negotiables

- Server-side authorization is authoritative; the client is never the final
  security authority.
- No secrets in source; configuration is environment-driven.
- No premature distributed systems, queues, Redis or heavyweight frameworks.
- Every task updates `docs/PROGRESS.md`, `docs/WORK_STATE.md`,
  `docs/CHANGELOG.md` (when appropriate) and `docs/AI_CONTEXT.md` (when project
  state changes).
