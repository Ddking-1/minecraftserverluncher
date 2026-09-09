# PTERODACTYL AUTH FINAL ARCHITECTURE

This document describes the final architecture of the authentication system for Minecraft launcher that integrates with Pterodactyl server management platform. This integration involves sending HWID (Hardware Identifier) information from a C++ Qt client to a Go backend.

## Overview

The authentication flow is designed as follows:
1. Client-side application connects to the backend via HTTP requests
2. Authentication data (username/password, device ID, and now HWID) is sent to the Go backend for verification
3. The Go backend checks if the account is banned or not by consulting its database
4. If banned, it returns a 403 Forbidden response that is properly handled by the Qt client

## Client-side Implementation Details (C++)

The `BackendAuthClient` class in `launcher/auth/BackendAuthClient.cpp` and `.h` files handles all authentication requests.
- The `login()` function now includes the HWID field with a constructed identifier like `applicationName_deviceId`
- The same logic is applied to `registerAccount()`
- Error handling has been improved to detect 403 Forbidden responses from backend

## Backend Implementation Details (Go)

The Go server in `/server/api/handlers.go` handles incoming authentication requests:
1. Parses the payload, including new HWID field
2. Verifies if the user is banned using database lookup
3. Returns appropriate HTTP status codes and messages for successful logins or errors

## Integration with Pterodactyl

- The launcher sends hardware identifiers to the backend service which can then validate against allowed devices
- If a device is marked as banned (e.g., due to cheating), it will block access entirely
- This prevents unauthorized use of the launcher with banned accounts

This architecture provides:
1. Device tracking for security purposes
2. Hardware-based access control
3. Integration with Pterodactyl's server management platform
4. Proper error handling and feedback to users when their account is blocked