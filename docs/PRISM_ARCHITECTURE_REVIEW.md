# PRISM ARCHITECTURE REVIEW

This document combines the architectural decisions and the dependency /
licensing risks of adopting the Prism minimal core. It resolves, at the
architecture level, the blocking checkpoint left by DECISION-001
("the Prism Launcher GPL analysis in docs/LICENSE_AUDIT.md").

The subsystem classification itself is in `docs/PRISM_MINIMAL_CORE.md`.
The decisions adopted here are recorded as DECISION-003, DECISION-004 and
DECISION-005 in `docs/DECISIONS.md`.

## 1. Status

- Architecture/licensing checkpoint: RESOLVED at architecture level.
- Final legal sign-off by the project owner on DECISION-004 is still
  REQUIRED before production launcher implementation starts.
- DECISION-005 (metadata source) must be resolved before the first launcher
  build.

## 2. Verified facts about the reference

- Location: `reference/prism-launcher` (local, read-only).
- Provenance: the snapshot has no VCS metadata (no `.git`). The exact
  upstream commit is UNKNOWN. This is a provenance gap: before any Prism code
  is copied or built, re-acquire Prism as a pinned clone (or submodule) and
  record the commit in `docs/LICENSE_AUDIT.md`.
- Primary license: GNU GPL v3 (Prism Launcher Contributors, 2022-2026),
  with incorporated works recorded in `COPYING.md`.

License inventory from `COPYING.md` (verified by reading the file):

| Component | License | Notes |
|---|---|---|
| Prism Launcher | GPL-3.0 | The launcher itself. |
| PolyMC | GPL-3.0 | Ancestor project. |
| MultiMC | Apache-2.0 | Heritage code ("incorporated work" notice). |
| `libraries/launcher` | GPL-3.0 with special linking exception (PolyMC Contributors) | The Java-side bootstrap library. |
| libnbt++ (`libraries/libnbtplusplus`) | LGPL-3.0-or-later | NBT library. |
| rainbow (`libraries/rainbow`) | LGPL-2.1-or-later (KGuiAddons heritage) | See also `COPYING.LIB`. |
| cmark | BSD-2-clause | Listed in `COPYING.md`; no `libraries/cmark` directory is present in this snapshot — re-verify against the pinned clone at build time. |
| Qt 6 | LGPL-3.0 | Toolkit dependency (C++20 / Qt 6 per `CMakeLists.txt`). |
| MinGW-w64 runtime | permissive with notice obligations | Relevant only if the Windows build uses the MinGW toolchain; using MSVC avoids distributing the MinGW runtime. |
| Batch icon set | attribution-free custom license | Branding replaced anyway. |
| Material Design Icons | SIL OFL 1.1 | Icon font; notice obligations if shipped. |

Vendored library directories present in the snapshot: `javacheck`,
`launcher`, `libnbtplusplus`, `LocalPeer`, `murmur2`, `qdcss`, `rainbow`.
The individual license of `javacheck`, `LocalPeer`, `murmur2` and `qdcss`
must be recorded in `docs/LICENSE_AUDIT.md` before the first build (their
license files exist in the upstream tree but were not individually read in
this task).

One build-relevant fact: the metadata endpoint is a CMake cache variable,
`Launcher_META_URL` (default `https://meta.prismlauncher.org/v1/`,
`CMakeLists.txt:200`), so pointing the launcher at a project metadata source
is a build-configuration change, not a code rewrite.

## 3. Licensing boundary resolution

### 3.1 What the GPL covers

- The distributed launcher is a modified GPLv3 work (Prism fork). Distributing
  its binaries to players triggers GPLv3 section 6: every recipient must be
  offered the complete corresponding source of the exact version they
  received.
- The project-owned launcher core is compiled into the same executable as the
  Prism core. That makes the combined launcher a single derived work: all of
  it, including project launcher core code, must be distributed under
  GPLv3-compatible terms. The project keeps its copyright, but GPL terms apply
  to the distributed launcher as a whole.
