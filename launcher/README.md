# Launcher

This directory contains the C++/Qt6 launcher application for the private
Minecraft network. It manages local game instances, authentication against
the Go backend, and the Minecraft launch pipeline.

## Layout

- `main.cpp` — entry point and CLI argument parsing.
- `core/` — shared core: instances, persistence, settings, skin cache,
  rotating logger, file integrity, HWID extraction.
- `auth/` — backend auth client, session manager, DPAPI token storage.
- `minecraft/` — the Minecraft instance type, pack profile, mod scanner.
- `launch/` — the launch pipeline: `LaunchTask` orchestrates ordered
  `LaunchStep`s (Java verification, library/asset/native resolution, JVM start).
- `updates/` — update manifest parser and version checker.
- `game/` — Java detector and low-level game process launcher.
- `tests/` — CTest unit tests.

## How it works

Each instance is stored in a folder under `instances/` with `instance.json`,
`launcher.json` and `version.json`. `InstanceList` is a `QAbstractListModel`
that loads instances through a factory. `MinecraftInstance::createLaunchTask`
assembles a `LaunchTask` whose steps verify Java, prepare libraries/assets/
natives, then start the JVM with the backend session identity (no Mojang/MSA
tokens). Library and asset download steps are stubs pending `DECISION-005`
(docs/DECISIONS.md).

## Dependencies

- Qt 6 (Core, Network).
- Build: CMake + MSVC 2022 (Visual Studio 17), 64-bit.
- Auth storage uses Windows DPAPI (`crypt32`).

## Build & test

```bash
cmake -S launcher -B launcher/build -G "Visual Studio 17 2022" -A x64
cmake --build launcher/build --config Release
ctest --test-dir launcher/build -C Release --output-on-failure
```
