# DECISIONS

This file records architectural decisions.

## Decision format

### DECISION-NNN

Date:
Status:
Decision:
Reason:
Alternatives:
Consequences:
Affected files:
Reversible:
Legal review required:

## Recorded decisions

### DECISION-001 — Launcher foundation direction

Date: 2026-09-05

Status: APPROVED (direction) / LEGAL REVIEW REQUIRED (licensing)

Decision:
The launcher is built as a Prism Launcher fork, but the approved direction is
**required Prism core + own launcher core + own backend**.

This explicitly means:
- only the Prism functionality actually required by this project should remain in the distributed launcher
- unused Prism subsystems are to be removed/disabled over time, not carried along
- all project-specific launcher logic (accounts, sessions, devices, trusted launch authorization, server profiles, updater, integrity) lives in the project's own launcher core layer
- the backend is project-owned (Go + SQLite)
- the Velocity plugin is project-owned
- Geyser, Floodgate and the Limbo components are external integrations/dependencies, not project-owned code

The architecture must NOT be described as "full Prism Launcher + custom layer".

Reason:
A trimmed Prism core reduces attack surface, maintenance cost, binary size and the scope of the GPL source-distribution obligation, and keeps project-specific security logic in code the project fully controls.

Alternatives:
- launcher written from scratch (rejected for now: high cost, mature Minecraft instance/launch logic already exists in Prism)
- unmodified full Prism plus a plugin/custom layer (rejected: carries unused subsystems and blurs the ownership boundary)

Consequences:
- a component-level inventory of "required Prism functionality" must be produced before implementation
- the Prism/GPL licensing analysis in `docs/LICENSE_AUDIT.md` remains a blocking checkpoint
- the launcher core boundary must be defined architecturally before code is written

Affected files:
`docs/ARCHITECTURE.md`, `docs/AI_CONTEXT.md`, `docs/LICENSE_AUDIT.md`, `docs/THIRD_PARTY_COMPONENTS.md`

Reversible: yes, before implementation starts

Legal review required: YES (Prism Launcher GPL obligations, distribution and public source availability)

### DECISION-002 — `reference/` is not tracked in Git

Date: 2026-09-05

Status: APPROVED

Decision:
The local third-party reference sources under `reference/` are excluded from version control via `.gitignore` and are not committed to the private repository.

Reason:
They are read-only research material (~67 MB, ~4300 files across 8 upstream projects), not product source. Committing them would mix third-party code into the project history before any license decision exists, bloat the repository, and make later license/provenance auditing harder.

Alternatives:
- commit them directly (rejected: history pollution, licensing/provenance risk)
- Git submodules pinned to exact upstream commits (deferred: preferred future option because it records the exact inspected commit, which the license audit needs)

Consequences:
- each agent/developer must obtain the reference sources locally
- exact upstream commit identification for the license audit is currently NOT captured by Git and must be recorded manually in `docs/LICENSE_AUDIT.md`

Affected files:
`.gitignore`, `AGENTS.md`, `docs/THIRD_PARTY_COMPONENTS.md`

Reversible: yes (may be replaced by pinned submodules)

Legal review required: no

### DECISION-003 — Prism minimal-core boundary

Date: 2026-09-05

Status: APPROVED (architecture baseline)

Decision:
The launcher core boundary is defined by the component-level inventory in
`docs/PRISM_MINIMAL_CORE.md`:

- retained (adapted where noted): meta/version/component resolution
  (Minecraft + Fabric), the instance model, the `net/` download core,
  library/asset/native download tasks, Java detection and optional runtime
  download, the launch step framework and the Java process bootstrap,
  settings/tasks/include infrastructure, minimal console
- replaced by project-owned launcher core: all Prism auth
  (MSA/Xbox/Mojang accounts), the ClaimAccount launch step, Prism's updater,
  the UI, branding
- not needed: news, screenshots, tools, archive/import, filelink, macsandbox,
  worlds, launcher-side skins, mod-platform integrations (first release),
  legacy loaders/assets, offline mode, pre/post-launch commands

Reason:
Smallest Prism subset that still provides mature instance/version/launch
logic, per DECISION-001; everything else is replaced by project-owned code
or removed.

Alternatives:
- unmodified full Prism (rejected: carries unused subsystems, DECISION-001)
- full from-scratch rewrite (rejected: cost, DECISION-001)

Consequences:
- trimming is incremental (build-level exclusion, compile check per step);
  removed subsystems must not be re-added without a new decision
- the launch-authorization injection seam (ClaimAccount / LauncherPartLaunch
  argument assembly / AccountList-AuthSession consumers) is the interface
  between the Prism core and the project launcher core
- the metadata source is a separate pending decision (DECISION-005)

Affected files:
`docs/PRISM_MINIMAL_CORE.md`, `docs/PRISM_ARCHITECTURE_REVIEW.md`, `docs/ARCHITECTURE.md`

Reversible: yes, before implementation

Legal review required: no (licensing handled by DECISION-004)

