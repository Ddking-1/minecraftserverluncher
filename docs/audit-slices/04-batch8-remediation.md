# SLICE 4 — BATCH 8: Remediation & Hardening

Fixing the findings produced in Slice 3. Per rule 20 of the master directive:

`DETECT -> CLASSIFY -> ROOT CAUSE -> FIX -> TEST -> RE-SCAN -> REGRESSION CHECK`

## Scope (remediate, do not just report)

- Go backend hardening: SQLi (all parameterized — verify), path traversal on
  skin/modpack/asset handling (`../`, absolute paths, symlinks, canonical
  path escape), IDOR, broken object/function authorization, mass assignment,
  unsafe deserialization, insecure error messages, CORS, missing rate limits,
  missing request size limits, timeouts, pagination abuse, replay.
- Auth/token: full lifecycle (issuance, validation, expiry, refresh,
  rotation, revocation, logout, replay, fixation, storage, transport,
  logging, error handling). Use `subtle.ConstantTimeCompare` where constant
  time is justified. Verify the DESIGN, not just the presence of `subtle`.
- Go concurrency/memory: `go test -race ./...` green; hunt goroutine leaks,
  deadlock, channel lifecycle, context cancellation, worker shutdown,
  unclosed `http.Response.Body` / `db.Rows` / files, timers, unbounded queues,
  unbounded request/JSON input.
- C++/Qt memory & RAII: `unique_ptr`, stack ownership, no raw owning pointers;
  check double-delete, use-after-free, dangling captures, QObject lifetime,
  signal/slot lifetime, cross-thread UI access, buffer/integer overflow,
  unchecked allocation/network results.
- Launcher attack surface: JRE/JDK auto-detect, executable path resolution,
  QProcess, JVM arg handling, classpath construction, native lib extraction,
  modpack extraction, asset download + SHA-256, archive/symlink handling,
  filesystem permissions, env vars, session token passing, PATH hijacking,
  DLL/SO loading, local privilege boundaries.
- Velocity/Java: plugin lifecycle, listener trust boundaries, backend TLS
  validation, token storage/leakage, unsafe deserialization/reflection/
  classloading, arbitrary URL, authorization, identity validation, skin
  property validation, malformed response handling, timeout/retry, resource
  leaks, concurrency.
- CI/CD security: unpinned actions, GITHUB_TOKEN permissions, write-when-read
  is enough, untrusted PR code, secret exposure, artifact/cache poisoning,
  dependency confusion, arbitrary `${{ }}` interpolation, branch/tag trust,
  release/deploy permissions, signing credentials, build isolation/provenance.
- Secrets: no hardcoded password/API key/bearer/production credential; no
  secret in git history/CI logs/command line/query params/crash dumps/fixtures.

## NO TEST CHEATING

Never delete/weaken a security or functional test to make a fix pass. A fix
must eliminate the root cause and add a regression test proving it is gone.

## Gate

- All verified Critical + High findings fixed.
- Medium/Low fixed or explicitly documented as accepted with rationale.
- `go vet ./...` green; `go test -race ./...` green.
- Relevant C++ sanitizer / CTest green; Gradle build green; E2E green.
