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
