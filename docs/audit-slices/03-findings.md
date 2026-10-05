# SLICE 3 — FINDINGS (Batch 8 audit)

Audit scope: real attack surface of this repo (Go backend auth/skins/queue/admin,
C++/Qt launcher, Velocity plugin, CI). Statuses: PASS / FINDING / N/A with
evidence. Findings are recorded here; remediation happens in SLICE 4.

## 3a — Dependencies, secrets, CI config

### Dependencies (Go)
- `go vet ./...` — PASS (clean).
- `go test -race` — NOT RUN (environment): no C compiler installed (`gcc`
  missing; MSVC `cl` rejects Go `-race` flags). Documented env limitation, not
  a code finding.
- `govulncheck` — NOT INSTALLED; cannot scan advisories locally. Manual review
  of `go list -m all` shows all modules from official upstreams
  (modernc.org, golang.org/x, github.com/google, dustin, mattn) — no
  hallucinated or suspicious package.
- N/A: no C++ third-party dependency via FetchContent/ExternalProject; CMake
  uses only system Qt6 (Core/Network/Widgets). Velocity uses only
  `velocity-api` + mavenCentral (no typo-squat risk confirmed).

### Secrets
- Grep for `Bearer <long>`, `api_key=`, `sk-`, `secret=`, plus
  `supersecret/secret-admin/e2e-admin/proxy-admin/change-me`:
  - All matches are **test fixtures** (inputs to `newAdminTestServer`,
    `register`, `scripts/e2e.go`). No production credential, API key, or
    real bearer token in committed source.
  - `velocity/src/main/resources/velocity.toml.example` contains a
    placeholder `change-me` — expected (example file, not a live secret).
  - PASS: no hardcoded production secret in source, config, git history
    (recent history reviewed), CI, or build output.
- Config values are read from env (`AUTH_ADMIN_TOKEN`, etc.) — PASS.

### CI config
- `backend.yml` / `launcher.yml`:
  - Unpinned third-party actions (`actions/checkout@v4`, `setup-go@v5`,
    `jurplel/install-qt-action@v4`) — FINDING (Low): use pinned SHA or
    release-pin for supply-chain hygiene.
  - Default `GITHUB_TOKEN` permissions not restricted — FINDING (Medium):
    add `permissions: { contents: read }` to both jobs (least privilege).
  - Triggers are branch/path scoped to `master`; no untrusted PR-driven shell
    interpolation of `${{ }}` from attacker input — PASS.
- FIXED this slice: added `permissions: { contents: read }` to both workflows.

### Recorded findings (3a)
| ID | Source | Severity | Status |
|----|--------|----------|--------|
| CI-1 | launcher/backend workflow | Medium | FINDING → FIXED (contents:read) |
| CI-2 | unpinned actions | Low | FINDING (pending: pin SHAs in SLICE 4) |
| SEC-1 | none | — | PASS (no hardcoded prod secret) |
| DEP-1 | go.mod | — | PASS (no suspicious dep; govulncheck unavailable) |

(3b and 3c sections appended as completed.)

## 3b — Go backend review

Verified by code inspection:

- **SQL injection** — PASS: all DB queries use `?` placeholders (the only
  `fmt.Sprintf` is for the DSN connection string, not a query). No SQLi.
- **Command injection** — PASS: `handleStartServer` runs `exec.CommandContext`
  with a config-fixed script path + `"python"` argv, no shell, no user input in
  the command. Output never echoed.
- **Timing-based user enumeration** — PASS: `handleLogin` runs a dummy bcrypt
  compare (`dummyPasswordHash`) when the account is not found.
- **Admin token compare** — PASS: `subtle.ConstantTimeCompare` in
  `adminAuthorized`.
- **Proxy auth endpoint** — PASS: `handleProxyAuth` is admin-token gated, returns
  only `ok`/`denied`, never account data; invalid token → `denied`.
- **Rate limiting** — PASS: per-IP token bucket on auth/account/admin/queue
  routes; skin upload has per-user `AllowSkinUpload` (2/3h).
