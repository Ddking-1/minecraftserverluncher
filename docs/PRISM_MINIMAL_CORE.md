<<<<<<< HEAD
Prism Launcher Reference Inventory

Minimal Core Architecture Decision Support

Scope: D:\\projektek\\ai\\minecraftserverluncher\\reference\\prism-launcher

Windows-only target: Java 24, Minecraft 1.21.11, Fabric 0.19.2

Project replaces: Prism auth (→ backend-authenticated accounts), Prism updater (→ signed-manifest system)



SUBSYSTEM 1: Minecraft Instance Management \& Version Handling

Key Files

Path	Responsibility

launcher/InstanceList.h/.cpp	QAbstractListModel managing all instances; loads/persists instance list from disk

launcher/BaseInstance.h/.cpp	Abstract base class; defines instance interface (launch, update, settings)

launcher/minecraft/MinecraftInstance.h/.cpp	Concrete impl for Minecraft; manages profile, mods, folders, createLaunchTask/createUpdateTask

launcher/minecraft/PackProfile.h/.cpp	QAbstractListModel of version components; install/revert/customize components

launcher/minecraft/Component.h/.cpp	Represents one versioned component (Fabric, Forge, minecraft-jar, etc.); has KNOWN\_MODLOADERS map

launcher/minecraft/VersionFile.h/.cpp	Metadata structure: uid, version, name, order, libraries, rules, agents, assets index

launcher/minecraft/ComponentUpdateTask.h/.cpp	Resolves component versions from meta Index; applies dependencies/conflicts

launcher/meta/Index.h/.cpp	Root meta container; loads version lists from network (lazy on first access)

launcher/meta/VersionList.h/.cpp	Stores versions for one uid (e.g., "net.fabricmc.fabric-loader"); provides recommended/latest

launcher/meta/Version.h/.cpp	Single version in a VersionList; holds VersionFile payload

launcher/minecraft/AssetsUtils.h/.cpp	Asset index JSON parsing, virtual asset folder reconstruction, asset download URLs

launcher/minecraft/update/LibrariesTask.h/.cpp	Download task: resolves libraries from components into NetJob

launcher/minecraft/update/AssetUpdateTask.h/.cpp	Download task: assets (indexed by asset ID) from Mojang CDN

External Services / URLs

