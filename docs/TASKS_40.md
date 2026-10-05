# TASKS_40 — Real roadmap

Ordered, small, actionable units. Batches stop for owner review with `Folytasd`.
Roughly 40 items total, grouped into phases. Each item names real paths.

## Phase 0 — Decisions & docs (Batch 1)
- T1 Update `docs/DECISIONS.md`: DECISION-004 final (GPL launcher publication),
  DECISION-005 decided (backend DB as metadata source). [DONE]
- T2 Regenerate this `docs/TASKS_40.md` grounded in the real codebase. [DONE]

## Phase 1 — Skin system (backend) (Batch 1–2)
- T3 `server/database/skins.go`: `skins` table (id, account_id, variant, data BLOB,
  content_type, created_at, updated_at; UNIQUE(account_id, variant)).
- T4 DB helpers: `UpsertSkin`, `SkinByAccountAndVariant`, `ListSkins`, `DeleteSkin`.
- T5 `server/api/skins.go`: endpoints behind bearer auth —
  `PUT /api/v1/skins/{variant}` (upload), `GET /api/v1/skins` (list own),
  `GET /api/v1/skins/{account}/{variant}` (public fetch), `DELETE /api/v1/skins/{variant}`.
- T6 Validate skin upload: size cap, content-type allowlist, account ownership.
- T7 `server/api/skins_test.go` and `server/database/database_test.go` additions:
  upload/get/list/delete lifecycle and ownership (403) cases.
- T8 Verify `go test ./...` green; commit/merge/push skin system.

## Phase 2 — Launcher splash + local editor (Batch 3)
- T9 `launcher/gui/SplashScreen.{h,cpp}`: QWidget showing a bundled image instantly.
- T10 Fake-render splash: block nothing, just show image while rest loads.
- T11 `launcher/CMakeLists.txt`: add splash sources + a test.
- T12 Local editor config mode: `launcher/core/EditorConfig.{h,cpp}` — pure-local
  UI layout config (no server calls), load/save JSON.
- T13 Editor end-to-end: read/write local editor config file; unit test.

## Phase 2b — Skin security & limits (Extended Batch)
- T-A (Backend) Skin upload per-user rate limit: max 2 uploads per 3 hours,
  resets after the window. Returns 429 when exceeded.
- T-B (Backend) Strict skin format validation: verify image dimensions are a
  Minecraft-valid skin size (64x64 or 64x32).
- T-C (Security) Payload inspection: strict magic-byte MIME sniffing so a
  PHP/EXE cannot be uploaded with a .png extension. Applied client- and
  server-side expectations.
- T-D (Launcher C++) Skin client UI: upload + display logic in the launcher.
- T-E (Web frontend) Skin manager UI on the website (login + upload).

## Phase 3 — Website (Batch 4)
- T14 `web/` Go http.Handler (embed) + template index.
- T15 Login form (HTML/CSS via local worker).
- T16 Skin manager page (upload/list) talking to the backend API.
- T17 Serve static assets; CSP headers.
- T18 `server` integration test for web handler mount.

## Phase 4 — Velocity plugin (Batch 5, after DECISION-004/005)
- T19 `velocity/plugin/` build scaffolding (Gradle + Velocity API dep).
- T20 Velocity plugin main class + lifecycle.
- T21 Proxy→backend auth validation endpoint contract.
- T22 Backend endpoint for proxy token validation (no client trust).
- T23 Wire login flow through backend; reject unauthenticated joins.
- T24 Plugin + backend integration test.

## Phase 5 — Backend hardening & tests (Batch 6+)
- T25 `server/api`: enforce device ownership on session create/revoke.
- T26 Refresh token rotation (issue new refresh on refresh).
- T27 Login: reject before bcrypt when HWID banned; avoid enumeration.
- T28 `server/auth`: password strength checks (length + basic classes).
- T29 `server/config`: validate TTLs, ports, bcrypt cost range.
- T30 `server/api`: limit refresh/logout by user, not just IP.
- T31 `server/database`: uniqueness index tests for devices/skins.
- T32 `server/api`: middleware test for request-body size cap.

## Phase 6 — Launcher core & tests (Batch 7+)
- T33 `launcher/core`: unit tests for `InstancePersistence` round-trip.
- T34 `launcher/core`: `RotatingFileLogger` retention-cap edge test.
- T35 `launcher/updates`: `UpdateChecker` incompatible-decision test.
- T36 `launcher/core`: `SettingsStore` JVM-arg injection rejection test.
- T37 `launcher/game`: `JavaDetector` test for plausible Java dirs.
- T38 `launcher/auth`: `BackendAuthClient` non-regression (no new deps).

## Phase 7 — CI & docs (Batch 8)
- T39 CI: add Go `-race` for backend; keep launcher on `windows-2022`.
- T40 Docs: refresh `docs/AI_CONTEXT.md`, `docs/WORK_STATE.md`, `docs/PROGRESS.md`
  to reflect phases 0–4 completion.
