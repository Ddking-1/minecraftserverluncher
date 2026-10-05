# SLICE 5 — Final Verification, Matrix & Close-out

## Steps

1. Produce the final security matrix:

```
Rule ID | Source | Severity | PASS/FINDING/N/A | Attack Surface | File(s)
        | Root Cause | Fix | Test | Verification
```

2. Sections: Critical, High, Medium, Low, N/A, External Research Findings,
   Dependency Findings, CI/CD Findings, AI/Agent Findings, Memory/Concurrency
   Findings, False Positive / False Negative Review.

3. Re-run the FULL suite after hardening to prove no new vulnerabilities were
   introduced:
   - full build
   - full unit test
   - full integration test
   - full E2E (`scripts/e2e.go` against a fresh isolated env)
   - static security scan
   - dependency scan
   - `go vet ./...`
   - `go test -race ./...`
   - C++ memory / sanitizer checks
   - Java/Gradle build
   - CI config security check
   - full checklist re-scan

4. Every finding goes through: DETECT -> CLASSIFY -> ROOT CAUSE -> FIX ->
   TEST -> RE-SCAN -> REGRESSION CHECK.

5. False-positive/false-negative review on every finding: is it really
   exploitable, what trust boundary, what attack path, what business impact,
   does the fix kill the root cause, does the same bug exist at another call
   site.

## Hardening gate (only then is the project GREEN / HARDENED)

- Full checklist audited with explicit statuses.
- No verified Critical/High finding left.
- No hardcoded production secret.
- No known auth bypass, authorization bypass, IDOR, SQLi, command injection,
  path traversal, SSRF, insecure deserialization, dependency confusion,
  hallucinated dependency, supply-chain problem, token leakage, security-
  relevant swallowed error, data race, resource leak, memory-safety bug,
  excessive filesystem/process privilege, AI-output-driven untrusted
  operation, tenant-isolation bug, or CI/CD privilege escalation.
- `go vet`, `go test -race`, CTest, Gradle, E2E all green.

## Deliverable

- `docs/audit-slices/05-matrix.md` (final matrix + all sections).
- Update `TASKS.md` marking B6/B7/B8 complete.
- Final Git commit and push to `master`.

## Honesty rule

A "GREEN / HARDENED" claim requires ALL of the above to be actually run and
green. Do not declare completion without running every check.
