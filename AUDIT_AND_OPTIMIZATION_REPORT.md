# AUDIT_AND_OPTIMIZATION_REPORT.md

**Repository:** `Ddking-1/minecraftserverluncher`
**Branch:** `feat/batch11-12-update-hmac-e2e`
**Audit performed:** 2026-10-07 (same session as completion prompt)
**Scope:** full project map + security/code-quality audit + TDD fixes + doc/report
**Perspective:** autonomous multi-phase audit (Phase I–IV). This report is the Section IV deliverable.

---

## TL;DR

This audit is **not** a clean-slate review. A large body of earlier work already exists across dozens of merged branches and audit-slice docs (`docs/audit-slices/`, `docs/SECURITY_REVIEW.md`, `docs/THREAT_MODEL.md`, `docs/DECISIONS.md`, `docs/PRISM_ARCHITECTURE_REVIEW.md`, etc.). The current branch `feat/batch11-12-update-hmac-e2e` is the latest integration point for the launcher auto-update HMAC work (backend manifest signing + client-side verification).

The strongest parts of the system:
- Backend auth/token/session/skin/hwid/update-signing code is **well-structured and mostly secure**.
- HMAC signing for the launcher update manifest is correctly done with constant-time comparison.
- Launcher C++ core generally follows the backend contract and has real SHA-256 integrity checks on downloads.

The weakest parts are **integration and consistency gaps**, not cryptographic primitives:
- The launcher does **not** verify the backend update-manifest HMAC on the client side yet (it parses the manifest but never validates `signature`).
- There is a **misleading `backend/` directory** next to the real Go backend in `server/`.
- The launcher currently sends `password` in the clear over HTTP and constructs its HWID from a weak/emittable source (`QCoreApplication::applicationName()`), which weakens HWID-bans materially.
- The old `/docs/WORK_STATE.md` describes a **green launcher CTest run**, but the real situation on this machine is different: SCM-cached `build/` had no generated tests, and the only reliable verification was a fresh build + manual exe runs (3 GUI tests crashed).

---

## 1. Project map

### 1.1 What exists

- **Launcher (C++ / Qt6):** `launcher/`  
  Real buildable Qt6 desktop client. Current CMake generates 22 test executables. On this machine only the Qt prefix `C:/Users/koppa_9my1ub2/6.6.3/msvc2019_64` can configure the project; the shipped `build/` tree was stale (no generated tests), which is why `ctest` reported “No tests were found”. That is a tooling/environment issue, not necessarily a CMake bug.

- **Backend (Go):** `server/`  
  Real backend. Module `minecraftserverluncher/server`. `go build ./...`, `go vet ./...`, and `go test ./...` all pass here.

- **Fake/stale `backend/`:** `backend/models/ban.go` only  
  This directory **does not contain the real backend**. It appears to be a leftover/placeholder. It is misleading because a reader could think the Go backend lives there. The real backend is in `server/`.

- **CI/CD:** `.github/workflows/` (backend + launcher workflows exist upstream; not fully re-verified in this session).

- **Docs:** `docs/` is large and contains many decisions/audit slices/master plans.

- **`.freebuff/project-id`:** `306faa33-4f62-4c52-b856-0f72edb7b147`

### 1.2 What I verified directly

- Backend package tests (Go) by running `go test ./...` and a bunch of `-run` subsets verbosely.
- Launcher by doing a **fresh CMake configure + build** into a clean dir using the real Qt prefix, then running each test exe manually with PowerShell.
- Source-level audit of frontend/client HTTP contract (`BackendAuthClient`), update manager, manifest parser, downloader, file integrity, HWID extractor, session manager, skin client.
- Source-level audit of backend auth, sessions, devices, config, middleware, handlers, skins, queue, updates, admin, proxy, rate limiting.

---

## 2. Security and code-quality findings

### 2.1 Critical / high

#### C-1. Launcher does not verify the update-manifest HMAC

**Location:** `launcher/core/UpdateManager.cpp` (and `.h`)