### DECISION-004 — Launcher GPLv3 distribution and source publication model

Date: 2026-09-05

Status: APPROVED (final)

Note (2026-10-05): Executive approval granted by project owner. DECISION-004
is now final and no longer blocking. The launcher source may be published in
a separate public repository under GPL-3.0-only, with per-release source tags
and preserved third-party notices, as recorded below.

Decision:
- the distributed launcher (Prism fork + project launcher core) is one
  combined GPLv3 work
- the launcher source is published in a separate PUBLIC repository; every
  distributed release binary has a matching source tag
- the central repository stays private (backend, Velocity plugin,
  infrastructure, governance docs) and contains nothing whose publication
  the launcher GPL would require
- the backend (Go + SQLite) and all server-side components are outside the
  launcher GPL boundary (separate programs / separate works)
- Prism/PolyMC/MultiMC and third-party notices are preserved; the per-release
  obligations checklist is `docs/PRISM_ARCHITECTURE_REVIEW.md` section 3.4
- client-side launcher logic is treated as public knowledge; security must
  never depend on launcher-source secrecy (server-side authorization is
  authoritative)

Reason:
GPLv3 section 6 requires offering complete corresponding source to every
recipient of the distributed binaries; players are third-party recipients.
A public launcher source repository is the lowest-risk compliance model and
standard practice for Prism forks.

Alternatives:
- private source with source-on-request (rejected: operationally fragile,
  compliance risk with anonymous downloads)
- no GPL launcher at all (rejected: contradicts DECISION-001)

Consequences:
- launcher core code is project-copyright but GPL-bound at distribution
  time; code that must stay proprietary must not be linked into the launcher
  executable
- creating the public launcher repository requires owner action when
  implementation starts
- no secrets, backend code or infrastructure may ever be committed to the
  launcher repository

Affected files:
`docs/PRISM_ARCHITECTURE_REVIEW.md`, `docs/LICENSE_AUDIT.md`, `docs/ARCHITECTURE.md`

Reversible: yes, with effort (before first public distribution)

Legal review required: YES (final sign-off before production implementation)

### DECISION-005 — Minecraft/loader metadata source

Date: 2026-09-05

Status: APPROVED

Note (2026-10-05): Executive approval granted by project owner. DECISION-005
is now decided. The Go backend's own internal database is the authoritative
metadata source; client-side resolution must not depend on a third-party meta
endpoint. Build-time metadata serving from the backend is the implementation
path.

Decision:

- (a) self-hosted mirror of the Prism meta index on project infrastructure
  (recommended: control, availability, no third-party trust dependency)
- (b) continue using `https://meta.prismlauncher.org` (zero maintenance,
  third-party availability/trust dependency)
- (c) vendored frozen meta snapshot in the launcher repository
  (deterministic, manual updates)

The endpoint is a build-time variable (`Launcher_META_URL`, upstream
`CMakeLists.txt:200`), so switching is a build-configuration change, not a
code rewrite.

Selected: (a) self-hosted metadata served by the Go backend's internal
database; `Launcher_META_URL` is pointed at the backend.

Reason:
The meta index drives Minecraft version and loader resolution; its source
affects availability, update freshness and trust.

Consequences:
- must be resolved before the first launcher build
- if (a): backend/static-hosting scope grows slightly and a sync process is
  needed

Affected files:
`docs/PRISM_MINIMAL_CORE.md`, `docs/PRISM_ARCHITECTURE_REVIEW.md`

Reversible: yes

Legal review required: no

## Rules

Do not silently overwrite decisions.

If a decision changes:
1. create a new decision
2. reference the old decision
3. explain why it changed

Important unresolved decisions must be explicitly marked:

STATUS: PENDING

Legal uncertainty must be marked:

LEGAL REVIEW REQUIRED

## DECISION-006 — Skin file API namespace and ownership

Date: 2026-11-15
STATUS: ACCEPTED (implemented)

Context:
Two skin APIs coexisted: account-scoped named skins
(`/api/v1/skins/{variant}`, backed by the `skins` table) and id-addressed
processed skin files (upload, raw PNG, avatar, delete, activate). Both were
registered under the same `/api/v1/skins` prefix, which creates a Go 1.22+
ServeMux pattern conflict (`{id}` vs `{variant}` for `DELETE`). The file
endpoints had been parked on a mux that was never mounted, so every launcher
upload/avatar request returned 404, and the destructive file endpoints had no
authorization when they were reachable.

Decision:
- Account-scoped named skins remain under `/api/v1/skins` (bearer token).
- Processed skin files live under `/api/v1/skin-files` (bearer token):
  `POST /api/v1/skin-files`, `GET /api/v1/skin-files/{id}`,
  `GET /api/v1/skin-files/{id}/avatar`, `DELETE /api/v1/skin-files/{id}`,
  `POST /api/v1/skin-files/{id}/active`.
- A skin-file id starts with the owning account id (`<accountID>-<uuid>`);
  delete and activate require ownership and answer 404 otherwise. Reading the
  image/avatar is available to any authenticated user.
