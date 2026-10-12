# Code Wiki — Minecraft Server Launcher Ecosystem

> ⚠️ **Disclaimer:** This codebase is heavily generated and maintained by AI assistants. There is no guarantee of continuous updates, bug fixes, or long-term maintenance. Use at your own risk.

This document is the structured, comprehensive Code Wiki for the **Minecraft Server Launcher Ecosystem** — a private Minecraft network client ecosystem targeting Windows 10/11, Minecraft 1.21.11, Fabric 0.19.2, and Java 24.

---

## Table of Contents

1. [Project Overview](#1-project-overview)
2. [Architecture Overview](#2-architecture-overview)
3. [Module Breakdown](#3-module-breakdown)
   - 3.1 [Launcher (C++ / Qt6)](#31-launcher-c--qt6)
   - 3.2 [Backend (Go)](#32-backend-go)
   - 3.3 [Velocity Plugin (Java)](#33-velocity-plugin-java)
   - 3.4 [Scripts & Tooling](#34-scripts--tooling)
4. [Key Classes & Functions](#4-key-classes--functions)
   - 4.1 [Launcher Core](#41-launcher-core)
   - 4.2 [Backend Server](#42-backend-server)
   - 4.3 [Velocity Plugin](#43-velocity-plugin)
5. [Dependency Relationships](#5-dependency-relationships)
6. [Data Flow & Trust Boundaries](#6-data-flow--trust-boundaries)
7. [Build & Run Instructions](#7-build--run-instructions)
8. [Testing Strategy](#8-testing-strategy)
9. [Security Model](#9-security-model)
10. [Configuration Reference](#10-configuration-reference)
11. [Open Issues & Known Limitations](#11-open-issues--known-limitations)

---

## 1. Project Overview

The Minecraft Server Launcher is a multi-component system providing a custom Minecraft client/launcher with a project-owned authentication backend and a Velocity proxy plugin. The system implements a **trusted-launcher path** where the launcher authenticates against the backend, obtains a short-lived launch authorization token, and passes it to Minecraft, which is then verified server-side by the Velocity plugin.

### Target stack

| Component | Technology |
|---|---|
| Launcher | C++17, Qt6 (Core, Network, Widgets, Gui) |
| Backend | Go 1.26, SQLite (modernc.org/sqlite, pure-Go) |
| Velocity plugin | Java, Velocity API 4.2.0, Gradle (Kotlin DSL) |
| CI/CD | GitHub Actions |
| Installer | NSIS (`packaging/nsis/`) |

### Top-level repository layout

```
minecraftserverluncher/
├── launcher/        # C++/Qt6 launcher client (GPL-3.0-only)
├── server/          # Go + SQLite auth/API backend
├── velocity/        # Java Velocity proxy plugin
├── scripts/         # Pterodactyl control, e2e, helper scripts
├── packaging/       # NSIS installer
├── docs/            # Architecture, decisions, API, progress docs
├── .github/workflows/  # CI pipelines
├── reference/       # READ-ONLY third-party reference repos (not tracked)
└── AGENTS.md        # Permanent development rules
```

---

## 2. Architecture Overview

### 2.1 Approved direction (DECISION-001)

The system follows **required Prism core + own launcher core + own backend**. It is *not* a "full Prism Launcher + custom layer".

| Component | Ownership |
|---|---|
| Launcher shell/instance/launch core | Prism Launcher fork, trimmed (GPLv3) |
| Launcher core (accounts, sessions, devices, trusted-launch authorization, server profiles, updater, integrity) | Project-owned |
| Auth/API backend | Project-owned (Go + SQLite) |
| Velocity plugin/integration | Project-owned |
| Geyser / Floodgate / Limbo components | External integrations |

### 2.2 Core layers

1. **Launcher** (C++/Qt6) — instance management, auth client, skin handling, update/integrity, game bootstrap.
2. **Authentication/API backend** (Go) — accounts, sessions, devices, HWID bans, skins, updates, admin.
3. **Velocity integration** (Java) — server-side session-token verification at login.
4. **Java fallback authentication** — LimboAuth/LimboFilter/CAPTCHA for non-launcher Java clients (external).
5. **Geyser/Floodgate Bedrock integration** — Bedrock path (external).
6. **Minecraft servers** — backend game servers.
7. **Skin system** — upload, validation, avatar extraction, storage.
8. **Update/artifact infrastructure** — signed update manifests, SHA-256 verification.
9. **Admin/web system** — admin dashboard, server control.

### 2.3 Trust boundaries

**Trusted launcher Java path:**
```
Launcher → Auth backend → device/session verification → short-lived launch
authorization → Minecraft → Velocity → server
```

**Normal Java path (non-launcher):**
```
Minecraft → Velocity → trusted authorization check → if absent/invalid:
LimboFilter → CAPTCHA → LimboAuth → server
```

**Bedrock path:**
```
Bedrock → Geyser/Floodgate → Velocity → server
```

### 2.4 Account architecture

Logical entities: Account, JavaIdentity, BedrockIdentity, Device, Session, LaunchAuthorization, Skin. A single logical common account system spans Java and Bedrock identities. Key invariant: **username alone never grants trust**; server-side authorization is authoritative.

---

## 3. Module Breakdown

### 3.1 Launcher (C++ / Qt6)

Directory: [`launcher/`](file:///d:/projektek/ai/minecraftserverluncher/launcher)

The launcher is a C++17/Qt6 application. Build output: `MinecraftLauncher.exe`. It is one combined GPLv3 work (Prism fork + project launcher core); source must be published per DECISION-004.

#### Subdirectories

| Directory | Responsibility |
|---|---|
| [`core/`](file:///d:/projektek/ai/minecraftserverluncher/launcher/core) | Project-owned launcher core: instances, settings, skins, integrity, updates, HWID, logging, downloads |
| [`auth/`](file:///d:/projektek/ai/minecraftserverluncher/launcher/auth) | Backend auth HTTP client, token storage (DPAPI), session manager |
| [`game/`](file:///d:/projektek/ai/minecraftserverluncher/launcher/game) | Java runtime detection, Minecraft JVM bootstrap |
| [`gui/`](file:///d:/projektek/ai/minecraftserverluncher/launcher/gui) | Qt widgets: MainWindow, LoginDialog, SkinUploadWidget, EditorSettingsWidget, SplashScreen |
| [`launch/`](file:///d:/projektek/ai/minecraftserverluncher/launcher/launch) | LaunchStep/LaunchTask pipeline framework, Java version, individual launch steps |
| [`minecraft/`](file:///d:/projektek/ai/minecraftserverluncher/launcher/minecraft) | MinecraftInstance, PackProfile, ModScanner |
| [`ui/`](file:///d:/projektek/ai/minecraftserverluncher/launcher/ui) | SkinPreviewWidget (renders Minecraft skins) |
| [`updates/`](file:///d:/projektek/ai/minecraftserverluncher/launcher/updates) | UpdateManifest parser, UpdateChecker |
| [`tests/`](file:///d:/projektek/ai/minecraftserverluncher/launcher/tests) | CTest unit tests + libFuzzer target |

#### Entry point

[`launcher/main.cpp`](file:///d:/projektek/ai/minecraftserverluncher/launcher/main.cpp) instantiates `QApplication`, `MainWindow`, parses CLI args (`--version`, `--filter`), loads the instance list via `InstanceList`, and runs the event loop.

---

### 3.2 Backend (Go)

Directory: [`server/`](file:///d:/projektek/ai/minecraftserverluncher/server) — Go module `minecraftserverluncher/server`.

A lightweight Go HTTP server using the standard library `net/http` (no external web framework). Database: SQLite via `modernc.org/sqlite` (pure Go, no CGo).

#### Packages

| Package | Responsibility |
|---|---|
| [`api`](file:///d:/projektek/ai/minecraftserverluncher/server/api) | HTTP handlers, routing, middleware, rate limiting, request validation |
| [`auth`](file:///d:/projektek/ai/minecraftserverluncher/server/auth) | bcrypt password hashing, opaque token generation, SHA-256 token digests |
| [`cache`](file:///d:/projektek/ai/minecraftserverluncher/server/cache) | In-memory TTL cache for read-heavy endpoints |
| [`config`](file:///d:/projektek/ai/minecraftserverluncher/server/config) | Environment-variable configuration loading |
| [`database`](file:///d:/projektek/ai/minecraftserverluncher/server/database) | SQLite schema, models, data access (accounts, devices, sessions, skins, RBAC) |
| [`devices`](file:///d:/projektek/ai/minecraftserverluncher/server/devices) | Device/HWID binding service |
| [`sessions`](file:///d:/projektek/ai/minecraftserverluncher/server/sessions) | Session lifecycle: create, refresh (with rotation), revoke, validate |
| [`skins`](file:///d:/projektek/ai/minecraftserverluncher/server/skins) | Skin PNG validation, storage, avatar-head extraction |
| [`pterodactyl`](file:///d:/projektek/ai/minecraftserverluncher/server/pterodactyl) | Pterodactyl Panel API client + webhook HMAC verification |
| [`web`](file:///d:/projektek/ai/minecraftserverluncher/server/web) | Static web frontend (admin dashboard, skin manager) |

#### Entry point

[`server/main.go`](file:///d:/projektek/ai/minecraftserverluncher/server/main.go): loads config, opens DB, wires skin handler + rate limiter, starts `http.Server` with hardened timeouts, graceful shutdown on SIGINT/SIGTERM.

---

### 3.3 Velocity Plugin (Java)

Directory: [`velocity/`](file:///d:/projektek/ai/minecraftserverluncher/velocity)

A Velocity proxy plugin that enforces backend session authorization at player login. The proxy is an **enforcement point, not the source of truth** — it asks the Go backend to validate the player's session token.

#### Classes

| Class | Responsibility |
|---|---|
| [`LauncherVelocityPlugin`](file:///d:/projektek/ai/minecraftserverluncher/velocity/src/main/java/net/minecraftserverlauncher/velocity/LauncherVelocityPlugin.java) | Plugin main class; loads config, validates access tokens on `PostLoginEvent` |
| `BackendAuthClient` (`auth/`) | HTTP client to `/api/v1/proxy/auth` with admin token, bounded timeouts |
| `PluginConfig` (`config/`) | Loads `velocity.toml` (backend URL, admin token) |
| `SkinInjection` (`skin/`) | Skin injection (stub) |

---

### 3.4 Scripts & Tooling

| Script | Purpose |
|---|---|
| [`scripts/ptero_control.py`](file:///d:/projektek/ai/minecraftserverluncher/scripts/ptero_control.py) | Pterodactyl server lifecycle control (start/stop) invoked by admin endpoint |
| [`scripts/e2e.go`](file:///d:/projektek/ai/minecraftserverluncher/scripts/e2e.go) | End-to-end test harness |
| [`scripts/run_all_checks.ps1`](file:///d:/projektek/ai/minecraftserverluncher/scripts/run_all_checks.ps1) / `.sh` | Local all-checks runner |
| [`scripts/auto_runner.ps1`](file:///d:/projektek/ai/minecraftserverluncher/scripts/auto_runner.ps1) | Automation helper |

---

## 4. Key Classes & Functions

### 4.1 Launcher Core

#### `core/BaseInstance` — [`BaseInstance.h`](file:///d:/projektek/ai/minecraftserverluncher/launcher/core/BaseInstance.h)

Abstract base class for every managed instance. Concrete type: `MinecraftInstance` (in `minecraft/`).

| Member | Description |
|---|---|
| `id()` / `setId()` | Unique stable identifier (folder name) |
| `instanceRoot()` / `setInstanceRoot()` | Root directory for instance data |
| `name()` / `setName()` | Human-readable display name |
| `traits()` (pure virtual) | Instance traits (e.g. "fabric", "modloader") |
| `modsRoot()` (pure virtual) | Loader-specific mods directory |
| `saveNow()` (pure virtual) | Persist instance state to disk |
| `createLaunchTask(session, target)` (pure virtual) | Build the launch pipeline |
| `isRunning()` / `setRunning()` | Launch-in-progress flag |
| Signals: `propertiesChanged`, `runningStatusChanged` | Qt signals for UI updates |

#### `core/SettingsStore` — [`SettingsStore.h`](file:///d:/projektek/ai/minecraftserverluncher/launcher/core/SettingsStore.h)

Namespace holding launcher settings persisted as `launcher.json`.

| Function | Description |
|---|---|
| `defaults()` | Documented default settings (min 2048 MB, max 4096 MB) |
| `applyDefaults(Settings&)` | Fill unset fields with defaults |
| `clamp(Settings&)` | Enforce memory bounds: min ≥ 512, max ≥ min, max ≤ 65536 |
| `isSafeJvmArg(QStringView)` | Validates a JVM arg (non-empty, ≤1024 chars, trimmed, no control chars) |
| `validate(Settings, QString*)` | Validates memory bounds and JVM args |
| `load(path, Settings&)` | Atomic JSON read with defaults + clamping |
| `save(path, Settings)` | Atomic JSON write |

`Settings` struct: `javaPath`, `minMemoryMb`, `maxMemoryMb`, `jvmArgs`.

#### `auth/BackendAuthClient` — [`BackendAuthClient.h`](file:///d:/projektek/ai/minecraftserverluncher/launcher/auth/BackendAuthClient.h)

Async HTTP client for the backend auth API. Replaces Prism's Mojang/MSA auth flow entirely.

| Method | Endpoint |
|---|---|
| `registerAccount(username, password, deviceId, deviceName, cb)` | `POST /api/v1/auth/register` |
| `login(username, password, deviceId, deviceName, cb)` | `POST /api/v1/auth/login` |
| `refresh(refreshToken, cb)` | `POST /api/v1/auth/refresh` |
| `logout(refreshToken, cb)` | `POST /api/v1/auth/logout` |
| `logoutAll(accessToken, cb)` | `POST /api/v1/auth/logout-all` |

15-second request timeout, status-code error mapping. `AuthResult` struct: `success`, `error`, `accessToken`, `refreshToken`, `expiresIn`.

#### `core/UpdateManager` — [`UpdateManager.h`](file:///d:/projektek/ai/minecraftserverluncher/launcher/core/UpdateManager.h)

Fetches the update manifest from `/api/v1/updates/manifest`, verifies its HMAC signature, and emits `updateInfo(version, url, sha256, mandatory)`. Actual download is handled by `DownloadAsset` (SHA-256 verified).

#### `core/HWIDExtractor` — [`HWIDExtractor.h`](file:///d:/projektek/ai/minecraftserverluncher/launcher/core/HWIDExtractor.h)

Derives a SHA-256 hardware identifier from SMBIOS UUID + CPUID + primary MAC address. Windows-only (uses `bcrypt`, `iphlpapi`, `wbemuuid`).

#### `core/SkinModel` — [`SkinModel.h`](file:///d:/projektek/ai/minecraftserverluncher/launcher/core/SkinModel.h)

Namespace with `SkinVariant` enum (Classic/Slim), `SkinRef` struct (id, sha256, url, variant), JSON serialization, and validators for skin IDs and SHA-256 hashes.

#### `core/GameLauncher` — [`GameLauncher.h`](file:///d:/projektek/ai/minecraftserverluncher/launcher/core/GameLauncher.h)

Starts the Minecraft JVM via `QProcess`. **Security:** the session token is delivered via environment variable only, never on the command line (so it does not appear in process listings).

`LaunchRequest` struct: `javaPath`, `classpath`, `mainClass`, `nativesDir`, `workingDir`, `maxMemoryMb`, `minMemoryMb`, `sessionToken`, `extraJvmArgs`, `gameArgs`.

#### `launch/LaunchTask` — [`LaunchTask.h`](file:///d:/projektek/ai/minecraftserverluncher/launcher/launch/LaunchTask.h)

Orchestrates a sequence of `LaunchStep`s. Each step signals `readyForLaunch()` before the next starts. States: `NotStarted`, `Running`, `Failed`, `Aborted`, `Finished`.

Launch steps (in `launch/steps/`):
- `ClaimAuthSession` — obtains backend auth session
- `VerifyJava` — checks Java runtime
- `ResolveLibraries` — resolves mod/loader libraries
- `ResolveAssets` — resolves game assets
- `ExtractNatives` — extracts native libraries
- `LauncherPartLaunch` — assembles and launches the JVM

#### `core/FileIntegrity` — SHA-256 file verification against expected hash.

#### `core/RotatingFileLogger` — size-rotating file logger (keep N files).

#### `core/SkinCache` — content-addressed skin cache with SHA-256 verification.

#### `minecraft/MinecraftInstance` — concrete `BaseInstance` for Minecraft; holds `PackProfile`.

#### `minecraft/ModScanner` — scans `mods/` directory, lists enabled/disabled mods sorted.

#### `minecraft/PackProfile` — modpack profile (version, loader, mods).

#### `game/JavaDetector` — detects installed Java runtimes.

---

### 4.2 Backend Server

#### `api.Server` — [`api/router.go`](file:///d:/projektek/ai/minecraftserverluncher/server/api/router.go)

Holds all handler dependencies: `db`, `sessions`, `devices`, `cfg`, `limiter`, `queue`, caches, `SkinHandler`, `PterodactylWebhook`.

`NewServer(db, cfg, skinHandler)` constructs the server with TTL caches:
- `statsCache` (30s)
- `accountCache` (5 min)
- `sessionCache` (2 min)

`Handler()` returns the root `http.Handler` with three route groups:
- **Public**: `/api/v1/auth/*`, `/api/v1/updates/manifest`, `/healthz`, `/readyz`, `/api/v1/proxy/auth`, skin upload/get/avatar/delete (if `SkinHandler` set), Pterodactyl webhook
- **Protected** (bearer token): `/api/v1/account/*`, `/api/v1/skins/*` (user-scoped), `/api/v1/tasks/*`
- **Admin** (`X-Admin-Token`): `/api/admin/hwid/*`, `/api/admin/users`, `/api/server/start`

Middleware chain (outermost first): security headers → panic recovery → request ID → request log → (auth middleware if protected) → rate limiting → handler.

#### Key handler functions

| Handler | Method + Path | Description |
|---|---|---|
| `handleRegister` | `POST /api/v1/auth/register` | Create account + session |
| `handleLogin` | `POST /api/v1/auth/login` | Authenticate, issue session (timing-safe user enumeration) |
| `handleRefresh` | `POST /api/v1/auth/refresh` | Rotate refresh token, issue new access token |
| `handleLogout` | `POST /api/v1/auth/logout` | Revoke session (idempotent) |
| `handleLogoutAll` | `POST /api/v1/auth/logout-all` | Revoke all sessions for account |
| `handleForgotPassword` | `POST /api/v1/auth/forgot` | Request password reset |
| `handleResetPassword` | `POST /api/v1/auth/reset` | Complete password reset |
| `handleAccountMe` | `GET /api/v1/account/me` | Current account profile |
| `handleChangePassword` | `POST /api/v1/account/password` | Change password, revoke all sessions |
| `handleListSessions` | `GET /api/v1/account/sessions` | List account sessions |
| `handleRevokeSession` | `POST /api/v1/account/sessions/revoke` | Revoke one session |
| `handleLogoutOtherDevices` | `POST /api/v1/account/sessions/logout-others` | Revoke all except current |
| `handleUpdatesManifest` | `GET /api/v1/updates/manifest` | Signed update manifest |
| `handleProxyAuth` | `POST /api/v1/proxy/auth` | Velocity access-token validation |
| `handleBanHWID` / `handleUnbanHWID` / `handleListHWIDBans` | `/api/admin/hwid/*` | HWID ban administration |
| `handleStartServer` | `POST /api/server/start` | Pterodactyl server start (admin) |
| `handleReadyz` | `GET /readyz` | DB-ping readiness probe |

#### `database.DB` — [`database/accounts.go`](file:///d:/projektek/ai/minecraftserverluncher/server/database/accounts.go)

Wraps `*sql.DB`. `Open(path)` runs idempotent schema migration. Connection pool tuned to 1 open connection (SQLite single-writer). `OpenPooled` variant (4 conns, WAL) for higher concurrency.

**Schema tables:** `accounts`, `devices`, `hwid_bans`, `sessions`, `access_tokens`, `skins`, `skin_upload_limits`, `password_resets`, `roles`, `permissions`, `account_roles`, `role_permissions`.

**Models:** `Account`, `Device`, `Session`, `AccessToken`, `HWIDBan`, `Role`, `AccountRole`, `Permission`, `RolePermission`.

Key sentinel errors: `ErrNotFound`, `ErrConflict`, `ErrForbidden`, `ErrUnauthorized`, `ErrRoleNotFound`, `ErrPermissionDenied`.

#### `sessions.Service` — [`sessions/service.go`](file:///d:/projektek/ai/minecraftserverluncher/server/sessions/service.go)

| Method | Description |
|---|---|
| `Create(accountID, deviceID)` | New session → fresh access + refresh tokens |
| `Refresh(refreshToken)` | Validate, rotate refresh token, issue new access token |
| `ValidateAccess(accessToken)` | Returns account ID or `ErrInvalidAccess` |
| `Logout(refreshToken)` | Revoke session (idempotent) |
| `LogoutAll(accountID)` | Revoke all sessions |

**Refresh-token rotation:** on refresh, a new refresh token is issued and the old one's stored digest is replaced. A stolen refresh token is only usable once.

#### `auth` package — [`auth/auth.go`](file:///d:/projektek/ai/minecraftserverluncher/server/auth/auth.go)

| Function | Description |
|---|---|
| `HashPassword(password, cost)` | bcrypt hash |
| `VerifyPassword(hash, password)` | bcrypt compare |
| `NewToken()` | 32-byte random opaque token, base64url |
| `HashToken(token)` | SHA-256 digest, base64url (only this is persisted) |

#### `config.Config` — [`config/config.go`](file:///d:/projektek/ai/minecraftserverluncher/server/config/config.go)

All config from environment variables. Key fields: `ListenAddr`, `DBPath`, `AccessTokenTTL` (default 15 min), `RefreshTokenTTL` (default 30 days, max 30 days), `BcryptCost` (default 12), `AdminToken`, `PteroScript`, `LauncherVersion`, `UpdateBaseURL`, `UpdateSHA256`, `UpdateMandatory`, `UpdateSigningKey`.

#### `skins` package — [`skins/processor.go`](file:///d:/projektek/ai/minecraftserverluncher/server/skins/processor.go)

| Function/Type | Description |
|---|---|
| `ProcessSkin(r, maxBytes)` | Validate + re-encode PNG (strips EXIF/metadata), size ≤ 100 KB, dimensions 64x64 or 64x32 |
| `ExtractAvatarHead(data)` | 64x64 head avatar from full skin (nearest-neighbor scale) |
| `SkinStorage` (map) | In-memory skin store: `SaveSkin`, `GetSkin`, `DeleteSkin` |
| `ThreadSafeSkinHistory` | Tracks recent skins (cap 10), active-skin selection, rotation |

#### `pterodactyl.Client` — [`pterodactyl/client.go`](file:///d:/projektek/ai/minecraftserverluncher/server/pterodactyl/client.go)

| Method | Description |
|---|---|
| `StartServer(ctx, serverID)` | Start server via Panel API |
| `StopServer(ctx, serverID)` | Stop server |
| `RestartServer(ctx, serverID)` | Restart server |
| `GetServerStatus(ctx, serverID)` | Read server state |

Retry: exponential backoff with jitter on 429/5xx, max 4 retries, 15s default timeout.

---

### 4.3 Velocity Plugin

#### `LauncherVelocityPlugin` — [`LauncherVelocityPlugin.java`](file:///d:/projektek/ai/minecraftserverluncher/velocity/src/main/java/net/minecraftserverlauncher/velocity/LauncherVelocityPlugin.java)

| Method | Description |
|---|---|
| `onProxyInitialization(ProxyInitializeEvent)` | Load `velocity.toml`, construct `BackendAuthClient` |
| `onPostLogin(PostLoginEvent)` | Extract `minecraftserverlauncher-token` GameProfile property, validate via backend; disconnect if missing/invalid |
| `onPlayerChooseInitialServer(PlayerChooseInitialServerEvent)` | Seam for future server-selection policy |
| `findProperty(player, name)` | Extract a named property from the player's GameProfile |

#### `BackendAuthClient` (Java) — [`BackendAuthClient.java`](file:///d:/projektek/ai/minecraftserverluncher/velocity/src/main/java/net/minecraftserverlauncher/velocity/auth/BackendAuthClient.java)

`validateAccessToken(token)` → `POST /api/v1/proxy/auth` with `X-Admin-Token`. 3s connect timeout, 5s request timeout. Returns `CompletableFuture<Boolean>`.

---

## 5. Dependency Relationships

### 5.1 Backend internal package graph

```
main.go
  ├── api (depends on: auth, cache, config, database, devices, sessions, skins, pterodactyl, web)
  ├── config (stdlib only)
  ├── database (stdlib + modernc.org/sqlite)
  ├── auth (stdlib + golang.org/x/crypto/bcrypt)
  ├── sessions (depends on: auth, database)
  ├── devices (depends on: database)
  ├── cache (stdlib only)
  ├── skins (stdlib only: image, image/png, etc.)
  ├── pterodactyl (stdlib only: net/http, crypto/hmac, etc.)
  └── web (stdlib only)
```

No circular dependencies. `skins` and `pterodactyl` are peers; neither depends on the other.

### 5.2 Backend external dependencies (`server/go.mod`)

| Dependency | Purpose |
|---|---|
| `golang.org/x/crypto v0.56.0` | bcrypt password hashing |
| `modernc.org/sqlite v1.58.0` | Pure-Go SQLite driver (no CGo) |
| `github.com/stretchr/testify v1.9.0` | Test assertions |
| `golang.org/x/sys v0.47.0` | System calls (transitive) |
| `modernc.org/libc`, `mathutil`, `memory` | SQLite runtime support (transitive) |

### 5.3 Launcher dependencies

| Dependency | Purpose |
|---|---|
| Qt6 (Core, Network, Widgets, Gui) | GUI, networking, JSON, concurrency |
| Windows APIs: `bcrypt`, `crypt32`, `iphlpapi`, `wbemuuid`, `ole32`, `oleaut32` | HWID extraction, DPAPI token storage |

CMake options: C++17, `-Wall -Wextra -Werror` (GCC/Clang) or `/W4 /WX` (MSVC), `AUTOMOC` ON.

### 5.4 Velocity plugin dependencies

| Dependency | Purpose |
|---|---|
| `com.velocitypowered:velocity-api:4.2.0` | Velocity proxy API (compileOnly) |
| `net.kyori:adventure-text` | Component text (transitive via velocity-api) |

Java release target: 25 (configured in `build.gradle.kts`).

### 5.5 Cross-component dependencies

```
Launcher (C++)  ──HTTP──▶  Backend (Go)  ◀──HTTP──  Velocity plugin (Java)
     │                            │
     ▼                            ▼
  Minecraft JVM           Pterodactyl Panel
  (session token via           │
   environment variable)       ▼
                          Game servers
```

The launcher never directly contacts Velocity or game servers. The Velocity plugin validates tokens with the backend at login.

---

## 6. Data Flow & Trust Boundaries

### 6.1 Login + launch flow

```
1. User enters credentials in LoginDialog
2. BackendAuthClient.login() → POST /api/v1/auth/login
3. Backend: bcrypt verify → check HWID ban → upsert device → create session
4. Backend returns { access_token, refresh_token, expires_in }
5. TokenStorage (DPAPI) persists tokens
6. SessionManager tracks current session
7. On launch: ClaimAuthSession step obtains a fresh access token
8. GameLauncher starts Minecraft JVM with token in environment variable
9. Minecraft connects to Velocity proxy
10. Velocity plugin extracts token from GameProfile property
11. Velocity → POST /api/v1/proxy/auth → backend validates
12. If valid: player allowed; if invalid: disconnect
```

### 6.2 Token lifecycle

| Token | Lifetime | Storage |
|---|---|---|
| Access token | 15 minutes (default) | SHA-256 digest in `access_tokens` table |
| Refresh token | 30 days (max) | SHA-256 digest in `sessions` table; rotated on each use |

Raw tokens are never stored — only their SHA-256 digests.

---

## 7. Build & Run Instructions

### 7.1 Backend (Go)

**Prerequisites:** Go 1.24+

```powershell
cd server
go mod download
go build ./...
go vet ./...
go test ./... -count=1 -race
```

**Run:**

```powershell
cd server
$env:AUTH_DB_PATH = "auth.db"
$env:AUTH_LISTEN_ADDR = ":8080"
go run .
```

**Environment variables** (see [`server/.env.example`](file:///d:/projektek/ai/minecraftserverluncher/server/.env.example) and [§10 Configuration Reference](#10-configuration-reference)):
- `AUTH_LISTEN_ADDR` (default `:8080`)
- `AUTH_DB_PATH` (default `auth.db`)
- `AUTH_ACCESS_TOKEN_TTL` (default `15m`)
- `AUTH_REFRESH_TOKEN_TTL` (default `720h`, max `720h`)
- `AUTH_BCRYPT_COST` (default `12`)
- `AUTH_ADMIN_TOKEN` (enables admin endpoints)
- `AUTH_PTERO_SCRIPT` (default `scripts/ptero_control.py`)
- `AUTH_LAUNCHER_VERSION`, `AUTH_UPDATE_BASE_URL`, `AUTH_UPDATE_SHA256`, `AUTH_UPDATE_MANDATORY`, `AUTH_UPDATE_SIGNING_KEY`

**Health checks:**
- `GET /healthz` → `{"status":"ok"}`
- `GET /readyz` → `{"status":"ready"}` (pings DB)

### 7.2 Launcher (C++/Qt6)

**Prerequisites:** Qt6 (Core, Network, Widgets, Gui), CMake 3.16+, MSVC 2022 or GCC/Clang with C++17.

**Configure & build (Windows / MSVC):**

```powershell
cmake -B build -S launcher -G "Visual Studio 17 2022" -A x64
cmake --build build --config Release --target MinecraftLauncher
```

**Run tests:**

```powershell
cmake --build build --config Release
ctest --test-dir build --output-on-failure -C Release
```

**Run the launcher:**

```powershell
.\build\Release\MinecraftLauncher.exe
```

CLI options: `--version` / `-v`, `--filter <text>` / `-f <text>`, optional instances directory argument.

### 7.3 Velocity Plugin (Java)

**Prerequisites:** JDK 21+, Gradle wrapper (bundled).

```powershell
cd velocity
.\gradlew.bat build
```

Output: `velocity/build/libs/minecraftserverlauncher-velocity-0.1.0.jar`

**Deploy:** copy the JAR to the Velocity proxy's `plugins/` directory. Create `plugins/minecraftserverlauncher/velocity.toml` from `velocity/src/main/resources/velocity.toml.example`.

### 7.4 All-in-one local checks

```powershell
.\scripts\run_all_checks.ps1
```

### 7.5 CI/CD

GitHub Actions workflows in [`.github/workflows/`](file:///d:/projektek/ai/minecraftserverluncher/.github/workflows):

| Workflow | Jobs |
|---|---|
| `ci.yml` | Go (vet, test -race, staticcheck); Launcher (MSVC build, CTest); local-checks script validation |
| `backend.yml` | Backend-specific pipeline |
| `launcher.yml` | Launcher-specific pipeline |

---

## 8. Testing Strategy

### 8.1 Backend (Go)

- `go test ./... -count=1 -race` — 200+ tests across `api`, `auth`, `config`, `database`, `devices`, `pterodactyl`, `sessions`, `skins`, `web`.
- Fuzz tests: `server/skins/processor_fuzz_test.go`, `server/api/ratelimit_fuzz_test.go`, `server/api/auth_fuzz_test.go`, `server/api/skins_fuzz_test.go`.
- Edge-case tests: `auth_edge_test.go`, `sessions_edge_test.go`.
- Static analysis: `staticcheck`, `go vet`.

### 8.2 Launcher (C++)

CTest targets (see [`launcher/CMakeLists.txt`](file:///d:/projektek/ai/minecraftserverluncher/launcher/CMakeLists.txt)):

| Test | Covers |
|---|---|
| `hwid` | HWID extraction |
| `settings` | Instance persistence |
| `update_manifest` | Update manifest parsing |
| `skin_model` | Skin model serialization |
| `backend_contract` | Backend auth client contract |
| `settings_store` | Settings load/save/clamp |
| `mod_scanner` | Mod directory scanning |
| `skin_cache` | Content-addressed skin cache |
| `rotating_logger` | Log rotation |
| `update_checker` | Semantic version comparison |
| `file_integrity` | SHA-256 file verification |
| `editor_config` | Editor config |
| `skin_client` | Skin HTTP client |
| `download_asset` | SHA-256-verified download |
| `download_queue` | Download queue |
| `java_detector` | Java runtime detection |
| `game_paths` | Game path resolution |
| `game_launcher` | JVM bootstrap (token-in-env assertion) |
| `editor_settings_widget` | Editor settings UI |
| `login_dialog` | Login dialog UI |
| `skin_upload_widget` | Skin upload UI |
| `update_manager` | Update manager + HMAC |

GUI tests run with `QT_QPA_PLATFORM=offscreen`. LibFuzzer target: `fuzz_manifest_parser` (when `LibFuzzer` found).

---

## 9. Security Model

### 9.1 Core principles

- **Client is never the final security authority.** Server-side authorization is authoritative.
- **Never trust:** username alone, client brand, arbitrary client metadata, local files, local configuration.
- **Server-side verification** of trusted-launcher authorization is the security boundary.

### 9.2 Backend security measures

| Measure | Implementation |
|---|---|
| Password hashing | bcrypt (cost 12, configurable 4–31) |
| Token storage | Only SHA-256 digests persisted; raw tokens never stored |
| Refresh-token rotation | Old refresh token invalidated on each refresh |
| Timing-safe login | Dummy bcrypt comparison when account not found |
| HWID bans | `hwid_bans` table, checked at login |
| Rate limiting | Token bucket (10 req/min) on mutating endpoints, 429 + Retry-After |
| Security headers | HSTS, X-Content-Type-Options, X-Frame-Options DENY, Referrer-Policy no-referrer, Cache-Control no-store |
| Request body cap | 1 MiB (`http.MaxBytesReader`) |
| Admin endpoints | Disabled unless `AUTH_ADMIN_TOKEN` set; constant-time comparison |
| Pterodactyl webhook | HMAC-SHA256 signature verification, 401 on missing/invalid |
| Update manifests | HMAC-SHA256 signing, client verifies before reporting update |
| Graceful shutdown | Signal handling, 10s shutdown timeout |
| HTTP timeouts | ReadHeader 10s, Read 15s, Write 30s, Idle 60s |
| Panic recovery | Middleware recovers panics, returns 500 |

### 9.3 Launcher security measures

| Measure | Implementation |
|---|---|
| Token storage | Windows DPAPI (`TokenStorage`) |
| Session token in JVM | Passed via environment variable, never command line |
| File integrity | SHA-256 verification (`FileIntegrity`) |
| Update verification | HMAC-signed manifest + SHA-256 of downloaded artifact |
| JVM arg validation | `isSafeJvmArg` rejects control chars, enforces length |
| Memory bounds | `clamp()` enforces 512 MB – 65536 MB |
| HWID | SHA-256 of SMBIOS UUID + CPUID + MAC |

### 9.4 RBAC

The backend implements a role-based access control system: `roles`, `permissions`, `account_roles`, `role_permissions` tables. Admin helpers: `HasAdminRole`, `AddAdminRole`, `RemoveAdminRole`, `HasPermissionFast`, `AdminAccounts`.

---

## 10. Configuration Reference

### 10.1 Backend environment variables

| Variable | Default | Description |
|---|---|---|
| `AUTH_LISTEN_ADDR` | `:8080` | HTTP bind address |
| `AUTH_DB_PATH` | `auth.db` | SQLite database path |
| `AUTH_ACCESS_TOKEN_TTL` | `15m` | Access token lifetime |
| `AUTH_REFRESH_TOKEN_TTL` | `720h` | Refresh token / session lifetime (max 720h) |
| `AUTH_BCRYPT_COST` | `12` | bcrypt cost (4–31) |
| `AUTH_ADMIN_TOKEN` | *(unset)* | Admin API token; enables admin endpoints |
| `AUTH_PTERO_SCRIPT` | `scripts/ptero_control.py` | Server-control script path |
| `AUTH_LAUNCHER_VERSION` | `0.1.0` | Current launcher version for manifest |
| `AUTH_UPDATE_BASE_URL` | *(unset)* | Base URL for launcher installers |
| `AUTH_UPDATE_SHA256` | *(unset)* | SHA-256 of shipped installer |
| `AUTH_UPDATE_MANDATORY` | *(unset)* | Set to `1` to force updates |
| `AUTH_UPDATE_SIGNING_KEY` | *(unset)* | HMAC-SHA256 key for update manifest |

### 10.2 Velocity plugin (`velocity.toml`)

| Key | Description |
|---|---|
| `backendUrl` | Go backend base URL |
| `adminToken` | Admin token for `/api/v1/proxy/auth` |

### 10.3 Launcher settings (`launcher.json`)

| Field | Default | Description |
|---|---|---|
| `javaPath` | *(empty)* | Path to Java executable |
| `minMemoryMb` | `2048` | Minimum JVM heap (MB) |
| `maxMemoryMb` | `4096` | Maximum JVM heap (MB) |
| `jvmArgs` | `[]` | Extra JVM arguments |

---

## 11. Open Issues & Known Limitations

1. **`server/api` compilation issue (pre-existing):** `handlers.go` references `rateLimiter`, `newRateLimiter`, and `Server.limit`, while `ratelimit.go` defines `RateLimiter`, `NewRateLimiter`, and `HandlerMiddleware`. These need bridging/renaming for `go build ./...` to be fully green.

2. **Skin API route conflict:** `DELETE /api/v1/skins/{id}` (SkinHandler) conflicts with the legacy `DELETE /api/v1/skins/{variant}` pattern. One must be relocated before the Skin API is wired in production.

3. **`SkinHandler` / `PterodactylWebhook` not wired in `main.go`:** Both are unit-tested but `server/main.go` does not set them on the `Server`, so the endpoints are not exposed in production yet.

4. **Velocity T11 skeleton:** The Velocity plugin is scaffolded but full integration remains blocked pending DECISION-004 legal sign-off and DECISION-005 (metadata source) — both now approved.

5. **Library/asset download steps:** Launcher `ResolveLibraries` and `ResolveAssets` remain stubs pending the metadata source decision (DECISION-005 resolved: backend serves metadata).

6. **GUI wiring:** `MainWindow` skeleton + `SkinPreviewWidget` `skinUploadFinished` signal were added but not yet compiled or wired into `main.cpp`.

7. **Skin avatar extraction:** `ExtractAvatarHead` uses nearest-neighbor scaling of the full skin rather than the precise 8x8 head-region + overlay extraction documented in `BACKEND_ARCHITECTURE.md`.

---

*This Code Wiki is a living document. Update it whenever architecture, module responsibilities, or key interfaces change.*
