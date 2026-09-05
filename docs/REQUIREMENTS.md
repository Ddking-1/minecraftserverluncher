# REQUIREMENTS

## Project goal

Create a Windows 10/11 Minecraft client/launcher and supporting backend for a private Minecraft network.

The first release targets one Minecraft server.

The architecture MUST support additional servers later without rewriting the launcher.

## First server

Minecraft:
1.21.11

Fabric Loader:
0.19.2

Java:
24

## Supported Java paths

### Trusted custom launcher

Windows
→ custom launcher
→ custom account authentication
→ persistent device session
→ Minecraft
→ Velocity
→ Minecraft server

Trusted launcher users do not need to perform `/login` or `/register` inside Minecraft.

### Normal Java client

Normal Java client
→ Velocity
→ trusted launcher authorization check
→ if not trusted:
→ LimboFilter
→ CAPTCHA
→ LimboAuth
→ `/login` or `/register`
→ Minecraft server

The same logical account must work in both systems.

### Bedrock

Xbox / PlayStation / Windows Bedrock / mobile / Nintendo
→ Geyser/Floodgate
→ Velocity
→ Minecraft server

Bedrock users must not be sent through the Java LimboAuth flow.

## Account requirements

A single logical account system must support:
- launcher registration
- launcher login
- normal Java `/register`
- normal Java `/login`
- website login
- Bedrock linking
- skin
- future cosmetics
- future shop
- future leaderboard
- future multi-server access

Java and Bedrock identities must be distinguishable.

Username must not be the permanent database primary key.

## Session requirements

Trusted launcher accounts require:
- persistent local login
- server-side session
- short-lived access authorization
- maximum session lifetime of 30 days
- explicit logout
- logout-all
- revocation
- device association
- protected local credential storage

TPM/hardware-backed protection should be used where available but must not be a mandatory requirement.

## Trusted launcher authorization

The server must never trust username alone.

The trusted launcher must obtain server-verifiable authorization.

Velocity must validate this authorization.

Invalid authorization must not bypass normal Java authentication.

## Java authentication compatibility

If an account was created through the launcher:

The same username/password must work through normal Java `/login`.

If an account was created through normal Java `/register`:

The same credentials must work in the custom launcher.

There must not be two unrelated Java account stores.

## Bedrock identity

Bedrock identity is separate from Java identity.

A Bedrock player may be linked to a common account.

The linking mechanism must be code/verification based.

LuckPerms must not be the authoritative account-linking database.

Floodgate remains responsible for Bedrock platform identity functionality.

## Launcher

The first UI shows one server only.

The internal architecture must use server profiles so future servers can be added.

Initial launcher capabilities:
- account login
- persistent session
- device security
- launcher update
- Minecraft update
- Fabric update
- modpack update
- integrity check
- automatic RAM
- manual RAM
- JVM profile
- Java detection
- Java acquisition if required
- borderless default
- remembered window settings
- configurable resolution
- fullscreen/borderless control
- automatic server connection

## RAM

Automatic mode:
- minimum 1.5 GB
- maximum 12 GB
- retain at least 512 MB safety reserve

Manual RAM:
available in advanced settings.

The exact calculation must be determined architecturally and documented.

## Mod management

Preferred:
1. Modrinth
2. CurseForge
3. custom sources when required

Users must not freely modify protected modpack files.

Modified protected files must be detected and restored or rejected according to the security design.

Do not perform constant full-disk hashing.

## Updates

Components must be independently versionable:
- launcher
- launcher core
- server profile
- Minecraft
- Fabric
- Java runtime
- mods
- configs
- JVM profile
- resources

Requirements:
- signed manifest
- cryptographic hashes
- anti-downgrade protection
- staging
- atomic update
- rollback
- stale-file cleanup
- interrupted-download recovery
- retry/backoff
- caching where appropriate

## Skins

Desired flow:

Web / Launcher
→ Skin API
→ account/identity
→ network/server skin assignment

Skin system should not be launcher-only.

## Web/admin

The backend should eventually provide:
- account management
- sessions
- devices
- Java identities
- Bedrock identities
- linking
- skins
- server profiles
- launcher versions
- manifests
- security logs
- future configuration
- future news
- disabled future shop/cosmetics/server-selector functionality

## Infrastructure

Preferred:
- Go backend
- SQLite
- small Proxmox LXC
- low RAM/CPU consumption
- API-first
- backend should not run as root

## Future features

Prepare architecture for:
- multiple server profiles
- multiple Minecraft versions
- multiple modpacks
- multiple loaders
- server selector
- shop
- cosmetics
- leaderboard

These must remain hidden/disabled in the first product UI unless explicitly enabled.