- **Request size limits** — PASS: `MaxBytesReader` on handlers.go, skins.go,
  queue_handlers.go.
- **Queue boundedness** — PASS: `chan SkinJob, 64` + `workerCount=4`; enqueue
  rejects when full (429/503).
- **CORS / CSRF** — N/A (no CORS headers → same-origin only; API is bearer-token
  JSON, no cookies/session fixation).
- **IDOR** — PASS: session revoke/logout-others verify `AccountID` ownership;
  skin delete is account-scoped; admin routes require the admin token.
- **Refresh token rotation** — FINDING (Medium): `Refresh` reuses the same
  refresh token instead of rotating on use. Recorded for SLICE 4.
- **Path traversal** — PASS for DB layer; skin data stored as BLOB (no filesystem
  path from user input). (Launcher file handling reviewed in 3c.)

### Recorded findings (3b)
| ID | Source | Severity | Status |
|----|--------|----------|--------|
| AUTH-1 | sessions/service.go Refresh | Medium | FINDING (rotation) |
| SQLi/CMDi/IDOR/etc | — | — | PASS (see above) |

## 3c — Launcher + Velocity + AI/agent surface

### C++ launcher
- **Ownership / double-free** — PASS: `LaunchTask` uses `std::vector<LaunchStep*>`
  with QObject parenting (single owner); no raw owning `new/delete` pair beyond
  QObject children (`new QProcess(this)`, `new QVBoxLayout`, etc., all parented).
  DownloadAsset/Queue use `unique_ptr`/parented children. No double-free.
- **Token on command line** — PASS: `GameLauncher` passes the session token via
  env (`MCSERVER_SESSION_TOKEN`), never in `preparedArguments`. Tested.
- **Java executable path** — PASS: `JavaDetector::isPlausibleJava` rejects
  arbitrary executables (only `bin/java(.exe)`, executable, non-dir). Mitigates
  `LAUNCHER_JAVA_HOME`/`JAVA_HOME` hijack.
- **Path traversal (classpath/natives)** — PASS: `GamePaths` canonicalizes and
  rejects `..`/absolute escapes; classpath only accepts `.jar` inside libs dir.
- **Asset download integrity** — PASS: `DownloadAsset` verifies SHA-256 before
  accepting; removes file on mismatch; canonical destination.
- **Archive/native extraction** — NOT YET PRESENT (no zip extraction module
  exists). N/A now; a future modpack extractor must re-audit for zip-slip.

### Velocity (Java)
- **Trust boundary** — PASS: proxy is enforcement point; backend authoritative;
  `PostLoginEvent` validates token and disconnects on deny.
- **Backend client timeouts** — PASS: 3s connect / 5s request / 5s
  `completeOnTimeout`; token in body, not query.
- **TLS** — PASS for design (no `setSSLParameters` disabling verification); the
  example config defaults to `http://` — FINDING (Low): production should use
  `https://` and TLS validation.
- **Skin metadata untrusted** — PASS: `SkinInjection` places backend texture/
  signature verbatim into a GameProfile Property; no parsing → no code exec.
- **Config** — PASS: `PluginConfig` rejects non-http(s) URLs and empty token.

### AI/agent specific
- No LLM/agent runtime, prompt-injection surface, MCP tool, or RAG exists in the
  shipped product. The launcher is a client; the backend is a normal API. Prompt
  injection / agent memory poisoning / tool poisoning: **N/A** (no such runtime).
  Untrusted data (skin bytes, modpack metadata, backend responses) is handled as
  data, never as instructions.

### Recorded findings (3c)
| ID | Source | Severity | Status |
|----|--------|----------|--------|
| CI-2 | workflows | Low | FINDING → FIXED (SHA-pinned actions) |
| VEL-1 | velocity.toml.example | Low | FINDING → FIXED (https+TLS guidance) |
| AUTH-1 | sessions/service.go | Medium | FINDING → FIXED (refresh rotation + test) |
| CXX-1 | none | — | PASS (ownership/token/path/integrity) |
| AI-1 | none | — | N/A (no AI runtime in product) |
