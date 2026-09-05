# PRISM MINIMAL CORE

## Purpose

This document defines the minimum Prism Launcher functionality retained in this
project's distributed launcher, as required by DECISION-001
("required Prism core + own launcher core + own backend") and adopted as the
launcher core boundary by DECISION-003.

It is the component-level inventory that had to be produced before any
launcher implementation.

## Scope and inputs

- First release target: Windows 10/11, Minecraft 1.21.11, Fabric Loader
  0.19.2, Java 24, one server profile (see `docs/REQUIREMENTS.md`).
- Project replacements that are already decided: Prism auth is replaced by the
  project backend account system (`docs/AUTH.md`, `docs/JAVA_AUTH.md`), and the
  project ships its own signed-manifest updater/integrity system.
- Reference inspected: `reference/prism-launcher` (GPLv3; license details and
  risks in `docs/PRISM_ARCHITECTURE_REVIEW.md`).
- Inspection method: targeted file and pattern inspection only (AGENTS.md
  rule 24). Every path listed below was verified to exist in the local
  snapshot on 2026-09-05.
- Provenance warning: the snapshot contains no VCS metadata (no `.git`), so
  the exact upstream Prism commit is NOT known. It must be pinned (clone or
  submodule, per DECISION-002) before any Prism code is copied or built.

## Classification legend

- KEEP — retained from Prism essentially as-is
- KEEP BUT ADAPT — retained and modified for this project (trimmed,
  re-pointed, or project policy applied)
- REPLACE — removed from the Prism core and reimplemented in the
  project-owned launcher core
- NOT NEEDED — removed; the functionality is not required by this project

"Removed" means excluded from the distributed launcher over time, starting
with build-level exclusion, per DECISION-001. The trim is incremental, not a
big-bang rewrite.

## 1. Minecraft instance management and version handling

What Prism provides: an instance list model, a Minecraft instance
implementation, a component ("PackProfile") system that resolves the game
version and loader from a metadata index into libraries, rules and launch
arguments.

| Component (verified path) | Classification | Notes |
|---|---|---|
| `launcher/Application.h` | KEEP BUT ADAPT | Process shell. Keep owned services needed by the core (instances, metacache, network, settings, java list, meta index). Remove accounts, updater, news, themes, profilers, skin upload (see sections 6-8). |
| `launcher/InstanceList.h` | KEEP BUT ADAPT | Qt model of instances; retained because the internal model must support multiple server profiles. The UI above it is replaced. |
| `launcher/BaseInstance.h` | KEEP BUT ADAPT | Instance interface. |
| `launcher/minecraft/MinecraftInstance.h` | KEEP BUT ADAPT | Central instance implementation: profile, folders, update and launch task assembly. |
| `launcher/minecraft/PackProfile.h`, `PackProfile_p.h` | KEEP BUT ADAPT | Component list model (game version + loader). |
| `launcher/minecraft/Component.h` | KEEP BUT ADAPT | One versioned component. `KNOWN_MODLOADERS` at `Component.cpp:52-58` includes Fabric uid `net.fabricmc.fabric-loader`. |
| `launcher/minecraft/ComponentUpdateTask.h`, `ComponentUpdateTask_p.h` | KEEP BUT ADAPT | Resolves component versions from the meta index. |
| `launcher/minecraft/VersionFile.h` | KEEP BUT ADAPT | Component metadata: libraries, arguments, rules, agents, asset index. |
| `launcher/minecraft/LaunchProfile.h` | KEEP BUT ADAPT | Resolved main class, JVM/game arguments, classpath. |
| `launcher/minecraft/Rule.h`, `Agent.h`, `GradleSpecifier.h` | KEEP BUT ADAPT | Rule evaluation (retain Windows rules, prune other platforms), agents, Maven coordinates. |
| `launcher/minecraft/MojangVersionFormat.h`, `OneSixVersionFormat.h`, `ProfileUtils.h`, `ParseUtils.h` | KEEP BUT ADAPT | Version JSON parsing. |
| `launcher/minecraft/MinecraftLoadAndCheck.h` | KEEP BUT ADAPT | Load and verify components as a launch step. |
| `launcher/minecraft/VanillaInstanceCreationTask.h` | KEEP BUT ADAPT | Version + loader instance creation; will be driven by server profiles. |
| `launcher/meta/Index.h`, `VersionList.h`, `Version.h`, `BaseEntity.h`, `JsonFormat.h` | KEEP BUT ADAPT | The metadata system. Its endpoint is a build-time variable (`Launcher_META_URL`, `CMakeLists.txt:200`); the metadata source decision is DECISION-005 (PENDING). |
| `launcher/minecraft/World.h`, `WorldList.h` | NOT NEEDED | No world/saves management in the requirements. |
| `launcher/minecraft/VersionFilterData.h` | NOT NEEDED (candidate) | Expected legacy version data; verify contents during the trim pass. |
| `launcher/minecraft/ShortcutUtils.h` | NOT NEEDED | Desktop shortcuts not required. |
| `launcher/minecraft/skins/` (`SkinModel.h`, `SkinList.h`, `SkinUpload.h`, `SkinDelete.h`, `CapeChange.h`) | NOT NEEDED | Skin system is project-owned (backend/web), see `docs/SKIN_SYSTEM.md`. |

