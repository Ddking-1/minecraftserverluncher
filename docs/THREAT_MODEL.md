# THREAT MODEL

## Security philosophy

The launcher is a convenience/trust client, not an absolute security boundary.

A compromised client machine is considered hostile.

The system must therefore combine:
- authentication
- device-bound credentials
- short-lived authorization
- server-side validation
- signed update manifests
- integrity checks
- server-side gameplay protection

## Threats

### Account attacks
- password guessing
- credential stuffing
- username enumeration
- stolen session
- stolen refresh credential
- session replay
- session fixation
- logout bypass
- expired-session reuse

### Device attacks
- local credential extraction
- device cloning
- token theft
- replay from another device
- compromised OS

### Network attacks
- MITM
- replay
- request tampering
- malicious redirects
- forged launch authorization

### Java bypass
- connecting directly with a normal Minecraft client
- username spoofing
- fake client brand
- fake launcher metadata
- attempting to bypass LimboFilter
- attempting to bypass LimboAuth
- reusing an expired trusted authorization

### Bedrock threats
- fake Bedrock identity
- account-link abuse
- linking code theft
- linking code replay
- connection flooding
- bot connections

### Update threats
- modified manifest
- old manifest replay
- downgrade
- corrupted file
- malicious file
- path traversal
- malicious executable replacement
- partial update corruption

### Local modification
- protected mod modification
- modified config
- modified runtime
- unknown protected files
- resource modification

## Security invariants

The following must remain true:

1. Username alone never grants trusted access.
2. A trusted launch authorization expires.
3. Revoked sessions cannot be used.
4. Expired sessions cannot be reused.
5. Private cryptographic keys are never uploaded as server-side secrets.
6. Protected update manifests are authenticated.
7. Untrusted local files are never assumed correct.
8. The Bedrock identity flow cannot accidentally become a Java-auth bypass.
9. Admin endpoints require explicit authorization.
10. Secrets are never written to ordinary logs.

## False confidence

Do not claim:
- perfect anti-cheat
- perfect anti-tamper
- impossible launcher modification
- impossible credential theft
- absolute hardware binding

Document limitations explicitly.
