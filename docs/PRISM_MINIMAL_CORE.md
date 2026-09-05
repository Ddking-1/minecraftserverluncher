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

