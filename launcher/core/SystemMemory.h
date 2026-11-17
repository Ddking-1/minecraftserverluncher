// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher core. Physical memory facts used to keep the RAM
// setting (and the JVM heap derived from it) inside what the machine can take.
#pragma once

namespace SystemMemory {

// Never allow a heap below this, the JVM barely starts under it.
constexpr int kMinMb = 512;
// Hard ceiling regardless of installed RAM (32-bit/legacy JVM safety).
constexpr int kAbsoluteMaxMb = 65536;
// Memory the launcher keeps free for Windows/other programs. This is why a
// machine with 8 GB reports "6 GB allowed".
constexpr int kReservedMb = 2048;

// Installed physical RAM in MB. Returns 0 when it cannot be detected, in which
// case the caller must fall back to kAbsoluteMaxMb (never to "no limit" logic
// that would produce a negative bound).
[[nodiscard]] int totalPhysicalMb();

// Highest heap allowed on a machine with `totalMb` of RAM:
// total - kReservedMb, clamped into [kMinMb, kAbsoluteMaxMb].
// An unknown total (<= 0) yields kAbsoluteMaxMb.
[[nodiscard]] int allowedMaxMb(int totalMb);

// allowedMaxMb(totalPhysicalMb()).
[[nodiscard]] int allowedMaxMb();

// Clamp `requestedMb` into [kMinMb, allowedMb]. Non-positive input means
// "unset" and becomes kMinMb.
[[nodiscard]] int clampToAllowed(int requestedMb, int allowedMb);

} // namespace SystemMemory
