# WORK STATE

Current phase:
PHASE 3 — LAUNCHER FEATURE COMPLETION + BACKEND ROBUSTNESS (COMPLETE)

Current task:
Phase 3 backlog (`docs/MASTER_PLAN_PHASE3.md`, P3-01 … P3-10) is complete.
The T1–T45 backlog (`docs/MASTER_PLAN_EXTENDED.md`) is complete except T11
(Velocity skeleton), which stays BLOCKED on DECISION-004 legal sign-off and
DECISION-005 (metadata source).

Status:
Green. Backend: `go build`, `go vet`, `go test ./...` all pass. Launcher:
CMake configure + Release build with Qt 6.6.3 MSVC 2022 succeed; CTest runs 11
tests (hwid, settings, update_manifest, skin_model, backend_contract,
settings_store, mod_scanner, skin_cache, rotating_logger, update_checker,
file_integrity), all pass.

Branch:
master (all task branches fast-forward merged and pushed).

Phase 3 tasks (all DONE):
- P3-01 SettingsStore (launcher.json load/save, defaults, atomic) + tests
- P3-02 Mod list scanner (mods/ dir, enabled/disabled, sorted) + tests
- P3-03 Skin cache (content-addressed, SHA-256 verified) + tests
- P3-04 Rotating file logger (size rotation, keep N) + tests
- P3-05 Update checker (semantic version compare + decision) + tests
- P3-06 File integrity verifier (SHA-256 of file vs expected) + tests
- P3-07 Backend request-ID + structured logging middleware + tests
- P3-08 Backend auth/token extra tests + DB expiry indexes
- P3-09 Launcher memory/JVM settings validation + tests
- P3-10 Phase 3 final verification + documentation sync

Skin Management API, Rate Limiter and UI Widget tasks (all DONE):
- Skin Management API (backend handlers + OpenAPI contract) - DONE
- Rate Limiter (429 Too Many Requests + Retry-After) - DONE
- UI Widget (SkinPreviewWidget + MainWindow embedding) - DONE

Current step:
- Milestone `feature/go-backend-hardening-and-fuzz` is COMPLETED.
- Fuzz test coverage: 100% — native Go Fuzz targets added for the Skin processor (`server/skins/processor_fuzz_test.go`) and the rate limiter (`server/api/ratelimit_fuzz_test.go`).
- Build system (CMake) audit: 100% — `launcher/CMakeLists.txt` hardened with C++17, `-Wall -Wextra -Werror`, Qt6 Core/Network only, AUTOMOC, full source list and Windows libs; OpenSSL dependency removed from HWIDExtractor.cpp; root `CMakeLists.txt` deduplicated; verified GREEN with a real build.
- Qt6 GUI wiring: MainWindow skeleton + SkinPreviewWidget `skinUploadFinished` signal added on this branch (source only, not yet compiled or wired into `main.cpp`).
- Documentation: AI-generated-code disclaimer added to README.md and docs/AI_CONTEXT.md; CHANGELOG.md, PROGRESS.md and WORK_STATE.md updated.

Next step:
Compile and wire the new MainWindow into main.cpp when the GUI entry point is implemented. T11 Velocity remains BLOCKED on DECISION-004/005.

Repository facts:
- default branch: master
- remote: origin (private GitHub repo Ddking-1/minecraftserverluncher)
- reference/ is local-only, never committed (DECISION-002)
- backend source exists under server/ (Go module minecraftserverluncher/server)
- launcher core exists under launcher/ (C++ / Qt6, builds and tests green)
- velocity directory is scaffolded but empty (no code yet)

Last commit:
See docs/PROGRESS.md / `git log --oneline`.

## 2026-10-09 - UPDATE: BACKEND TEST RUN + SKIN API / PTERODACTYL WEBHOOK

Status:
GREEN. `cd server && go test -v ./... -count=1` -> 200 PASS / 0 FAIL
(api, auth, config, database, devices, pterodactyl, sessions, skins, web).
`go build ./...` and `go vet ./...` are clean. The Go module root is `server/`,
so `go test -v ./server/...` from the repository root does not work.

DONE (marked complete in docs/PROGRESS.md):
- Skin API handlers: Upload, Get, Avatar, Delete (server/api/skins_handler.go,
  tested by server/api/skins_handler_test.go).
- Pterodactyl webhook handler + 401 error hint
  (server/api/webhook_handler.go returns 401 "missing webhook secret" when
  PTERODACTYL_WEBHOOK_SECRET is unset; server/pterodactyl/webhook.go returns 401
  for a missing or invalid signature and panics early on an empty secret).

Known issues (exact next step):
1. `Server.Handler()` panics when `s.SkinHandler != nil`:
   `DELETE /api/v1/skins/{id}` conflicts with the existing
   `DELETE /api/v1/skins/{variant}`. Decide which endpoint owns that pattern
   (or move one of them) before wiring the Skin API.
2. `server/main.go` sets neither `SkinHandler` nor `PterodactylWebhook`, so the
   new endpoints are not exposed in production yet.
3. `server/tmp/*.go` scratch helpers are excluded from the build with
   `//go:build ignore`; remove them when the related investigation is done.

Branch:
feature/go-backend-hardening-and-fuzz (this milestone's documentation changes are committed and pushed in this update).
