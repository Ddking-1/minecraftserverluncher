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
- [ ] Gradle wrapper environment and Velocity plugin structure (DONE — `ae6d4c5`)
- [ ] Main class, event listeners, backend auth client (bearer validation)
- [ ] Skin injection logic (inject GameProfile texture from the Go DB)

## BATCH 8 — CI/CD Cross-Compilation & End-to-End Suite
- [ ] Go & C++ Windows/Linux release build pipeline hardening
- [ ] **Local E2E test harness (no Docker, no GPU, <=6GB RAM)**: isolated
      `./tmp_test_env` on port `:8089` running the Go binary; full lifecycle
      (register → login → forgot/reset → async skin upload+poll → admin
      list/ban/delete); auto-fix loop via lmstudio5 until 100% green; cleanup
      of processes and temp dir. (`scripts/e2e_test.go` or `server/integration_test.go`)
- [x] **Auto-runner two-step stop test**: `scripts/auto_runner.ps1` created;
      `scripts/test_runner_check.tmp` created, verified, then deleted on manual
      "continue". (DONE — `scripts/auto_runner.ps1`)