Subsystem verdict: KEEP BUT ADAPT. This is the heart of the retained Prism
core. Coupling: Qt item models (acceptable, Qt is retained) and the
Application singleton (trimmed); no direct auth coupling.

## 2. Fabric / loader handling

What Prism provides: loaders are ordinary meta components. Fabric is
identified by uid `net.fabricmc.fabric-loader` (`Component.cpp:52-58`), is
mutually exclusive with Forge/NeoForge/Quilt there, and is resolved through
the same meta index as the game version. Loader handling is data-driven, not
hard-coded per loader. Separately, `launcher/modplatform/` contains optional
mod and modpack platform integrations (Modrinth, CurseForge/Flame, ATLauncher,
Technic, FTB, packwiz).

| Component (verified path) | Classification | Notes |
|---|---|---|
| Component/PackProfile/ComponentUpdateTask loader mechanism (section 1) | KEEP BUT ADAPT | Fabric arrives as meta data for uid `net.fabricmc.fabric-loader`. Keep the multi-loader abstraction (requirements demand future multi-loader support); prune non-Fabric loader metadata over time. |
| `launcher/modplatform/ModIndex.h` | KEEP | Shared type definitions (`ModLoaderType`) used by `Component`; required while the component system is retained. |
| `launcher/modplatform/ResourceAPI.h`, `EnsureMetadataTask.h`, `CheckUpdateTask.h` | NOT NEEDED (first release) | Mod distribution is backend-driven with protected files. Keep the seam: `docs/REQUIREMENTS.md` lists Modrinth/CurseForge as preferred future mod sources. |
| `launcher/modplatform/modrinth/` (`ModrinthAPI.h`, `ModrinthCheckUpdate.h`, `ModrinthInstanceCreationTask.h`, `ModrinthPackIndex.h`, `ModrinthPackExportTask.h`) | NOT NEEDED (first release) | Future mod browsing candidate; excluded from the first build. |
| `launcher/modplatform/flame/` (`FlameAPI.h`, `FlameCheckUpdate.h`, `FlameInstanceCreationTask.h`, `FileResolvingTask.h`, `PackManifest.h`, `FlameModIndex.h`, `FlamePackExportTask.h`) | NOT NEEDED (first release) | Same as Modrinth. |
| `launcher/modplatform/atlauncher/`, `technic/`, `ftb/`, `import_ftb/`, `legacy_ftb/`, `packwiz/` | NOT NEEDED | Third-party pack platforms are not used; modpacks come from the project backend. |
| `launcher/modplatform/helpers/HashUtils.h` | KEEP BUT ADAPT (candidate) | Hash helpers may be reusable by the project integrity checks; decide at implementation. |
| `launcher/modplatform/helpers/OverrideUtils.h`, `ExportToModList.h` | NOT NEEDED | Modpack export/import tooling. |

Subsystem verdict: KEEP BUT ADAPT (loader resolution via the meta component
system); the mod platform layer is NOT NEEDED for the first release but is
removed at build level, not deleted blindly, so it can be re-enabled when
mod browsing is approved.

## 3. Java runtime selection

What Prism provides: local Java detection by scanning and by running a small
checker process, Java version parsing, and optional download of Mojang-published
Java runtimes from Mojang's runtime manifests.

| Component (verified path) | Classification | Notes |
|---|---|---|
| `launcher/java/JavaInstall.h`, `JavaInstallList.h` | KEEP BUT ADAPT | Detected Java installation list (model). |
| `launcher/java/JavaChecker.h` (+ `libraries/javacheck`) | KEEP BUT ADAPT | Runs the checker to obtain version/vendor/architecture. |
| `launcher/java/JavaVersion.h` | KEEP BUT ADAPT | Parses Java version strings, including early-access forms; must cover Java 24. |
| `launcher/java/JavaUtils.h` | KEEP BUT ADAPT | Java scanning utilities. |
| `launcher/java/JavaMetadata.h` | KEEP BUT ADAPT | Mojang runtime metadata model. |
| `launcher/java/download/ManifestDownloadTask.h`, `ArchiveDownloadTask.h` | KEEP BUT ADAPT | Mojang runtime manifest download and archive installation ("Java acquisition if required" in requirements). |
| `launcher/java/download/SymlinkTask.h` | NOT NEEDED (expected Unix-only) | Verify during the trim pass; Windows uses archive extraction. |
| `launcher/minecraft/launch/AutoInstallJava.h` | KEEP BUT ADAPT | Opt-in runtime download as a launch step. |
| Java selection UI (under `launcher/ui/`) | REPLACE | Project UI (see section 7). |

