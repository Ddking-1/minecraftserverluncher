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

## Alpha channel handling

The skin processor (`server/skins/ProcessSkin`) normalizes alpha channels to
prevent invisible players:

- Body region pixels (head, body, arms, legs) are forced to full opacity (alpha=255).
- The overlay layer (hat/jacket at x>=32, y>=32 in 64x64 skins) keeps its original alpha.
- Transparent pixels outside the skin template but within the 64x64 box get minimal alpha (1/255)
  to avoid rendering holes while preserving intended transparency.
- Pixels outside the 64x64 box remain fully transparent.

Only 64x64 and 64x32 PNGs are accepted. Non-PNG files and wrong dimensions are rejected.
The server re-encodes the PNG to strip metadata, EXIF, and extra chunks.

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
