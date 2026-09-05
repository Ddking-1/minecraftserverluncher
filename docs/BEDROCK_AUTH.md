# BEDROCK AUTHENTICATION

## Connection

Bedrock:
Xbox / PlayStation / Windows Bedrock / Mobile / Nintendo
→ Geyser/Floodgate
→ Velocity
→ Minecraft server

## Separation

Bedrock players must not be forced into the Java LimboAuth authentication flow.

Floodgate is the platform identity mechanism.

## Common account

A Bedrock identity may be linked to a common website account.

Conceptually:

Account
→ BedrockIdentity
→ XUID / platform identity

Bedrock identity remains separate from JavaIdentity.

## Linking

Use a short-lived verification/link code.

Requirements:
- one-time use
- expiration
- rate limiting
- no code reuse
- no unauthorized account takeover
- clear success/failure states
- safe unlinking

LuckPerms must not be the authoritative account-linking database.

## Prefix

Floodgate should remain responsible for its Bedrock identity prefix behavior.

The database should store platform identity separately instead of treating a username prefix as the authoritative platform indicator.

## Bot considerations

Floodgate identity validation does not automatically equal complete anti-bot protection.

The infrastructure must consider:
- connection rate limiting
- login burst control
- per-IP limits
- per-identity limits
- Geyser/Velocity resource protection

## Security

Floodgate keys/secrets must be protected.

Never expose Floodgate private material in the launcher or public repository.