Subsystem verdict: KEEP BUT ADAPT.

Open item: verify that Mojang's runtime manifests publish a Windows Java 24
runtime. If not, the launcher must rely on detection plus guided manual
installation instead of automatic acquisition.

## 4. Libraries and assets download infrastructure

What Prism provides: a concurrent HTTP download framework with pluggable
sinks, checksum validation, and a persistent metadata cache, plus Minecraft-
specific tasks that turn resolved components into library/asset downloads
(Mojang and loader repositories as sources).

| Component (verified path) | Classification | Notes |
|---|---|---|
| `launcher/net/NetJob.h`, `NetRequest.h`, `Download.h` | KEEP | Concurrent download jobs with retry and failure collection. |
| `launcher/net/Sink.h`, `ByteArraySink.h`, `FileSink.h`, `MetaCacheSink.h`, `DummySink.h` | KEEP | Composable output sinks. |
| `launcher/net/HttpMetaCache.h` | KEEP | ETag/Last-Modified disk cache; supports the project's "use local cache and manifests" performance principle. |
| `launcher/net/ChecksumValidator.h`, `Validator.h` | KEEP | Post-download hash validation (SHA-1/SHA-256); basis for download integrity. |
| `launcher/net/HeaderProxy.h`, `RawHeaderProxy.h`, `Mode.h`, `NetUtils.h`, `Logging.h` | KEEP | Network plumbing. |
| `launcher/net/Upload.h`, `ApiUpload.h`, `ApiDownload.h`, `ApiHeaderProxy.h`, `PasteUpload.h` | NOT NEEDED | Mod-platform API-key plumbing and paste-service upload. |
| `launcher/minecraft/Library.h`, `MojangDownloadInfo.h` | KEEP BUT ADAPT | Library model with natives rules; retain Windows natives rules, prune other platforms. |
| `launcher/minecraft/AssetsUtils.h` | KEEP BUT ADAPT | Asset index parsing and asset download mapping. |
| `launcher/minecraft/update/LibrariesTask.h`, `AssetUpdateTask.h`, `FoldersTask.h` | KEEP BUT ADAPT | Download orchestration for libraries and assets. |
| `launcher/minecraft/update/LegacyFMLLibrariesTask.h` | NOT NEEDED | Legacy Forge library handling. |

Subsystem verdict: KEEP for the `net/` core (the cleanest infrastructure in
Prism: no auth coupling, no UI coupling beyond Application services); the
Minecraft-specific tasks are KEEP BUT ADAPT. Minecraft libraries/assets
continue to come from Mojang/loader CDNs with their published hashes; the
project's signed-manifest updater covers project-controlled content (mods,
configs, launcher, profiles).

## 5. Minecraft launch process

What Prism provides: a step-based launch task (about 20 steps in upstream
Prism) that prepares folders, resolves components, verifies Java, extracts
natives, claims an account, downloads missing files, and finally spawns the
game process through a small Java bootstrap, capturing its output.

