# VELOCITY PROTOCOL

## Purpose

Velocity is the central network gateway.

The custom integration must enforce the distinction between:

1. Bedrock/Floodgate
2. trusted custom-launcher Java
3. normal/untrusted Java

## Trusted launcher

The launcher must obtain a server-verifiable launch authorization.

Conceptual flow:

Launcher
→ Auth backend
→ device/session verification
→ short-lived launch authorization
→ Minecraft
→ Velocity verification
→ allow

## Untrusted Java

If no valid trusted authorization exists:

Java
→ Velocity
→ normal authentication path
→ LimboFilter
→ CAPTCHA
→ LimboAuth

## Bedrock

Bedrock/Floodgate should use the Bedrock path and must not be routed into Java `/login`.

## Trust rules

Never trust:
- username alone
- client brand
- arbitrary client metadata
- local config
- launcher UI state

## Authorization properties

Trusted authorization should be:
- short-lived
- server/profile bound
- associated with a valid session
- associated with device proof where appropriate
- replay resistant
- revocable where practical

## Direct connection

A player must not be able to bypass trusted-launcher requirements merely by connecting directly to Velocity with a normal client.

## Failure behavior

If verification fails:
- reject trusted path
- fall back to normal Java authentication only where safe and intended
- never silently grant trusted privileges

## Logging

Record security-relevant outcomes without logging raw credentials/tokens.
