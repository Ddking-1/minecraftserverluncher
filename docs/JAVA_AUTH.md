# JAVA AUTHENTICATION

## Two Java entry paths

### Trusted launcher

Launcher
→ trusted authorization
→ Velocity
→ server

No Minecraft `/login`.

### Normal Java

Java client
→ Velocity
→ no valid trusted authorization
→ LimboFilter
→ CAPTCHA
→ LimboAuth
→ `/login` or `/register`
→ server

## Shared account requirement

Both paths must use the same logical account system.

Launcher-created account:
must work through normal Java `/login`.

Normal Java-created account:
must work in launcher.

## Velocity responsibility

Velocity-side integration must distinguish:

1. Bedrock/Floodgate
2. trusted launcher Java
3. untrusted/normal Java

It must not infer trust from username alone.

## Limbo integration

LimboFilter/LimboAuth must remain responsible for normal Java authentication where possible.

Do not recreate mature functionality unnecessarily.

## Important edge cases

Must eventually test:
- duplicate username
- username normalization
- existing account
- launcher-created account
- normal-Java-created account
- password change
- disabled account
- session revoked
- trusted authorization expired
- direct connection without launcher
- invalid launcher proof
- replayed launcher proof

## Username rules

Username validation, normalization and Minecraft-compatible naming must be defined centrally.

Different components must not independently invent different username uniqueness rules.

## Password compatibility

If an external authentication component has a password storage format, determine whether:
- the common backend can use the same format
- a migration path is needed
- a bridge is needed
- authentication must be proxied

Do not silently maintain two passwords.
