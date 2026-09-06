# HWID and Ban System Implementation

## Overview

This document describes the implementation of the Hardware Identification (HWID) module in C++ and the Ban system in Go for the Minecraft launcher project.

## C++ HWID Extractor Module (`launcher/core/HWIDExtractor`)

### Purpose
The HWID extractor component is responsible for generating a unique hardware identifier based on multiple system properties that remain consistent across different launches. This ID can be used to track devices and implement device-based bans.

### Implementation Details

#### Class: `HWIDExtractor`
- **getHardwareUUID()** - Main method that combines SMBIOS UUID, CPU ID, and primary MAC address into a SHA-256 hash
- **SMBIOS UUID Extraction**: Uses WMI (`Win32_ComputerSystemProduct`) to retrieve the system's UUID
- **CPU Identification**: Retrieves CPU vendor string using x86 `__cpuid` intrinsic 
- **MAC Address Collection**: Gets the primary network adapter MAC address via `GetAdaptersInfo`

#### Security and Fallbacks
- Implements proper COM object cleanup for WMI operations  
- Provides fallback mechanisms when system properties are not available
- Handles WMI errors gracefully with appropriate error codes

### Dependencies
- Windows API (`windows.h`, `wmiutils.h`, `iphlpapi.h`)
- OpenSSL library for SHA-256 hashing
- Microsoft Visual C++ compiler (MSVC)

## Go Ban System (`backend/models/ban.go`)

### Purpose
The ban system manages account, IP-based, and hardware-based bans for the launcher. It provides functionality to check if a user is currently banned.

### Implementation Details

#### Struct: `BanRecord`
- **ID**: Unique identifier in database 
- **AccountID**: Minecraft username or account ID that was banned
- **IPAddress**: IP address associated with the ban (optional)
- **HardwareID**: Hardware UUID used for device-based bans  
- **Reason**: Human-readable reason for the ban
- **BannedAt**: Timestamp when ban was issued
- **ExpiresAt**: Expiration timestamp; empty means permanent ban
- **BanType**: Type of ban - "ACCOUNT", "HWID", or "IP"

#### Function: `IsBanned(accID, ip, hwid string)`
- Checks if a user is currently banned by verifying against any active bans 
- Returns boolean indicating whether the user is banned and the reason
- Filters out expired bans

### Implementation Notes
The actual database query logic has been left as placeholder for now. This would typically involve:
1. Querying ban records with matching AccountID, IPAddress, or HardwareID  
2. Filtering by BanType when appropriate
3. Ensuring ExpiresAt > current time (or empty for permanent bans)

## Integration

Both modules are designed to integrate seamlessly with the existing launcher architecture and backend systems.

## Future Considerations

- Add database schema definitions for ban records 
- Implement full ban persistence layer
- Add logging for ban events