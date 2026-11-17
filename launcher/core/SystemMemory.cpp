// SPDX-License-Identifier: GPL-3.0-only
#include "SystemMemory.h"

#include <QtGlobal>

#if defined(Q_OS_WIN)
#  include <windows.h>
#elif defined(Q_OS_LINUX)
#  include <unistd.h>
#endif

namespace SystemMemory {

int totalPhysicalMb()
{
#if defined(Q_OS_WIN)
    MEMORYSTATUSEX status{};
    status.dwLength = sizeof(status);
    if (!GlobalMemoryStatusEx(&status))
        return 0;
    const qulonglong bytes = static_cast<qulonglong>(status.ullTotalPhys);
    if (bytes == 0)
        return 0;
    const qulonglong mb = bytes / (1024ULL * 1024ULL);
    return mb > static_cast<qulonglong>(kAbsoluteMaxMb) ? kAbsoluteMaxMb : static_cast<int>(mb);
#elif defined(Q_OS_LINUX)
    const long pages = ::sysconf(_SC_PHYS_PAGES);
    const long pageSize = ::sysconf(_SC_PAGE_SIZE);
    if (pages <= 0 || pageSize <= 0)
        return 0;
    const qulonglong bytes = static_cast<qulonglong>(pages) * static_cast<qulonglong>(pageSize);
    const qulonglong mb = bytes / (1024ULL * 1024ULL);
    return mb > static_cast<qulonglong>(kAbsoluteMaxMb) ? kAbsoluteMaxMb : static_cast<int>(mb);
#else
    return 0;
#endif
}

int allowedMaxMb(int totalMb)
{
    if (totalMb <= 0)
        return kAbsoluteMaxMb;
    return qBound(kMinMb, totalMb - kReservedMb, kAbsoluteMaxMb);
}

int allowedMaxMb()
{
    return allowedMaxMb(totalPhysicalMb());
}

int clampToAllowed(int requestedMb, int allowedMb)
{
    const int upper = qBound(kMinMb, allowedMb, kAbsoluteMaxMb);
    if (requestedMb <= 0)
        return kMinMb;
    return qBound(kMinMb, requestedMb, upper);
}

} // namespace SystemMemory