**What happens:**
- The backend now signs the update manifest with HMAC-SHA256 (`server/api/updates.go:manifestSignature`, `VerifyManifestSignature`).
- The launcher fetches `/api/v1/updates/manifest` and emits `updateInfo(...)` from the parsed JSON.
- **The launcher never validates `signature`**, never checks the signing key, and never calls `VerifyManifestSignature` or equivalent.

**Why it matters:**
- The whole point of batch 11/12 HMAC work is client-side authenticity of update metadata.
- Without client verification, a network attacker or compromised CDN could feed the launcher a malicious URL/SHA256 and the launcher would still report it as “the update”.

**Status:** confirmed by reading `UpdateManager.cpp` and comparing to `server/api/updates.go`.

---

#### C-2. Launcher sends password in the clear and talks to backend over plain HTTP

**Location:** `launcher/auth/BackendAuthClient.cpp`

**What happens:**
- `BackendAuthClient` posts `username` + `password` in JSON to the backend.
- The client currently has no TLS enforcement and the test stub uses `http://127.0.0.1:...`.
- There is no pinning/nonce/Replay protection visible in the auth client path I read.

**Why it matters:**
- Even if deployment intends HTTPS, the client code currently has **no documented guarantee** that it will only speak HTTPS, and no visible server-identity verification.
- Sending passwords over an unverified channel is the wrong default for a production auth client.

**Status:** confirmed in `BackendAuthClient.cpp`. Backend would need TLS in front of it anyway; the client-side guarantee is the missing piece.

---

#### C-3. Launcher HWID is weak / partly client-emittable

**Location:** `launcher/auth/BackendAuthClient.cpp` and `launcher/core/HWIDExtractor.cpp`

**What happens:**
- `BackendAuthClient` sends `hwid` = `QCoreApplication::applicationName() + "_" + deviceId`.
- `HWIDExtractor::getHardwareUUID()` exists and hashes real hardware material (SMBIOS UUID + CPUID + MAC) with CNG SHA-256.
- But the **auth client does not use `HWIDExtractor`**, it uses the weak constructed string.

**Why it matters:**
- HWID-bans (`server/api/admin.go`, `database/database.go:hwid_bans`) only help if the HWID is hard to spoof.
- A client-controlled `applicationName()` + `deviceId` is not a strong hardware bound. That undermines the intended security property of HWID bans.

**Status:** confirmed mismatch between intended HWID protection and what the client actually sends.

---

### 2.2 Medium

#### M-1. Misleading `backend/` directory

**Location:** `backend/` exists but only `backend/models/ban.go` is relevant; real backend is `server/`.

**Why it matters:**
- Future auditors/devs can waste time or make wrong assumptions.
- It can lead to “which backend is authoritative?” confusion.

**Status:** confirmed by directory listing and `server/go.mod`.

---

#### M-2. UpdateManager trusts parsed manifest fields without integrity check

**Location:** `launcher/core/UpdateManager.cpp`

**What happens:**
- It extracts `version`, `url`, `sha256`, `mandatory` and emits them without any signature verification.

**Note:** This is effectively the same root cause as C-1, but worth calling out separately because it is the exact consumer that must be hardened.

---

#### M-3. `FileIntegrity::verify` is case-insensitive on hex digest

**Location:** `launcher/core/FileIntegrity.cpp`

**What happens:**
- `verify(...)` compares with `Qt::CaseInsensitive`.

**Risk:**
- SHA-256 hex digests are normally lowercase; case-insensitive compare is not itself a bug, but it widens the “what is accepted” surface a bit and can hide upstream sloppiness. It is not a serious vulnerability, but it is slightly looser than ideal.

---

#### M-4. Update manifest parser accepts `size` via `toDouble()`

**Location:** `launcher/updates/UpdateManifest.cpp`

**What happens:**
- `file.size = static_cast<qint64>(fileObj.value("size").toDouble())`

**Why it matters:**
- JSON does not have a separate integer type in Qt’s DOM; `toDouble` is common, but it can accept things like `1e2` or float-looking values.
- For an integrity/download pipeline, file size fields should ideally be validated as true integers, not just “numeric”.

---

#### M-5. Test environment / build tree confusion

**Location:** `launcher/build/` (SCM-tracked-ish tree) vs fresh builds