- This is acceptable only if accepted consciously: launcher core code is
  project-owned but GPL-bound at distribution time. Code that must stay
  proprietary must not be linked into the launcher executable.

### 3.2 What stays outside the launcher GPL

- The Go + SQLite backend: a separate program communicating over a network
  API. It is not a derivative of the launcher and remains proprietary in the
  private repository.
- The Velocity plugin, server-side mods, Geyser/Floodgate and the Limbo
  components: separate works with their own licenses. No AGPL Limbo code may
  be copied into the launcher or the backend; integration stays across clean
  component boundaries (already required by `docs/LICENSE_AUDIT.md`).
- Server configuration, infrastructure and deployment material: not
  distributed to players.

### 3.3 Distribution model decision (DECISION-004)

Launcher binaries will be distributed to players, who are third-party
recipients. The GPL then applies to every recipient. Options:

1. Publish the launcher source in a public repository (chosen). Per-release
   tags, preserved notices, no backend/infra/secrets in that repository.
   This is the standard practice for Prism/PolyMC forks and the lowest-risk
   compliance model.
2. Keep the source private and provide it only on request (rejected):
   operationally fragile with anonymous player downloads and easy to get
   wrong; higher legal risk for no real benefit.
3. Do not distribute a GPL launcher at all (rejected): would mean abandoning
   the Prism foundation, contradicting DECISION-001.

Consequence: the central repository stays private (backend, Velocity plugin,
infra, docs), and a separate public repository is created for the launcher
(fork + launcher core) when implementation starts. Creating that public
repository requires owner action/approval at that point (GitHub rule in
AGENTS.md).

### 3.4 Per-release obligations checklist

- Ship the GPL-3.0 text and the Prism/PolyMC/MultiMC copyright and license
  notices (preserve and carry over `COPYING.md` content).
- State that the launcher is a modified fork and that modifications were
  made (prominent notice).
- Provide the complete corresponding source for every distributed binary
  (public repository tag matching each release).
- Preserve LGPL notices and source-availability obligations for Qt,
  libnbt++ and rainbow (dynamic linking; do not statically link Qt).
- Record and preserve licenses of all vendored libraries (see section 2) in
  `docs/LICENSE_AUDIT.md` and in an in-app About/Licenses view.
- Toolchain notices: if MinGW is used, the MinGW-w64 runtime notice applies;
  prefer MSVC to keep the notice set small.
- Trademark hygiene: replace Prism name/logo/branding with project branding
  while preserving copyright attribution (required notices are about
  copyright, not branding).

### 3.5 Security consequence (must be stated explicitly)

Publishing the launcher source means every client-side check in the launcher
is public knowledge. The project's security model already assumes this: the
client is never the final security authority, server-side authorization is
authoritative (AGENTS.md security rule, `docs/THREAT_MODEL.md`). No security
mechanism may depend on launcher-source secrecy. Trusted-launcher
authorization must be server-verifiable and remain sound when the launcher
source is public.

### 3.6 Provenance requirement

Before copying or building any Prism code: re-acquire Prism at a pinned
commit, record the repository URL and commit in `docs/LICENSE_AUDIT.md`, and
diff the snapshot against it to confirm what was actually inspected.

## 4. Adopted architectural decisions

- DECISION-003 (APPROVED): the Prism minimal-core boundary defined in
  `docs/PRISM_MINIMAL_CORE.md` — instance/version handling, Fabric component
  handling, Java selection, the net/ download core and the launch framework
  are retained (adapted); auth, updater, UI and out-of-scope subsystems are
  replaced or removed.
- DECISION-004 (APPROVED direction / LEGAL REVIEW REQUIRED): GPLv3
  distribution with a public launcher source repository; private central
  repository for backend and infrastructure.
- DECISION-005 (PENDING): the metadata source for version/loader resolution.

## 5. Dependency risk register