- The Pterodactyl webhook is mounted on the public group and authenticated by
  its HMAC signature. `/api/server/start` is mounted on the admin group.
- Rate limiting is applied exactly once per route group.

Consequences:
- The launcher must use `/api/v1/skin-files` and send the bearer token; it now
  syncs the access token from `SessionManager` into `BackendAuthClient`.
- The OpenAPI document, security endpoint table and endpoint READMEs are
  updated with the new namespace.
- The skin file store is concurrency-safe and can persist to disk
  (`AUTH_SKIN_STORAGE_DIR`).

Affected files:
`server/api/router.go`, `server/api/skins_handler.go`, `server/api/ratelimit.go`,
`server/api/queue.go`, `server/skins/processor.go`, `server/main.go`,
`launcher/core/SkinClient.cpp`, `launcher/auth/BackendAuthClient.cpp`,
`docs/openapi.yaml`

Reversible: yes

Legal review required: no

### DECISION-007 — Update host: GitHub Releases primary, backend as redirect/backup

Date: 2026-11-16

Status: APPROVED (owner decision, 2026-11-16 discussion)

Decision:
Update artifacts (manifest + files) are published on GitHub Releases of the
public launcher repository. The client checks BOTH the backend and the GitHub
host; the backend may tell the client that the update host moved (DNS/URL
change), and the client refreshes its stored update location from that
response. This gives a primary (GitHub) and a backup (backend-directed) path,
so a lost domain or moved repository does not brick installed clients.

Reason:
The repository is already public for legal reasons (DECISION-004/008), so
GitHub Releases is free, versioned and CDN-backed. The backend redirect keeps
clients recoverable if the GitHub URL ever changes.

Alternatives:
- backend-only hosting (rejected: single point of failure, bandwidth cost)
- client-only GitHub URL (rejected: no recovery path if the URL changes)

Consequences:
- The launcher needs an "update host" setting that can be overwritten by a
  backend response, with signature verification on the manifest (already
  implemented in UpdateManager).
- Update host changes must be authenticated (admin-only) on the backend.
- The manifest format (docs/UPDATE_SPEC.md) stays the contract.

Affected files:
`launcher/updates/UpdateManager.{h,cpp}`, `docs/UPDATE_SPEC.md`,
`server/api/*` (new redirect endpoint), `docs/AI_CONTEXT.md`

Reversible: yes

Legal review required: no

### DECISION-008 — Legal publication scope: publish only what is legally required

Date: 2026-11-16

Status: APPROVED (owner decision, 2026-11-16 discussion)

Decision:
Keep the repository private by default. Publish publicly ONLY the components
that a license (notably Prism's GPLv3 for the launcher) legally requires to be
public, and prefer publishing slightly more than the bare minimum when that
reduces legal risk. Do not publish the backend, the website, secrets, or
internal tooling unless a license requires it.

Reason:
The owner wants to minimise public exposure while staying legally safe; when in
doubt, err on the side of publishing a little more rather than risking a
license violation.

Alternatives:
- fully public repository (rejected by owner: unnecessary exposure)
- fully private with no public source (rejected: violates GPLv3 for the
  Prism-derived launcher)

Consequences:
- The launcher source (and its Prism-derived parts) must be published; the
  backend and website need not be.
- Update artifacts (manifest + files) are published on GitHub Releases of the
  public launcher repository (DECISION-007).
- A public-repository split may be needed later; document the boundary in
  docs/LICENSE_AUDIT.md.

Affected files:
`docs/LICENSE_AUDIT.md`, `docs/THIRD_PARTY_COMPONENTS.md`,
`docs/PRISM_MINIMAL_CORE.md`, `docs/PRISM_ARCHITECTURE_REVIEW.md`

Reversible: yes, before the first public release

Legal review required: YES (confirm exact GPLv3 obligations with counsel)

### DECISION-009 — Metadata source: online with local cache fallback

Date: 2026-11-16

Status: APPROVED (owner decision, 2026-11-16 discussion)

Decision:
The launcher fetches Minecraft and Fabric version metadata from the official
online sources (Mojang version manifest, Fabric meta server) at runtime, and
caches the result locally so it keeps working offline (cache-first with
network refresh, bundled fallback list as last resort).

Reason:
Always up to date without shipping a new launcher release for every Minecraft
update, while still working on a machine without internet (cache). Matches
the owner's stated preference ("online + local cache").

Alternatives:
- online only (rejected: breaks offline use)
- bundled offline list only (rejected: goes stale, requires frequent releases)

Consequences:
- `ResolveLibraries`/`ResolveAssets` launch steps can be implemented against
  this contract; network calls must be cached and must not block the UI.
- The cache location and TTL must be documented and tested.
- A bundled fallback list ships with the installer as last resort.

Affected files:
`launcher/updates/*`, `launcher/minecraft/*`, `docs/UPDATE_SPEC.md`,
`docs/AI_CONTEXT.md`

Reversible: yes

Legal review required: no
