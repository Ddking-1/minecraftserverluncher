# THIRD-PARTY COMPONENTS

## Local reference sources

The following repositories are stored under `reference/`:

- Prism Launcher
- Geyser
- Floodgate
- LimboAPI
- LimboAuth
- LimboFilter
- LimboAuth-ClientMod
- WebAuthn implementation/reference

## Reference rule

The source under `reference/` is read-only.

It exists so that agents can:
- inspect APIs
- inspect architecture
- compare implementations
- understand protocols
- identify reusable functionality
- perform license analysis

## Never assume

Reference source is not automatically product source.

Before copying any code:
- identify exact file
- identify upstream commit
- identify license
- record copyright
- record required notice
- determine whether copying is allowed
- determine whether publication is required

## Component status table

This document must eventually contain:

| Component | Version/Commit | License | Intended Use | Code Copied | Runtime Dependency | Public Source Needed | Status |
|---|---|---|---|---|---|---|---|

## Preferred integration strategy

Prefer, in order:

1. use a compatible external component without copying source
2. use a permissively licensed library
3. wrap/integrate through an API
4. fork only when necessary
5. rewrite only when required

Do not rewrite mature functionality without a reason.

## Attribution

Where a license requires attribution or notices:
- preserve them
- expose them in an appropriate About / Licenses section
- keep them documented in the repository