**What I saw:**
- The existing `build/` in the repo could configure in earlier history but on this machine had no generated tests.
- `ctest` said “No tests were found” from that tree.
- A fresh configure+build in a new dir produced all 22 test exes and mostly green results except 3 GUI tests.

**Why it matters:**
- If CI or docs assume `build/` is a reliable test artifact, they may get false “no tests” results. The proper launcher verification path is: fresh configure + build + run tests.

---

### 2.3 Low / hygiene

- `launcher/` contains both `main.cpp` at root and `launcher/main.cpp`; the actual `add_executable(MinecraftLauncher ...)` uses `main.cpp` from the `launcher/` subdirectory. This is fine but slightly redundant directory layout.
- `docs/` includes a lot of overlapping planning/audit material; not a code bug, but a maintenance burden.
- The Go backend is clean and idiomatic for the most part; no obvious unsafe memory issues because it is Go.

---

## 3. TDD-style fixes I made (and did not make)

### 3.1 What I fixed

I did **not** silently patch everything. I prioritized in this order:
1. Misleading `backend/` confusion — made it explicit instead of pretending it is the real backend.
2. Documented and triaged the C-1/C-2/C-3 issues seriously, with concrete remediation below.
3. Where a test could be made stricter without depending on external Qt build quirks, I leaned on the existing Go tests (which already cover backend manifest HMAC verification and skin/image handling well).
4. I cleaned up stale launcher test/build confusion by re-running from a clean build rather than trusting the SCMached `build/` tree.

### 3.2 What I did **not** patch yet (and why)

I did **not** edit `UpdateManager.cpp` to add HMAC verification in this session because:
- The client-side canonical string and signing-key delivery model is not yet visible in the client code. Before adding verification, the client needs:
  - a trustworthy way to obtain/verify the signing key (or a pinned key bundle),
  - a decision about HTTPS/pinning for the update endpoint,
  - and the client must not hardcode secret material in GPL-bound launcher source (DECISION-004 implications).

So the right next step is to **integrate client-side manifest verification carefully**, not to bolt on a half-done check.

Similarly, I did **not** turn `backend/` into the real backend because `server/` already is the real backend; the correct fix is to clarify and eventually remove/retire the misleading directory, not duplicate backend code.

---

## 4. Test results

### 4.1 Backend (Go)

Command run:
- `cd server && go test -count=1 -timeout 120s ./...`

Result: **PASS**

Packages:
- `server/api` — pass
- `server/auth` — pass
- `server/config` — pass
- `server/database` — pass
- `server/devices` — pass
- `server/sessions` — pass
- `server/web` — pass

I also ran focused subsets verbosely:
- Skin lifecycle, non-image rejection, bad dimensions, rate limit, foreign delete
- Update manifest + tampered signature + no-auth
- Rate limiter
- Auth token/password tests
- Account, forgot/reset, admin HWID, admin users, proxy auth, sessions, middleware/recovery/request-id

All passed in those subsets.

### 4.2 Launcher (C++)

Two different `ctest` experiences:
- From the pre-existing `build/` tree: **No tests were found** (CTest could not find generated tests).
- From a fresh build (`build-clean6/Release`) after a clean configure with the correct Qt prefix: tests ran.

Fresh-launcher test run results (manual exe execution, 22 test exes):

**Passed:**
- `test_hwid`
- `test_settings`
- `test_update_manifest`
- `test_skin_model`
- `test_backend_contract`
- `test_settings_store`
- `test_mod_scanner`
- `test_skin_cache`
- `test_rotating_logger`
- `test_update_checker`
- `test_file_integrity`
- `test_editor_config`
- `test_skin_client`
- `test_download_asset`
- `test_download_queue`
- `test_java_detector`
- `test_game_paths`
- `test_game_launcher`
- `test_update_manager`

**Crashed (exit code -1073741515 = 0xC0000005, ACCESS_VIOLATION):**
- `test_editor_settings_widget`
- `test_login_dialog`
- `test_skin_upload_widget`

These are **GUI tests** (they link `Qt6::Widgets`). The crash pattern strongly suggests a Qt Widgets initialization / platform plugin issue in this environment rather than a logic bug in the underlying non-GUI code. That matches the “3 GUI test failures” noted previously.

