# SLICE 2 — BATCH 7: Velocity Auth Proxy Plugin (Java / Gradle)

Security posture: the proxy is an enforcement point, never the source of
truth. The Go backend is authoritative. Treat all backend responses, player
metadata, texture/property data and plugin config as UNTRUSTED until validated.

## Goals (each ends green + commit)

- B7.1 Gradle wrapper (`./gradlew`) and `build.gradle.kts` setup.
- B7.2 Velocity plugin Main class (`@Plugin`), logger, data dir setup.
- B7.3 Java REST client helper talking to the Go backend.
- B7.4 `PlayerChooseInitialServer` event listener.
- B7.5 `AuthenticateEvent` event listener.
- B7.6 Skin injection engine: parse texture property from the backend response.
- B7.7 Skin injection engine: dynamically modify `GameProfile` in Velocity.
- B7.8 Velocity plugin config file handling (`velocity.toml`).
- B7.9 Local Gradle build (`./gradlew build`) and JAR check.
- B7.10 Mock proxy E2E test (Go backend + Velocity plugin response test).

## Hard rules (non-negotiable)

1. Micro-delegate to `lmstudio5`; 8-attempt auto-fix loop; Architect fallback.
2. `JAVA_HOME` must point at a JDK (e.g. `C:\Program Files\Java\jdk-25`),
   not the JRE — otherwise the build cannot compile. Do not claim green
   without actually running `./gradlew build`.
3. Backend TLS must be validated; never disable cert verification.
4. Tokens/keys come from env/config, never hardcoded, never logged.
5. Backend-supplied texture/property metadata must NOT lead to arbitrary code
   execution or unauthorized object modification (validate shape, sizes, no
   unsanitized reflection/classloading).
6. Timeouts + bounded retry on backend calls; no unbounded resource use.
7. Player identity validation before server selection.
8. Malformed backend response handling — never trust the response blindly.

## Verification gate for the slice

- `./gradlew build` green (JDK toolchain pointed correctly).
- JAR produced under `build/libs/`.
- Mock E2E green against the real Go backend.
