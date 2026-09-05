# SKIN SYSTEM

## Desired architecture

Web / Launcher
→ Skin API
→ Account / Minecraft identity
→ server/network skin assignment

## Requirements

The skin belongs to an account/identity, not only to one launcher installation.

The system should eventually support:
- custom launcher Java
- normal Java
- linked Bedrock
- future cosmetics

## User operations

Potential operations:
- upload skin
- validate skin
- select skin
- change skin
- remove/reset skin

## Security

Only the account owner may modify the account skin.

Uploaded files must be validated.

Do not accept arbitrary file paths or server-side executable content.

## Storage

Prefer simple external/object storage or a lightweight backend strategy.

Do not put large binary assets into SQLite unless architecture proves it necessary.

## Third-party system

Before implementing custom skin processing, evaluate mature existing solutions compatible with:
- Velocity
- custom/offline identities
- Geyser/Floodgate
- website/launcher workflows

Prefer reuse where legally and technically appropriate.
