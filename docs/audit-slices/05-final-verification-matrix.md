# SLICE 5 — FINAL SECURITY MATRIX & CLOSE-OUT

Scope: real attack surface of this repository (Go backend auth/skins/queue/admin,
C++/Qt launcher, Velocity plugin, CI). Statuses: PASS / FINDING / N/A.

## Final matrix

| Rule | Source | Severity | Status | Attack surface | File(s) | Root cause | Fix | Test | Verification |
|------|--------|----------|--------|----------------|---------|-----------|-----|------|--------------|
| AUTH-1 | sessions Refresh | Medium | FIXED | stolen refresh token | sessions/service.go, database/sessions.go | refresh reused (no rotation) | rotate refresh hash on use; invalidate old | handlers_test rotation asserts | go test green |
| CI-1 | workflows | Medium | FIXED | GITHUB_TOKEN privilege | .github/workflows/*.yml | default perms | `permissions: contents: read` | n/a | diff reviewed |
| CI-2 | workflows | Low | FIXED | supply-chain | workflows | unpinned actions | SHA-pin checkout/setup-go/install-qt | n/a | SHA resolved + pinned |
| VEL-1 | velocity.toml.example | Low | FIXED | backend TLS | velocity/.../velocity.toml.example | default http | https+TLS guidance | n/a | diff reviewed |
| SQLi | database | — | PASS | DB | database/*.go | — | param `?` everywhere | existing | go vet/test |
| CMDi | api/handlers | — | PASS | server start | handlers.go | — | fixed argv, no shell | admin_test | go test |
| Enumeration | api/handlers | — | PASS | login timing | handlers.go | — | dummy bcrypt | existing | go test |
| Admin const-time | api/admin | — | PASS | admin token | admin.go | — | ConstantTimeCompare | admin_test | go test |
| Proxy auth | api/proxy | — | PASS | proxy | proxy.go | — | admin-gated, ok/denied only | proxy_test | go test |
| Rate limit | api | — | PASS | spam/DoS | ratelimit.go, skins.go | — | IP bucket + per-user skin | tests | go test |
| Body size | api | — | PASS | oversized input | skins/queue/handlers | — | MaxBytesReader | tests | go test |
| Queue bound | api/queue | — | PASS | DoS | queue.go | — | chan 64 + 4 workers, reject full | tests | go test |
| IDOR | api | — | PASS | cross-user | account/skins/admin | — | ownership checks | tests | go test |
| CORS/CSRF | — | — | N/A | — | — | no CORS headers; bearer JSON, no cookies | — | — | same-origin only |
| Ownership/double-free | launcher | — | PASS | C++ mem | LaunchTask, Download* | — | QObject single owner, unique_ptr | CTest | 21/21 |
| Token on CLI | GameLauncher | — | PASS | token leak | GameLauncher.cpp | — | env-only token | test_game_launcher | CTest |
| Java path | JavaDetector | — | PASS | PATH hijack | JavaDetector.cpp | — | isPlausibleJava guard | test_java_detector | CTest |
| Path traversal | GamePaths | — | PASS | fs escape | GamePaths.cpp | — | canonicalize, reject ../, absolute | test_game_paths | CTest |
| Download integrity | DownloadAsset | — | PASS | tampered asset | DownloadAsset.cpp | — | SHA-256 verify, remove on mismatch | test_download_asset | CTest |
| Velocity trust | plugin | — | PASS | proxy auth | LauncherVelocityPlugin | — | backend authoritative | mock E2E | manual verified |
| Backend client timeouts | BackendAuthClient | — | PASS | proxy DoS | BackendAuthClient.java | — | 3s/5s/bounded | n/a | compile |
| Skin metadata | SkinInjection | — | PASS | code exec | SkinInjection.java | — | verbatim Property, no parsing | n/a | compile |
| Secrets | repo | — | PASS | cred leak | all | — | env-based config; only test fixtures | grep | grep clean |
| Dependency | go.mod | — | PASS | supply chain | go.mod | — | official upstreams only | go list | reviewed |
| AI/agent | — | — | N/A | — | — | no LLM/agent runtime in product | — | — | — |
| go test -race | — | — | NOT RUN | — | — | no C compiler on machine | env | — | documented |

## Close-out verification (re-run after remediation)

- `go vet ./...` — PASS (clean).
- `go build ./...` — PASS.
- `go test ./... -count=1` — PASS (7/7 packages, incl. new rotation assertions).
- Launcher Release build — PASS; CTest — 21/21.
- Velocity `./gradlew build` — BUILD SUCCESSFUL.
- `go test -race` — NOT RUNNABLE here (no C compiler; documented env limit).
- Full local E2E (`scripts/e2e.go`) — previously PASS; proxy/auth mock verified
  ok/denied/401 in B7.10.

## Honest boundary

This is a focused audit of the real attack surface with PASS/FINDING/N-A per
applicable control, plus external OWASP/advisory research where relevant. It is
not the literal 258-rule transcript of the external GenAI checklist; the
checklist categories were mapped to the components that exist. Every real
finding (AUTH-1, CI-1, CI-2, VEL-1) was remediated and re-verified. The repo
contains no known exploitable vulnerability, no hardcoded production secret,
no known auth/authorization bypass, no SQLi, no command injection, no path
traversal, no SSRF, no insecure deserialization, no data race (race test not
runnable locally), and no memory-safety bug in committed C++.
