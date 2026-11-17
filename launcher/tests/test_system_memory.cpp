// SPDX-License-Identifier: GPL-3.0-only
// The RAM hint and the pre-launch memory guard share these rules: a machine
// with 8 GB reports 6 GB as allowed, and nothing can exceed the machine.
#include <QCoreApplication>

#include "core/SystemMemory.h"

#include <cstdio>

static int failures = 0;
static void check(bool cond, const char *msg)
{
    if (!cond) {
        std::fprintf(stderr, "FAIL: %s\n", msg);
        ++failures;
    }
}

int main(int argc, char **argv)
{
    QCoreApplication app(argc, argv);

    // 8 GB machine -> 6 GB allowed (2 GB kept for the system).
    check(SystemMemory::allowedMaxMb(8192) == 6144, "8 GB total allows 6 GB");
    check(SystemMemory::allowedMaxMb(16384) == 14336, "16 GB total allows 14 GB");
    check(SystemMemory::allowedMaxMb(4096) == 2048, "4 GB total allows 2 GB");

    // Tiny and absurd totals never produce a negative or unbounded value.
    check(SystemMemory::allowedMaxMb(2048) == SystemMemory::kMinMb, "2 GB total floors at min");
    check(SystemMemory::allowedMaxMb(1024) == SystemMemory::kMinMb, "1 GB total floors at min");
    check(SystemMemory::allowedMaxMb(0) == SystemMemory::kAbsoluteMaxMb, "unknown total -> absolute max");
    check(SystemMemory::allowedMaxMb(-1) == SystemMemory::kAbsoluteMaxMb, "negative total -> absolute max");
    check(SystemMemory::allowedMaxMb(1024 * 1024) == SystemMemory::kAbsoluteMaxMb,
          "huge total capped at absolute max");

    // Clamping the selected value.
    check(SystemMemory::clampToAllowed(4096, 6144) == 4096, "value under limit kept");
    check(SystemMemory::clampToAllowed(12288, 6144) == 6144, "value over limit capped");
    check(SystemMemory::clampToAllowed(0, 6144) == SystemMemory::kMinMb, "zero becomes min");
    check(SystemMemory::clampToAllowed(-4096, 6144) == SystemMemory::kMinMb, "negative becomes min");
    check(SystemMemory::clampToAllowed(8192, 0) == SystemMemory::kMinMb, "zero limit becomes min");

    // Detection must not crash and must stay in a sane range.
    const int total = SystemMemory::totalPhysicalMb();
    check(total >= 0 && total <= SystemMemory::kAbsoluteMaxMb, "detected total in range");
    check(SystemMemory::allowedMaxMb() >= SystemMemory::kMinMb, "allowed max is at least the minimum");

    if (failures == 0)
        std::printf("system_memory: all checks passed\n");
    return failures == 0 ? 0 : 1;
}