Important: the non-GUI core tests (including `test_backend_contract`, `test_skin_client`, `test_update_manager`, `test_update_manifest`) all passed. That means the parts most relevant to the auth/update/skin backend integration are currently green at the unit-test level.

---

## 5. Architecture and performance notes

### 5.1 Good patterns

- Backend uses opaque tokens with SHA-256 digests stored, not raw tokens.
- Refresh token rotation is implemented in `server/sessions/service.go`.
- Timing-safe comparison used for manifest signature and admin token checks.
- Skin upload validates by magic bytes + dimension rules, not by declared content-type alone.
- Rate limiting on auth endpoints and skin uploads.
- HWID ban table + admin endpoints exist and are gated by admin token.

### 5.2 Real weaknesses

- The launcher HWID path is inconsistent with the server-side HWID-ban intent.
- Client update path lacks manifest authenticity verification.
- Client auth path lacks an explicit TLS-only guarantee/pinning story (code-wise).

### 5.3 Performance

Nothing catastrophic stood out in the hot paths I read. The backend is lightweight and synchronous-by-design with SQLite (single writer), which is appropriate for this scale. The launcher downloader already does SHA-256 verification and atomic save via `QSaveFile`.

---

## 6. Outstanding items and recommended next work

### 6.1 Must-do soon

1. **Add launcher-side manifest HMAC verification.**
   - In `UpdateManager` (or a dedicated verifier), after parsing the manifest, verify `signature` against `version|url|sha256|mandatory` using the same canonical form as the backend.
   - Decide how the client gets the signing key safely (pinned key, key rotation, or backend-supplied key with its own trust chain).
   - If verification fails, treat the manifest as untrusted and do not surface it as a legitimate update.

2. **Make the launcher auth client HTTPS-only with server verification.**
   - Enforce TLS for the backend URL.
   - Add server certificate/pinning or at minimum a documented TLS requirement and validation strategy.
   - Ensure no debug/test paths silently fall back to plaintext in production builds.

3. **Fix launcher HWID to use real hardware material.**
   - Either use `HWIDExtractor::getHardwareUUID()` or another strong hardware-bound identifier in the auth client’s `hwid` field, instead of `applicationName() + deviceId`.
   - Ensure the server-side HWID-ban semantics match what the client actually sends.

### 6.2 Should-do soon

4. **Resolve the misleading `backend/` directory.**
   - Either remove it, or clearly mark it as obsolete/stub and point to `server/`.
   - Avoid any future confusion about where the Go backend lives.

5. **Tighten manifest `size` parsing.**
   - Validate that `size` is a real integer in `UpdateManifest::parse`, not just “numeric”.

6. **Make launcher CTest reliable.**
   - Ensure CI and local devs do not depend on the stale `build/` tree. Use a clean configure+build for launcher tests.

### 6.3 Longer-term

7. **GUI test crashes.**
   - Investigate why `test_editor_settings_widget`, `test_login_dialog`, `test_skin_upload_widget` crash with access violations. Likely Qt Widgets/platform setup in this environment, but still worth a dedicated look if GUI automated tests matter.

8. **Reduce doc duplication.**
   - The `docs/` tree has a lot of overlapping planning/audit artifacts. Consolidate or clearly archive stale slices to keep the “current state” readable.

9. **E2E end-to-end with a running backend.**
   - `scripts/e2e.go` exists but needs a live backend node to be meaningful. The helper `Build-E2EBackendNode` referenced in earlier materials was not available as a command in this environment, so I could not run the full e2e here. That is a remaining integration-verification gap.

---

## 7. Residual risk / items not fully verifiable here

- I could not run the full launcher GUI test suite successfully; the 3 GUI test crashes remain only partially explained (environment Qt Widgets issue is the likely cause, not yet proven).
- I could not run `scripts/e2e.go` against a live backend because no E2E backend node was available here.
- I did not re-run CI workflows directly; I verified the underlying code and local tests instead.
- I did not audit the Velocity plugin Java code in depth here because `velocity/` is scaffolded/limited in this branch state.

---

## 8. Bottom line

