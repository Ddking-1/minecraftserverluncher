# TASKS

Authoritative task list for the project. Mirrors the live TODO widget in the
editor. Micro-delegation protocol: small isolated chunks go to the local
worker (`lmstudio5`, Qwen2.5-Coder-7B) via the LM Studio REST API; the
Architect integrates, tests, and commits. If the worker fails to reach green
after 5 attempts, the Architect fixes the specific piece (Smart Fallback).

## BATCH 4 — Web Dashboard & Admin Backend
- [ ] Admin API: user list, search, pagination, ban/unban, account deletion
      (`server/api/admin.go`, `admin_test.go`)
- [ ] User Session Manager API: active session list, revoke single session,
      "log out from all other devices"
- [ ] Web Dashboard Front-End: HTML/CSS/JS integration with the Go backend
- [ ] Forgot-password reset endpoints (DONE — `d265d12`)

## BATCH 5 — Server Resiliency, Rate Limiter & Async Job Queue
- [ ] Skin processing worker pool: buffered channel (`chan SkinJob`), 1–4 threads
- [ ] Token bucket rate limiter: 1 second / 1 skin processing limit
- [ ] Async API: `POST /api/v1/skins/upload` -> `202 Accepted` + `task_id`
- [ ] Polling endpoint: `GET /api/v1/tasks/{id}`

## BATCH 6 — Launcher C++ / Qt6 UI & Modpack Engine
- [ ] Qt widgets wired to `EditorConfig`, `SkinClient`, `AuthClient`
- [ ] Async asset & modpack downloader (SHA-256 verify, progress)
- [ ] Game launcher engine (Java detection, classpath, RAM, start)

## BATCH 7 — Velocity Auth Proxy Plugin (Java / Gradle)
- [x] Gradle wrapper environment and Velocity plugin structure (DONE — `ae6d4c5`)
- [x] Main class, event listeners, backend auth client (bearer validation)
      (DONE — `275b42b` proxy-auth endpoint, `cc2a72d` plugin)
- [x] Skin injection logic (inject GameProfile texture from the Go DB)
      (DONE — `cc2a72d` `SkinInjection`)
- [x] **B7.10 mock E2E**: started the Go backend on `:8089`, exercised
      `/api/v1/proxy/auth` the way the Java client calls it — valid token →
      `{"status":"ok"}`, invalid → `{"status":"denied"}`, wrong admin token →
      401. (DONE — manual mock verification)

## BATCH 8 — CI/CD Cross-Compilation & End-to-End Suite
- [x] Go & C++ Windows/Linux release build pipeline hardening
      (DONE — Windows launcher `windows-2022`, least-privilege CI)
- [x] **Local E2E test harness (no Docker, no GPU, <=6GB RAM)**: isolated
      `./tmp_test_env` on port `:8089` running the Go binary; full lifecycle
      (register → login → forgot/reset → async skin upload+poll → admin
      list/ban/delete); auto-fix loop via lmstudio5 until 100% green; cleanup
      of processes and temp dir. (`scripts/e2e.go` — DONE)
- [x] **Auto-runner two-step stop test**: `scripts/auto_runner.ps1` created;
      `scripts/test_runner_check.tmp` created, verified, then deleted on manual
      "continue". (DONE — `scripts/auto_runner.ps1`)
- [x] **Batch 8 GenAI/Vibe-coding security audit** (DONE — `11e07a0`,
      `7e28a1c`, `040ecc1`, `6445375`): findings + remediation (refresh-token
      rotation AUTH-1, CI least-privilege CI-1, SHA-pinned actions CI-2,
      https guidance VEL-1) + final matrix
      (`docs/audit-slices/05-final-verification-matrix.md`)

## BATCH 9 — Packaging, Enterprise Hardening & Auto-Update
- [x] **B9.1 NSIS installer**: NSIS v3.12 installed via winget, `makensis`
      verified, Qt DLLs + qwindows plugin staged,
      `packaging/nsis/MinecraftServerLauncher.nsi` built to
      `MinecraftServerLauncher-0.1.0-setup.exe` (6.8 MB). (DONE — `652ffec`)
- [ ] B9.2 CMake staging target (copy Qt DLLs + platforms next to exe)
- [ ] B9.3 C++ UpdateManager (manifest fetch + SHA-256)
- [ ] B9.4 Go backend `GET /api/v1/updates/manifest` + tests