| Component (verified path) | Classification | Notes |
|---|---|---|
| `launcher/launch/LaunchTask.h`, `LaunchStep.h`, `TaskStepWrapper.h`, `LogModel.h` | KEEP | Launch step framework and game log model. |
| `launcher/LaunchController.h` | KEEP BUT ADAPT | Launch orchestration; decouple from Prism accounts and Prism UI. |
| `launcher/launch/steps/CheckJava.h` | KEEP BUT ADAPT | Runs the Java checker before launch. |
| `launcher/minecraft/launch/CreateGameFolders.h` | KEEP | Ensures the instance folder structure exists. |
| `launcher/minecraft/launch/VerifyJavaInstall.h` | KEEP | Validates the selected Java meets the version requirement. |
| `launcher/minecraft/launch/ExtractNatives.h` | KEEP | Extracts Windows natives from library jars. |
| `launcher/minecraft/launch/MinecraftTarget.h` + `launcher/launch/steps/LookupServerAddress.h` | KEEP BUT ADAPT | Automatic server connection, driven by the server profile. |
| `launcher/minecraft/launch/PrintInstanceInfo.h` | KEEP BUT ADAPT | Pre-launch logging of the resolved configuration. |
| `launcher/minecraft/launch/LauncherPartLaunch.h` | KEEP BUT ADAPT | Spawns the game via the Java bootstrap. Adapt argument assembly: project session identity and launch authorization replace Prism's Minecraft-account arguments. |
| `libraries/launcher/org/prismlauncher/launcher/` (incl. `Launcher.java`) | KEEP BUT ADAPT | The Java-side process bootstrap; strip legacy code paths. |
| `libraries/launcher/legacy/` (incl. `net/minecraft/Launcher.java`) | NOT NEEDED | Legacy bootstrap path. |
| `launcher/minecraft/launch/ScanModFolders.h` | KEEP BUT ADAPT | Inventories mod folders; feeds the project's protected-file integrity check. |
| `launcher/minecraft/launch/EnsureAvailableMemory.h` | KEEP BUT ADAPT | Adapt to the project RAM policy (automatic 1.5-12 GB with a 512 MB reserve; exact algorithm to be defined by the launcher core design). |
| `launcher/minecraft/launch/AutoInstallJava.h` | KEEP BUT ADAPT | See section 3. |
| `launcher/minecraft/launch/ClaimAccount.h` | REPLACE | Prism account locking. Replaced by a project step that claims the local session and obtains short-lived launch authorization from the backend. |
| `launcher/minecraft/launch/ModMinecraftJar.h` | NOT NEEDED | Jar-mod feature; not used by a Fabric modpack. |
| `launcher/minecraft/launch/ReconstructAssets.h` | NOT NEEDED | Legacy virtual asset reconstruction. |
| `launcher/minecraft/launch/EnsureOfflineLibraries.h` | NOT NEEDED | Offline launch mode is not a project requirement for the first release. |
| `launcher/launch/steps/PreLaunchCommand.h`, `PostLaunchCommand.h` | NOT NEEDED | Arbitrary user-configured shell commands are removed for the trusted-launcher security posture. |
| `launcher/launch/steps/QuitAfterGameStop.h`, `PrintServers.h`, `TextPrint.h` | NOT NEEDED | Not required (quit-after-game, legacy servers.dat handling, banner printing). |

Subsystem verdict: KEEP the step framework and most steps; REPLACE the
account-related step; drop legacy and power-user steps.

## 6. Auth-coupled components (REPLACE)

These components are tightly coupled to Prism's own account system
(Microsoft/Xbox/Mojang authentication) and are replaced by the project-owned
launcher core. The project's network does not use Microsoft accounts at all:
Java authentication is handled by Velocity + LimboAuth for normal clients and
by trusted launcher authorization for launcher users (`docs/JAVA_AUTH.md`).

| Component (verified path) | Classification | Notes |
|---|---|---|
| `launcher/minecraft/auth/MinecraftAccount.h`, `AccountData.h` | REPLACE | Prism account model; replaced by the project account/session model (backend-issued, max 30-day session, device association). |
| `launcher/minecraft/auth/AccountList.h` | REPLACE | Prism account store (`accounts.json`); replaced by the project credential/session store with protected OS storage. |
| `launcher/minecraft/auth/AuthSession.h`, `AuthFlow.h`, `AuthStep.h`, `Parsers.h` | REPLACE | Prism auth flow abstractions; replaced by the project backend auth flow. |
| `launcher/minecraft/auth/steps/` (`MSAStep.h`, `MSADeviceCodeStep.h`, `XboxUserStep.h`, `XboxAuthorizationStep.h`, `EntitlementsStep.h`, `LauncherLoginStep.h`, `MinecraftProfileStep.h`, `GetSkinStep.h`) | REPLACE | Microsoft/Xbox/Mojang login flow; not used by this network. |
| `launcher/minecraft/launch/ClaimAccount.h` | REPLACE | See section 5. |
| `launcher/updater/` (`ExternalUpdater.h`, `PrismExternalUpdater.h`, `MacSparkleUpdater.h`, `prismupdater/`) | REPLACE | Prism's GitHub-releases updater is replaced by the project's signed-manifest updater (staging, atomic update, rollback, anti-downgrade). `ExternalUpdater.h` may be kept temporarily as the injection seam for the project updater. |

Design note for the replacement: the Minecraft client will start with a
deterministic offline-mode identity derived from the project account, and the
trusted launch authorization is proven to Velocity separately. The exact
mechanism (client mod, protocol handshake, or other) is a launcher core +
Velocity design task; the injection points identified here are
`ClaimAccount`, `LauncherPartLaunch` argument assembly, and the
`AccountList`/`AuthSession` consumers.