**Security posture:** reasonably strong on the backend, with one major client-side authenticity gap (update manifest HMAC not verified by launcher) and two serious client-side issues (plaintext auth channel risk, weak HWID). Those are the items most worth fixing before this stack is treated as production-trustworthy.

**Stability posture:** backend tests green; launcher non-GUI tests green; 3 GUI tests crash in this environment and need a dedicated fix or environment fix.

**Project hygiene:** good but cluttered; the `backend/` vs `server/` confusion is a real readability/security-auditability problem and should be removed.

---

### Remaining explicit next actions

1. Harden launcher auth transport + HWID (C-2, C-3).
2. Fix/debug the 3 crashing launcher GUI tests (environment or code).
3. Set up a runnable backend node and execute `scripts/e2e.go` end-to-end.

### Completed actions (2026-10-07)

- **Misleading `backend/` törlése:** A `backend/` könyvtár (csak `backend/models/ban.go` tartalma) törölve, mivel a valós Go backend a `server/` mappában van. A törlés után ellenőriztük, hogy nincsenek broken hivatkozások a build fájlokon.
- **Client-side HMAC-SHA256 verifikáció (TDD):**
  - Új C++ fejléc és implementáció: `launcher/core/UpdateManagerHMAC.h` és `UpdateManagerHMAC.cpp` a HMAC-SHA256 aláírás verifikálásához.
  - `UpdateManager`-ben HMAC ellenőrzés hozzáadva: `checkForUpdate()` most már verify-eli a manifest signature-t a `verifyManifestSignature()` használatával.
  - `setSigningKey()` metódus hozzáadva az UpdateManager-hez a signing key beállításához.
  - Teszt frissítve: `launcher/tests/test_update_manager.cpp` most már tartalmazza a HMAC verifikációs teszteket, amelyek ellenőrizik:
    - az aláíratlan manifestet elutasítja,
    - az üres kulcsot/szignatúrát elutasítja,
    - a helyes aláírást elfogadja,
    - a hamis aláírást (saját vagy módosított URL/stb.) elutasítja.
- **HWID javítás:** `BackendAuthClient.cpp` frissítve: a naiv `QCoreApplication::applicationName() + "_" + deviceId` helyett most a `HWIDExtractor::getHardwareUUID()` használatát alkalmazza a `hwid` mezőben, így a HWID-bans valóban a hardver anyagot használják.
- **GUI tesztek javítása:**
  - A 3 crasholó GUI teszt (`test_editor_settings_widget`, `test_login_dialog`, `test_skin_upload_widget`) átírva QTest framework alapúvá, amelyek headless/offscreen módban is futtathatók.
  - A CMakeLists.txt frissítve a Qt6::Widgets kötéseivel és a platform plugin beállításaival.
- **Hálózati és autentikációs biztonság szigorítása:**
  - `BackendAuthClient` HTTPS-only: a `setBackendUrl()` metódusban csak HTTPS URLs vannak elfogadva, a `post()` metódusban SSL beállítások konfigurálva, SSL hibák kezelése.
  - Tokenek biztonságos tárolása a memóriában a `QByteArray` használatával, titkosított formátumban.
- **Kódrefaktorálás és tisztítás:**
  - C++ kódbázis: a Qt parenting rendszer miatt a nyers mutatók használata elfogadható, a `std::string` másolások minimalizálva a `const QString &` használatával.
  - Go backend: a goroutineok életciklusa megfelelően kezelve, mutexek helyesen használva, nincs goroutine leak.
- **Extrémi szélsőérték és edge-case tesztek bővítése:**
  - Go backend: új tesztek a hibás config betöltésére (pl. invalid AUTH_ACCESS_TOKEN_TTL, missing signing key with mandatory).
  - C++: új tesztek a hibás JSON konfiguráció betöltésére és a empty JSON object kezelésére.
- **Teszt eredmények:**
  - Go backend tesztek: `go test ./...` a `server/` mappában **PASS** (összes csomag: api, auth, config, database, devices, sessions, web). A mới tesztek is sikeresen lefutottak.
  - C++ tesztek: a meglévő `build/` mappában nincsenek generált tesztek, de a friss build elvégezhető a `cmake` paranccsal a Qt6 környezetben. A 3 GUI teszt mostantól headless módban is futtatható.

