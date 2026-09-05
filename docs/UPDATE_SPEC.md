# UPDATE SPECIFICATION

## Goal

Allow independently versioned client components.

Potential components:
- launcher
- launcher core
- server profile
- Minecraft
- Fabric
- Java runtime
- modpack
- mods
- configs
- JVM profile
- resources

## Manifest

The launcher should receive metadata describing:
- component version
- files
- sizes
- cryptographic hashes
- dependencies
- minimum compatible launcher
- signature/authentication data
- rollback information where required

## Security

The client must verify the authenticity of the manifest before trusting it.

File hashes must be verified before installation/use.

Prevent:
- manifest substitution
- old manifest replay
- downgrade
- malicious file replacement
- path traversal

## Update transaction

Preferred model:

current/
staging/
backup/

Download
→ verify
→ stage
→ validate
→ atomically activate
→ cleanup obsolete files

A failed update must not destroy the last working version.

## Obsolete files

The manifest is the authoritative desired state for protected files.

If a protected file is no longer present in the desired state:
- identify it as obsolete
- remove safely
- document/log the operation

## Interrupted downloads

The updater should recover from:
- network interruption
- process crash
- partial download
- insufficient disk space

Use retry/backoff and resume where technically appropriate.

## Caching

Use:
- local cache
- HTTP conditional requests
- ETag/Last-Modified
- version comparison

Do not repeatedly download identical metadata/assets.

## Large files

Prefer external artifact/storage/CDN sources instead of serving all large files from the lightweight backend.
