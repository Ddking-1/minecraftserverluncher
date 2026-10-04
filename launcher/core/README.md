# core

Shared core module of the launcher. Pure logic and persistence helpers that do
not depend on the Minecraft-specific or launch subsystems.

## Contents

- `BaseInstance` / `InstanceList` / `InstancePersistence` — abstract instance,
  a `QAbstractListModel` over instances, and JSON persistence
  (`instance.json`, `launcher.json`).
- `InstanceTypes` — shared value types (`AuthSession`, `MinecraftTarget`).
- `SettingsStore` — launcher settings with defaults, clamping and JVM-arg
  safety validation.
- `SkinModel` / `SkinCache` — skin representation and a content-addressed cache
  with SHA-256 verification.
- `RotatingFileLogger` — append logger with size-based rotation and retention.
- `FileIntegrity` — SHA-256 digest and expected-digest verification.
- `HWIDExtractor` — Windows hardware identifier extraction (bcrypt/iphlpapi).

## Ownership note

`InstanceList` stores instances as `std::unique_ptr<BaseInstance>`. Launch
tasks and their steps are owned by QObject parenting inside `launch/`, not by
this module.

## Build

Consumed as the `launcher_core` static library; requires Qt6::Core.