| Risk | Impact | Mitigation |
|---|---|---|
| Qt 6 (LGPL-3) toolchain and upgrade risk | Build breaks, LGPL compliance | Pin a Qt version; dynamic linking; per-release notice check. |
| Toolchain choice (MSVC vs MinGW) | Notice obligations differ | Prefer MSVC; decide in the launcher build task. |
| `meta.prismlauncher.org` availability/trust | Version resolution fails; third party influences component data | DECISION-005: self-hosted mirror recommended; hashes are still validated on download. |
| Mojang endpoints (piston-meta, libraries CDN, resources CDN) | Downloads fail | Keep hash validation; a backend mirror of Mojang artifacts is a possible later extension, not first release. |
| Fabric loader metadata comes via the meta index | Loader updates depend on meta freshness | Covered by DECISION-005; direct Fabric meta source is a fallback option. |
| Mojang Java runtime manifests may lack a Windows Java 24 runtime | Auto Java acquisition fails | Detection-first policy with guided manual install fallback (PRISM_MINIMAL_CORE.md section 3). |
| GPLv3 source publication operations | Per-release source obligation | Release checklist in section 3.4; public repo tag per release. |
| Snapshot provenance gap (no pinned commit) | Cannot prove what was inspected/copied | Re-pin before any code copying (section 3.6). |
| Vendored libraries drift/security updates | Stale third-party code in the fork | Track upstream Prism periodically; record versions in LICENSE_AUDIT. |
| Unsigned Windows binaries (SmartScreen) | Distribution friction, not a license issue | Code signing decision deferred to the release engineering task. |

## 6. Coupling and architecture risks

- Application singleton: kept subsystems reach services through
  `APPLICATION->...`. Trimming removed services (accounts, updater, news,
  themes, profilers) must be compile-guarded; each trim step ends with a
  build check.
- Qt-model entanglement: retained subsystems (InstanceList, PackProfile,
  JavaInstallList) are Qt item models. This is accepted: Qt is retained, and
  de-Qt-ing the core would effectively be the from-scratch rewrite that
  DECISION-001 rejected.
- Auth seam: the replacement injection points are `ClaimAccount`,
  `LauncherPartLaunch` argument assembly, and `AccountList`/`AuthSession`
  consumers (PRISM_MINIMAL_CORE.md section 6).
- Offline-mode removal: offline launch support is dropped for the first
  release; consequences are acceptable because the trusted path is always
  online and normal Java clients do not use this launcher.
- Updater seam: replace `launcher/updater/` with the project updater;
  `ExternalUpdater.h` may serve as the temporary injection interface.
- Settings schema: Prism settings keys must be pruned to the project schema;
  stale keys removed during the trim pass.
- Trim strategy: incremental build-level exclusion in a defined order, with
  compilation verified after each step; no unrelated refactoring during
  trims (AGENTS.md rule 5).

## 7. Checkpoint resolution statement

The three checkpoint items from `docs/WORK_STATE.md` are resolved at the
architecture level:

1. Distribution model: GPLv3 launcher distributed to players; public launcher
   source repository (DECISION-004).
2. Modified-source publication: required per release; per-release public tags
   and notice checklist (section 3.4).
3. Public repository need: yes — for the launcher only; the central
   repository remains private (DECISION-004).

Production launcher implementation may begin only after:
- owner legal sign-off on DECISION-004, and
- resolution of DECISION-005 (metadata source).

## 8. Next steps

1. Owner: legal sign-off on DECISION-004 (manual action).
2. Architecture: resolve DECISION-005 (metadata source).
3. Launcher bootstrap task: pin the Prism upstream commit, create the
   launcher fork repository (public, per DECISION-004), record vendored
   library licenses in `docs/LICENSE_AUDIT.md`, produce the trim plan.
4. Launcher core design task: accounts/sessions/devices, launch
   authorization, server profiles, updater/integrity, RAM algorithm,
   UI concept.