Metadata: meta.prismlauncher.org (Prism's own meta server, pulled on demand via Meta::Index)

Minecraft assets: Mojang piston-meta https://piston-meta.mojang.com/mc/ (asset index JSON per asset ID)

Minecraft libraries: Mojang CDN + fabric, forge repos (URLs extracted from component metadata → Library structs)

Legacy assets: Hardcoded fallback https://piston-meta.mojang.com/mc/assets/legacy/.../legacy.json (VersionFile)

Coupling Analysis

Application singleton: Via APPLICATION->icons(), APPLICATION->metadataIndex(), APPLICATION->metacache(), APPLICATION->settings() in MinecraftInstance

UI coupling: PackProfile is QAbstractListModel (data-driven UI tables). MinecraftInstance has methods for UI integration (loaderModList(), worldList(), etc.)

Auth coupling: MinecraftInstance::createLaunchTask needs AuthSessionPtr; no auth logic here, just plumbing

Branding: VersionFile has order hints for Prism Launcher ordering; Component has conflicting-component lists (Forge/Fabric/Quilt mutual exclusion)

Cross-platform dead weight: Asset download logic is Windows-compatible; no platform-specific branches in version resolution

Notable Facts

Data-driven: Component versioning is fully JSON/metadata-driven via Meta::VersionList

Lazy-loaded: Meta::Index loads version lists on-demand, cached via HttpMetaCache

Declarative dependency model: Components declare dependencies in VersionFile; ComponentUpdateTask resolves tree

Fabric in KNOWN\_MODLOADERS: uid = "net.fabricmc.fabric-loader" with conflicting list \[forge, neoforge, quilt]

Steps complexity: createUpdateTask() returns 4 steps (FoldersTask, LibrariesTask, LegacyFMLLibrariesTask, AssetUpdateTask); createLaunchTask() orchestrates 20+ steps

SUBSYSTEM 2: Fabric / Loader Handling

Key Files

Path	Responsibility

launcher/minecraft/Component.cpp lines 52–58	KNOWN\_MODLOADERS const map: Fabric uid + conflicting loaders

launcher/minecraft/ComponentUpdateTask.h/.cpp	Resolves Fabric version from Meta::Index via component uid resolution

launcher/modplatform/ModIndex.h	ModLoaderType enum: Fabric, Forge, NeoForge, Quilt, LiteLoader

launcher/modplatform/EnsureMetadataTask.h/.cpp	Hashes mods, queries Modrinth/Flame APIs for metadata + version info

launcher/modplatform/CheckUpdateTask.h	Abstract task to check mod updates across multiple Modrinth/Flame mods

launcher/modplatform/modrinth/ModrinthAPI.h/.cpp	REST client for Modrinth mod queries (versions, files, hashes)

launcher/modplatform/flame/FlameAPI.h/.cpp	REST client for CurseForge/Flame mod queries

launcher/minecraft/mod/Mod.h/.cpp	Metadata container (hash, provider, project ID, mod name)

External Services / URLs

Modrinth API: https://api.modrinth.com/ (mod metadata, version queries, hash lookups)

CurseForge/Flame API: https://api.curseforge.com/ (mod/modpack queries; Flame also used for instance export)

Mod files: URLs embedded in API responses (Modrinth CDN, CurseForge CDN) — hashvalidated on download

Coupling Analysis

ModIndex: Defines ModPlatform::ModLoaderType used across modplatform subsystem; independent of Component

EnsureMetadataTask \& CheckUpdateTask: Operate on Resource (Mod/Resource abstraction), not directly on Components

UI coupling: ModrinthAPI/FlameAPI are HTTP-only; UI calls CheckUpdateTask/EnsureMetadataTask as background tasks

Auth coupling: Zero — Modrinth/Flame are public APIs, no account needed

Branding: Fabric is hardcoded in KNOWN\_MODLOADERS; NeoForge/Forge/Quilt also present, but Fabric is the only Fabric loader

Notable Facts

Not required for your project: Prism's mod update system assumes free mod sources (Modrinth/Flame); your project uses backend-distributed protected files

Data-driven: ModLoader is enum-based + VersionFile declares loader dependencies; no hardcoded resolution

Duplicate capability: EnsureMetadataTask \& CheckUpdateTask both fetch metadata; modplatform system is designed for optional mod management

Platform-agnostic: All modplatform code is Windows-compatible (no OS-specific branches)

Size signal: \~8 major classes (API clients + tasks); lightweight for optional feature

SUBSYSTEM 3: Java Runtime Selection

Key Files

Path	Responsibility

launcher/java/JavaInstall.h	Struct: path, version (JavaVersion), arch, is\_64bit

launcher/java/JavaInstallList.h/.cpp	QAbstractListModel of detected/managed Java installations; loads list via JavaListLoadTask

launcher/java/JavaListLoadTask.h/.cpp	Async task: spawns JavaChecker processes in parallel; emits Result list on completion

launcher/java/JavaChecker.h/.cpp	Runs java -version or custom javacheck helper; parses stdout for version/vendor/platform

launcher/java/JavaVersion.h/.cpp	Parses Java version strings (e.g., "24", "11.0.16") into major/minor/patch/pre-release

launcher/java/JavaMetadata.h/.cpp	(Meta-like) Java runtime metadata (e.g., Mojang-provided JRE manifests for auto-download)

launcher/java/download/ManifestDownloadTask.h/.cpp	Downloads Mojang Java runtime manifest JSON

launcher/java/download/ArchiveDownloadTask.h/.cpp	Downloads \& extracts Mojang Java runtime archive (auto-install)

launcher/SysInfo.h/.cpp	System info: detected Java paths, platform, architecture

External Services / URLs

Mojang Java manifests: Implicit in download tasks; URLs hardcoded or in config (not explicitly shown in headers, but referenced in comments)

No version lookup service: Java detection is local filesystem scan + running java -version

Coupling Analysis

Application singleton: Application::javalist() returns JavaInstallList; lazily loaded

UI coupling: JavaInstallList is QAbstractListModel; JavaWizardWidget displays list in install dialog

Settings coupling: launcher/ui/widgets/JavaWizardWidget.cpp reads launcher/minecraft/MinecraftInstance.cpp settings key "JavaVersion"

Auth coupling: Zero

Branding: Auto-install of Mojang JRE is optional (AutoInstallJava launch step); can be disabled

Cross-platform dead weight: JavaChecker spawns QProcess; works on Windows, macOS, Linux equally (not relevant to Windows-only goal)

Notable Facts

Detection-first: Scans common Java paths (PATH env, C:\\Program Files\\Java, etc.); does NOT query online for installed versions

Version parsing: JavaVersion struct parses semver; supports early-access (ea) versions like "24-ea"

Launch-time checking: VerifyJavaInstall step validates selected Java meets game requirements (≥1.8 for modern Minecraft)

Auto-install optional: AutoInstallJava step downloads Mojang JRE if missing; can be turned off

Minimal network usage: Only if user opts into auto-install; otherwise local-only

SUBSYSTEM 4: Libraries \& Assets Download Infrastructure

Key Files

Path	Responsibility

launcher/net/NetJob.h/.cpp	ConcurrentTask: batch multiple NetRequest actions (downloads/uploads); manages retry, failure collection

launcher/net/Download.h/.cpp	NetRequest subclass: HTTP GET with optional caching (MetaEntryPtr), checksumming, sink selection

launcher/net/NetRequest.h/.cpp	Base HTTP action; wraps QNetworkRequest; emits progress/finish signals

launcher/net/HttpMetaCache.h/.cpp	Disk cache for HTTP metadata (ETag, Last-Modified, local file path); persistent across runs

launcher/net/Sink.h	Abstract interface: ByteArraySink, FileSink, MetaCacheSink, DummySink (write targets)

launcher/net/ByteArraySink.h	Memory sink: accumulates response in QByteArray

launcher/net/FileSink.h/.cpp	File sink: writes response to disk path

launcher/net/ChecksumValidator.h	Sink wrapper: validates SHA-1/SHA-256/MD5 after download

launcher/net/MetaCacheSink.h/.cpp	Sink wrapper: caches response + updates HttpMetaCache entry metadata

launcher/net/ApiDownload.h/.cpp	Download with ApiHeaderProxy (Modrinth/CurseForge API key injection)

launcher/minecraft/Library.h/.cpp	Represents a Java library JAR; has download URL, hash, native extraction rules

launcher/minecraft/AssetsUtils.h	Asset index struct; AssetObject with hash → download URL mapping

launcher/minecraft/update/LibrariesTask.h/.cpp	Constructs NetJob for all component libraries; applies rules (OS-specific natives)

launcher/minecraft/update/AssetUpdateTask.h/.cpp	Constructs NetJob for all asset objects from asset index

External Services / URLs

Minecraft libraries: Mojang repo https://launcher.mojang.com/v1/objects/ + third-party repos (Fabric, Forge, Maven Central)

Assets: Mojang CDN https://resources.download.minecraft.net/ (index from piston-meta; content by SHA-1 hash)

Natives: Platform-specific (Windows: .dll extraction from JARs) — part of library download logic

Coupling Analysis

Application singleton: Application::network() returns QNetworkAccessManager; Application::metacache() returns HttpMetaCache

UI coupling: MinecraftInstance::createUpdateTask() appends AssetUpdateTask/LibrariesTask to launch steps; UI shows progress via LaunchTask signals

Settings coupling: Download retry settings, proxy settings read from Application::settings()

Auth coupling: Zero for library/asset downloads; API key injection for mod platform APIs (separate ApiDownload class)

Branding: HttpMetaCache uses Application::dataRoot() for cache storage

Cross-platform: Native extraction is OS-aware (reads Library rules for platform); logic in LibrariesTask is generic

Notable Facts

Cache-first: HttpMetaCache uses ETag/Last-Modified; redownload only if remote version changed

Parallel downloads: NetJob uses ConcurrentTask; configurable max concurrency

Hash validation: ChecksumValidator wraps sinks; validates SHA-1 of library/asset after download, fails step if mismatch

Composable sinks: Sink chain pattern (MetaCacheSink wraps FileSink wraps ChecksumValidator)

Size signal: \~15 classes; moderately complex but self-contained

SUBSYSTEM 5: Minecraft Launch Process

Key Files

Path	Responsibility

launcher/launch/LaunchTask.h/.cpp	Composite task: holds list of LaunchStep objects; executes in sequence; handles process lifetime \& logging

launcher/launch/LaunchStep.h/.cpp	Abstract base for a launch step (async task or state machine)

launcher/launch/TaskStepWrapper.h/.cpp	Adapter: wraps a Task as a LaunchStep

launcher/minecraft/launch/ClaimAccount.h/.cpp	LaunchStep: locks account in-use flag; prevents concurrent logins with same account

launcher/minecraft/launch/VerifyJavaInstall.h/.cpp	LaunchStep: validates selected Java meets minimum version requirements

launcher/minecraft/launch/ExtractNatives.h/.cpp	LaunchStep: extracts .dll/.so from native JAR dependencies to temp folder

launcher/minecraft/launch/ScanModFolders.h/.cpp	LaunchStep: reads mods/ folder; populates ModFolderModel for game launch event logging

launcher/minecraft/launch/LauncherPartLaunch.h/.cpp	LaunchStep: spawns minecraft.jar process; handles stdin/stdout/stderr; manages lifetime

launcher/minecraft/launch/CreateGameFolders.h/.cpp	LaunchStep: creates .minecraft folder structure

launcher/minecraft/launch/ReconstructAssets.h/.cpp	LaunchStep: symlinks/copies virtual assets if needed (legacy versions)

launcher/minecraft/launch/PreLaunchCommand.h/.cpp	LaunchStep: runs user-configured shell command before game launch

launcher/minecraft/launch/PostLaunchCommand.h/.cpp	LaunchStep: runs user-configured shell command after game exit

launcher/minecraft/launch/AutoInstallJava.h/.cpp	LaunchStep: downloads \& installs Mojang JRE if selected Java missing

launcher/minecraft/launch/EnsureAvailableMemory.h/.cpp	LaunchStep: warns if system RAM < game heap requirement

launcher/minecraft/launch/PrintInstanceInfo.h/.cpp	LaunchStep: logs instance metadata (Java, version, args) before launch

launcher/minecraft/launch/MinecraftLoadAndCheck.h/.cpp	LaunchStep wrapper: loads PackProfile metadata (via ComponentUpdateTask)

launcher/launch/steps/TextPrint.h/.cpp	LaunchStep: prints static text to console

launcher/launch/steps/CheckJava.h/.cpp	LaunchStep: runs Java version check (JavaChecker task)

launcher/launch/steps/PreLaunchCommand.h/.cpp	(Listed above; launches pre-game script)

launcher/launch/steps/PostLaunchCommand.h/.cpp	(Listed above; launches post-game script)

launcher/launch/steps/LookupServerAddress.h/.cpp	LaunchStep: DNS lookup for server hostname (if --server parameter used)

launcher/minecraft/LaunchProfile.h/.cpp	Resolved launch data: main class, args, JVM args, libraries, classpath (built by VersionFile::applyTo)

launcher/minecraft/launch/MinecraftTarget.h/.cpp	Server/world target struct: address, port, world name

Launch Order (from MinecraftInstance::createLaunchTask)

TextPrint: "Minecraft folder is..." banner

CreateGameFolders: ensure .minecraft dirs exist

LookupServerAddress: (if --server) DNS lookup

MinecraftLoadAndCheck: load \& resolve component metadata

AutoInstallJava: (if enabled) fetch Mojang JRE

CheckJava: run java -version check

VerifyJavaInstall: validate version ≥ minimum

PreLaunchCommand: (if configured) run pre-launch script

ClaimAccount: lock account in-use

LibrariesTask, AssetUpdateTask, LegacyFMLLibrariesTask, FoldersTask: (if online mode) download resources

EnsureOfflineLibraries: (if offline mode) copy cached libraries

ModMinecraftJar: install jar mods (if any)

ScanModFolders: inventory mods

EnsureAvailableMemory: warn if RAM insufficient

PrintInstanceInfo: log final config

ExtractNatives: unzip native JARs

ReconstructAssets: setup virtual assets (legacy)

LauncherPartLaunch: spawn minecraft.jar process

PostLaunchCommand: (if configured) run post-exit script

QuitAfterGameStop: (if enabled) exit launcher when game closes

External Services / URLs

None directly in launch steps: Auth, metadata, libraries, assets all pre-fetched before launch

DNS (for server lookup): LookupServerAddress uses QHostInfo::lookupHost()

Coupling Analysis

Application/UI coupling: LaunchTask emits signals (readyForLaunch, finish, progress); UI listens for progress bar updates

Auth coupling: ClaimAccount requires AuthSessionPtr; LauncherPartLaunch receives session for auth token injection into Java args

Account data injection: AuthSession::session\_id, username, token → Java args (-Dauthlibinjector.yggdrasil.sessionserver=... for legacy, or --username=... for modern)

Settings coupling: Pre/post-launch commands, quit-on-stop setting read from instance settings; RAM requirement from instance settings

Branding: Prism-specific features: pre/post-launch commands, icon saving, console logging with color parsing

Cross-platform: Native extraction uses QDir/QFile (Windows-compatible); no OS-specific launch logic except in rules evaluation (OS-specific library filters)

Notable Facts

Step architecture: \~22 steps; highly modular; each step is re-entrant and abortable

Async throughout: Every step that does I/O is async (Task-based); UI doesn't block

Logging: LaunchTask::LogModel captures stdout/stderr with MessageLevel coloring

Censor filter: Tokens/passwords in args automatically hidden from logs

Offline mode: Skips all online steps (ClaimAccount, asset/library downloads); uses cached data

Java args assembly: LaunchProfile struct is built bottom-up from Component VersionFiles; final args list is deterministic

BRIEF SUBSYSTEM COVERAGE

Authentication System (launcher/minecraft/auth/)

File	Responsibility

launcher/minecraft/auth/MinecraftAccount.h/.cpp	Represents one authenticated Mojang/MSA account; holds ClientToken, AccessToken, profiles

launcher/minecraft/auth/AccountList.h/.cpp	QAbstractListModel of MinecraftAccount; persists to accounts.json

launcher/minecraft/auth/AccountData.h/.cpp	Serialization helper for account JSON

launcher/minecraft/auth/AuthSession.h/.cpp	Transient session data: active access token, session ID, launch mode (Online/Offline/LegacyFMod)

launcher/minecraft/auth/AuthFlow.h/.cpp	Abstract flow for account auth (MSA OAuth, Yggdrasil legacy, offline)

launcher/minecraft/auth/steps/	Specific flow steps (e.g., MSA token refresh, profile lookup)

Why REPLACE: Prism's auth system assumes Mojang/MSA accounts with token refresh. Your project has backend-authenticated accounts with 30-day session tokens; auth flow is completely different. AccountList serialization can be adapted but MinecraftAccount/AuthFlow/AuthSession are Prism-specific.



Updater System (launcher/updater/)

File	Responsibility

launcher/updater/ExternalUpdater.h	Abstract interface for external updater (platform-agnostic signals)

launcher/updater/PrismExternalUpdater.h/.cpp	Windows/Linux impl: checks GitHub Releases for new Prism Launcher versions; launches prismupdater.exe

launcher/updater/prismupdater/	Standalone updater executable source (downloads release, patches files, restarts)

Why REPLACE: Prism's updater fetches from GitHub releases and replaces launcher binaries. Your project has a signed-manifest update system via backend. The updater UI hooks (Application::updater(), triggerUpdateCheck()) can be stubbed, but the GitHub/release-checking logic is dead weight.



Application Class (launcher/Application.h/.cpp)

Global singletons owned by Application:



m\_accounts (AccountList) — REPLACE

m\_instances (InstanceList) — KEEP

m\_metacache (HttpMetaCache) — KEEP BUT ADAPT (URL endpoints change)

m\_metadataIndex (Meta::Index) — KEEP BUT ADAPT (point to your backend for metadata)

m\_javalist (JavaInstallList) — KEEP

m\_updater (ExternalUpdater) — REPLACE

m\_network (QNetworkAccessManager) — KEEP

m\_settings (SettingsObject) — KEEP (structure unchanged, content differs)

m\_icons, m\_translations, m\_themeManager, m\_profilers — NOT NEEDED for minimal core

m\_mcedit, profilers, skin upload — NOT NEEDED

Coupling to UI:



MainWindow, InstanceWindow, ViewLogWindow, SetupWizard all created/managed by Application

Application::launch() slot handles user clicks; orchestrates LaunchController

Recommendation: Keep Application structure but stub/replace the REPLACE subsystems.



NOT-NEEDED SUBSYSTEMS

Directory	Contents	Reason

launcher/news/	NewsChecker, NewsEntry	In-launcher news feed; not relevant to server launcher

launcher/screenshots/	ImgurUpload, ImgurAlbumCreation	Screenshot sharing to Imgur; not needed

launcher/ui/pages/ (partial)	GlobalSettingsPage, AccountsPage, NewsPage, SkinUploadPage	Settings UI for Prism features; your project UI is custom

launcher/ui/skins/	SkinUploadDialog, SkinViewer	Minecraft skin management; not relevant

launcher/ui/dialogs/ (partial)	ImportInstanceDialog, ExportInstanceDialog (modpack export)	Modpack import/export; your project distributes via backend

launcher/tools/	JProfiler, JVisualVM, MCEditTool, BaseProfiler	Profilers/external tools; not needed for minimal launcher

launcher/archive/	(Not examined; legacy MultiMC zip handling)	Likely legacy pack support; not needed

launcher/console/	(Not examined; likely console widget code)	UI feature; check if used by logging

launcher/filelink/	(Not examined)	Likely file association handling; Windows-only but not essential

launcher/macsandbox/	(Directory name)	macOS sandbox security; not needed (Windows-only)

launcher/themes/ (likely)	Theme management; check ITheme	UI customization; not needed for minimal core

launcher/modplatform/atlauncher/, .../technic/, .../ftb/	ATLauncher, Technic, FTB pack support	Your project uses backend-distributed packs; not needed

launcher/minecraft/skins/	SkinUpload, SkinViewer	Minecraft skins; not relevant

launcher/minecraft/mod/tasks/ (partial)	GetModDependenciesTask, resource metadata tasks	Only needed if you use Modrinth/Flame mods; your project controls mods via backend

COUPLING MAP SUMMARY

HIGHLY COUPLED (to UI, Account, Application, Branding)

Instance Management (Subsystem 1): QAbstractListModel coupling to UI tables; Application singleton access; Prism-specific folder structure

Authentication (Brief subsystem): Deep Application integration; UI login dialogs; account list persistence

Launch Process (Subsystem 5): Emits signals to UI progress bar; account locking; pre/post-launch commands (Prism-specific)

MODERATELY COUPLED (to Application, Settings, Network)

Java Runtime (Subsystem 3): Application::javalist(); settings read for Java selection; UI integration optional

Libraries \& Assets (Subsystem 4): Application::metacache/network/settings; can be refactored to pass dependencies instead of using singleton

Updater (Brief subsystem): Application::updater(); UI buttons; GitHub API integration (Prism-specific)

LOOSELY COUPLED (independent, reusable infrastructure)

Fabric Loader (Subsystem 2): Component handles Fabric as just another uid; ModIndex/EnsureMetadataTask are optional

Meta system (Subsystem 1, meta/ dir): Abstract BaseEntity/Version system; can point to any URL; cached via HttpMetaCache

CROSS-PLATFORM DEAD WEIGHT

None in subsystems 1–5 for Windows-only target

Exception: Updater system is Windows/macOS (prismupdater.exe on Windows)

Exception: JavaChecker works on all platforms (but will run on Windows without issue)

Asset/library extraction logic is platform-aware but works correctly on Windows

DEPENDENCY RELATIONSHIPS



Application

├── InstanceList ──┬─→ MinecraftInstance ──┬─→ PackProfile

│                 │                       ├─→ Component ──→ Meta::VersionList/Version

│                 │                       └─→ LaunchTask ──→ LaunchStep (20+ types)

│                 └─→ BaseInstance

├── AccountList ──→ MinecraftAccount ──→ AuthSession ──→ AuthFlow (REPLACE)

├── JavaInstallList ──→ JavaChecker (local process)

├── HttpMetaCache ──→ MetaEntry (disk cache)

├── Meta::Index ──→ Meta::VersionList/Version (lazy-loaded from network)

├── QNetworkAccessManager (for all downloads)

├── ExternalUpdater ──→ PrismExternalUpdater (REPLACE)

└── Settings ──→ instance config, Java path, console settings, etc.



MinecraftInstance::createLaunchTask() ──→ 20+ LaunchStep objects

├── ClaimAccount ──→ AccountList

├── MinecraftLoadAndCheck ──→ ComponentUpdateTask ──→ Meta::Index

├── LibrariesTask / AssetUpdateTask ──→ NetJob ──→ Download/FileSink/ChecksumValidator

└── LauncherPartLaunch ──→ AuthSession (for token injection)



Component::KNOWN\_MODLOADERS = {

&#x20; "net.fabricmc.fabric-loader": Fabric,

&#x20; "net.minecraftforge": Forge,

&#x20; "net.neoforged": NeoForge,

&#x20; "org.quiltmc.quilt-loader": Quilt,

&#x20; ...

}

FILES TO VERIFY BEFORE ARCHITECTURE FINALIZATION

CMakeLists.txt — Check if Prism meta server URL is hardcoded or configurable

launcher/Application.cpp — How Meta::Index is initialized; can it be pointed to a custom server?

launcher/minecraft/ComponentUpdateTask.cpp — Full dependency resolution algorithm

launcher/minecraft/auth/steps/ — Complete auth flow implementation (needed to understand replacement scope)

launcher/updater/prismupdater/updater\_main.cpp — Standalone updater entry point (assess removal feasibility)

Inventory prepared for minimal Prism core decision.

Classification: Ready for KEEP/KEEP-BUT-ADAPT/REPLACE/NOT-NEEDED mapping in PRISM\_MINIMAL\_CORE.md.

=======
﻿# PRISM MINIMAL CORE

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
>>>>>>> 7755c6aa4610870c08027efd34cf96233a7541d3
