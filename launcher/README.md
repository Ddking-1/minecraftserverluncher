# Launcher Core

Minimal C++ instance management, launch-task framework, and backend auth
client for the project's Minecraft launcher. This is the project-owned launcher
core (docs/DECISIONS.md DECISION-003); it does not vendor Prism Launcher code.

## Structure

- `core/` — `BaseInstance`, `InstanceList` (`QAbstractListModel`),
  `InstancePersistence` (JSON read/write). Uses `std::vector<std::unique_ptr<BaseInstance>>` for safe move-only ownership.
  Includes `HWIDExtractor` for hardware identification.
- `minecraft/` — `MinecraftInstance`, `PackProfile` (version/loader
  composition).
- `auth/` — `BackendAuthClient`, `TokenStorage`, `SessionManager`. Replaces
  Prism's Mojang/MSA auth flow with the project backend auth API.
- `launch/` — `LaunchTask` / `LaunchStep` framework (`std::vector<std::unique_ptr<LaunchStep>>`) and the minimal launch
  steps (`ClaimAuthSession`, `ResolveLibraries`, `ResolveAssets`,
  `ExtractNatives`, `VerifyJava`, `LauncherPartLaunch`).

## Compilation and Execution Status

**STATUS: FULLY FUNCTIONAL & VERIFIED**

The project is successfully built and executed using Qt 6.6.3 (MSVC 2022) and CMake. Container logic has been migrated from `QList` to `std::vector` across core and launch modules to correctly support move-only `std::unique_ptr` elements.

Requirements:
- CMake 3.16+
- Qt 6.6.3 (MSVC 64-bit)
- MSVC 2022 C++17 compiler

Build:
```powershell
cmake -B build -DCMAKE_PREFIX_PATH="C:\Users\<user>\6.6.3\msvc2019_64"
cmake --build build --config Release
