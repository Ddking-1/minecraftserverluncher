# TASKS

Authoritative task list for the project. Mirrors the live TODO widget in the
editor. Micro-delegation protocol: small isolated chunks go to the local
worker (`lmstudio5`, Qwen2.5-Coder-7B) via the LM Studio REST API; the
Architect integrates, tests, and commits. If the worker fails to reach green
after 5 attempts, the Architect fixes the specific piece (Smart Fallback).

## BATCH 4 — Web Dashboard & Admin Backend (DONE)
- [x] Admin API: user list, search, pagination, account deletion
      (`server/api/admin_users.go` + `admin_users_test.go` — `55e1ff7`)
- [x] User Session Manager API: list, revoke single session,
      "log out from all other devices" (`b91ad44`)
- [x] Web Dashboard Front-End: HTML/CSS/JS served by the Go backend
      (`d091cb0`, `03ca9ed`)
- [x] Forgot-password reset endpoints (`d265d12`)

## BATCH 5 — Server Resiliency, Rate Limiter & Async Job Queue (DONE)
- [x] Skin processing worker pool: buffered channel (`chan SkinJob`), 1–4 threads
      (`48a9122`)
- [x] Token bucket rate limiter (per-IP) + per-user skin upload limit (`6c70007`)
- [x] Async API: `POST /api/v1/skins/upload` -> `202 Accepted` + `task_id` (`48a9122`)
- [x] Polling endpoint: `GET /api/v1/tasks/{id}` (`48a9122`)

## BATCH 6 — Launcher C++ / Qt6 UI & Modpack Engine (DONE)
- [x] Qt widgets wired to `EditorConfig` / `SkinClient` / `AuthClient`
      (`658757a` EditorSettingsWidget, `26a818a` SkinUploadWidget, `c444ab4` LoginDialog)
- [x] Async asset & modpack downloader (SHA-256 verify, worker pool, retry)
      (`f1a59db` DownloadAsset, `7fb6887` DownloadQueue, `ff93276` retry)
- [x] Game launcher engine: Java detection, classpath/natives, RAM, process start
      (`51d4173`, `c876bec`, `d7454c4`)

## BATCH 7 — Velocity Auth Proxy Plugin (Java / Gradle) (DONE)
- [x] Gradle wrapper + plugin structure (`ae6d4c5`)
- [x] Main class, event listeners, backend auth client (`275b42b`, `cc2a72d`)
- [x] Skin injection logic (`cc2a72d` `SkinInjection`)
- [x] B7.10 mock E2E (proxy/auth ok/denied/401 verified)

## BATCH 8 — CI/CD & Security Audit (DONE)
- [x] Go & C++ release pipeline hardening (Windows `windows-2022`, least-privilege CI)
- [x] Local E2E harness (`scripts/e2e.go`)
- [x] Auto-runner two-step stop test
- [x] GenAI/Vibe-coding security audit + remediation
      (`11e07a0`, `7e28a1c`, `040ecc1`, `6445375`)

## BATCH 9 — Packaging, Enterprise Hardening & Auto-Update (DONE)
- [x] B9.1 NSIS installer (makensis v3.12, setup.exe 6.8MB) (`652ffec`)
- [x] B9.2 CMake Qt runtime staging target (`8ff3d0f`)
- [x] B9.3 C++ UpdateManager (manifest fetch) + CTest (`7e06826`)
- [x] B9.4 Go backend `GET /api/v1/updates/manifest` + tests (`18ca45c`)

## BATCH 10 — Release Security
- [ ] Update manifest checksum/signature verification
- [ ] Rate-limit protection on `GET /api/v1/updates/manifest`

## BATCH 11 — Auto-Update Wire-In
- [ ] Wire `UpdateManager` into launcher startup (compare local vs manifest)
- [ ] Download update with SHA-256 verify + prepare swap process

## BATCH 12 — Final E2E & Documentation
- [ ] Extend `scripts/e2e.go` with update-endpoint checks
- [ ] Update release documentation; mark v2.0 milestone complete
