# Launcher Core

Minimal C++ instance management, launch-task framework, and backend auth
client for the project's Minecraft launcher. This is the project-owned launcher
core (docs/DECISIONS.md DECISION-003); it does not vendor Prism Launcher code.

## Structure

- `core/` — `BaseInstance`, `InstanceList` (`QAbstractListModel`),
  `InstancePersistence` (JSON read/write).
- `minecraft/` — `MinecraftInstance`, `PackProfile` (version/loader
  composition).
- `auth/` — `BackendAuthClient`, `TokenStorage`, `SessionManager`. Replaces
  Prism's Mojang/MSA auth flow with the project backend auth API.
- `launch/` — `LaunchTask` / `LaunchStep` framework and the minimal launch
  steps (`ClaimAuthSession`, `ResolveLibraries`, `ResolveAssets`,
  `ExtractNatives`, `VerifyJava`, `LauncherPartLaunch`).

## Compilation status

**BLOCKED: Qt 6 is not installed in the current build environment.**

The approved architecture (DECISION-001/003) requires Qt 6 (`QAbstractListModel`,
`QObject`, `QNetworkAccessManager`, `QProcess`). The code is written against
Qt 6 but has not been compiled here because no Qt toolchain is present.

Requirements:
- CMake 3.16+
- Qt 6.2+ (Core, Gui, Widgets, Network)
- MSVC or MinGW C++17 compiler

Build (once Qt is available):
```
cmake -B build -DCMAKE_PREFIX_PATH="C:\Qt\6.x.x\msvc2019_64"
cmake --build build --config Release
```

## Notes

- Library/asset download steps are stubs that prepare the local layout; actual
  download depends on the metadata-source decision (docs/DECISIONS.md
  DECISION-005, pending).
- Token storage is a first-pass plaintext JSON file under the launcher data
  directory; OS-protected credential storage (Windows DPAPI) is a documented
  hardening step (docs/AUTH.md).