## 7. Application shell and UI

| Component (verified path) | Classification | Notes |
|---|---|---|
| `launcher/Application.h` | KEEP BUT ADAPT | Keep: process shell, instance list, metacache, network access manager, settings, java list, meta index. Remove: Prism accounts, Prism updater, news, themes, profilers, external tools, skin upload, Prism branding. |
| `launcher/ui/` | REPLACE | Project-owned UI: one server, login/session, settings, update progress. Individual widgets may be reused selectively at implementation time (recorded as KEEP BUT ADAPT candidates then); the page/dialog architecture is not carried over. |
| `launcher/console/` | KEEP BUT ADAPT | A minimal game log window is retained (game output visibility). Prism console extras are trimmed. |
| `launcher/settings/` | KEEP BUT ADAPT | Settings infrastructure kept; schema pruned to project settings (session, RAM auto/manual, JVM profile, window settings, language). |
| `launcher/tasks/` | KEEP | Task/concurrent-task framework used by every subsystem. |
| `launcher/include/` | KEEP BUT ADAPT | OS/filesystem helpers; prune Unix/macOS-only paths. |
| `launcher/icons/`, `launcher/resources/`, `launcher/translations/` | KEEP BUT ADAPT | Rebranded to the project identity; minimal icon set; first release is single-language (translation infrastructure may be retained or trimmed). |

## 8. NOT NEEDED subsystems (out of scope)

Verified `launcher/` subdirectories that are excluded from the minimal core:

| Path | Reason |
|---|---|
| `launcher/news/` | In-launcher news feed; the project web/admin system owns news. |
| `launcher/screenshots/` | Screenshot upload/sharing; not required. |
| `launcher/tools/` | External profiler/editor integrations (JProfiler, JVisualVM, MCEdit); not required. |
| `launcher/archive/` | Zip import/export plumbing for third-party modpack formats; backend distributes packs. |
| `launcher/filelink/` | File-association helper executable; not required for the first release. |
| `launcher/macsandbox/` | macOS-only sandboxing; Windows-only target. |
| `launcher/minecraft/World.h`, `WorldList.h`, `minecraft/skins/` | See section 1. |
| `launcher/modplatform/` platform importers | See section 2. |
| macOS/Linux platform code across the tree | Windows 10/11 is the only first-release target; platform code paths are pruned during the trim pass. |

## 9. Resulting minimal core

What remains of Prism in the distributed launcher:

- meta system + component/version resolution (Minecraft 1.21.11 + Fabric)
- instance model, trimmed to server-profile-driven instances
- net/ download core with checksum validation and metadata cache
- library/asset/native download tasks (Mojang/loader sources, hash-verified)
- Java detection, version parsing, optional Mojang runtime download
- launch step framework and most launch steps + the Java process bootstrap
- settings/tasks/include infrastructure, minimal console, rebranded resources

What is project-owned (launcher core, not Prism code):

- accounts, sessions, devices, protected credential storage
- trusted launch authorization (server-verifiable)
- server profiles
- signed-manifest updater and file integrity (protected modpack files)
- the UI
- branding

What is project-owned (backend, outside the launcher):

- auth API, metadata/manifest hosting (see DECISION-005), skins, admin

Dependency direction (launcher):

Application (adapted)
-> InstanceList / MinecraftInstance (adapted)
   -> PackProfile / Component / meta (adapted)
      -> net/ core (kept) -> Mojang/loader repositories
   -> JavaInstallList / java download (adapted)
   -> LaunchTask (kept framework)
      -> project session claim + launch authorization (REPLACED, launcher core)
      -> LauncherPartLaunch (adapted) -> Java bootstrap (adapted) -> game

## 10. Open items before implementation

1. Pin the upstream Prism commit (clone/submodule) and record it in
   `docs/LICENSE_AUDIT.md`; the current snapshot has no VCS metadata.
2. Resolve DECISION-005 (metadata source: self-hosted mirror recommended).
3. Owner legal sign-off on DECISION-004 (GPLv3 distribution model).
4. Verify Mojang runtime manifests offer a Windows Java 24 runtime; define the
   detection/guided-install fallback if not.
5. Design the launch-authorization injection (ClaimAccount replacement,
   LauncherPartLaunch arguments, offline-mode identity derivation) together
   with the Velocity plugin design.
6. Define the automatic RAM algorithm (1.5-12 GB, 512 MB reserve) in the
   launcher core design.
7. Produce the trim plan: ordered, build-level exclusion of NOT NEEDED
   subsystems with compile verification after each step.
