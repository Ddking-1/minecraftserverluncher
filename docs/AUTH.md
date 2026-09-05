# AUTH

## Goal

Create one common account system shared by:

- custom launcher
- normal Java authentication
- website
- Bedrock linking
- future skins
- future cosmetics
- future shop

## Common account principle

There must be one logical account.

Do not create:
- a launcher-only account database
- an unrelated LimboAuth account database
- a separate website account database

without an explicit migration/synchronization architecture.

## Identity principle

Account and game identity are different concepts.

At minimum conceptually support:

Account
JavaIdentity
BedrockIdentity
Device
Session

## Registration compatibility

A user registering in the launcher must be able to use the same credentials with normal Java `/login`.

A user registering through normal Java `/register` must be able to use the same credentials in the launcher.

## Passwords

Never:
- store plaintext passwords
- log plaintext passwords
- expose password hashes unnecessarily

Password hashing must use a suitable modern password hashing method selected by the architecture.

## Session

Trusted launcher authentication must support:
- persistent login
- short-lived access authorization
- maximum session lifetime 30 days
- server-side revocation
- logout
- logout-all
- device association

The server must enforce expiration.

Client-side expiration alone is insufficient.

## Device

A device may possess a registered cryptographic key.

Private device keys remain on the client.

Prefer protected OS credential/key storage.

Hardware-backed protection is optional, not required.

## Token principle

Avoid long-lived bearer tokens wherever practical.

Short-lived authorization should be used for:
- server launch
- privileged actions

## Rate limiting

Protect:
- registration
- login
- password verification
- account linking
- session refresh
- admin login
- sensitive account operations

## Logging

Security logs may contain:
- account identifier
- event type
- time
- outcome
- relevant device/session identifier

Never log raw:
- password
- refresh secret
- private key
- authentication token

## Future WebAuthn

The account architecture should allow future WebAuthn/passkey authentication.

It is not required for the first release unless explicitly approved.
