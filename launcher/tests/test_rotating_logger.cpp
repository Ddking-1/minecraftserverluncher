// SPDX-License-Identifier: GPL-3.0-only
// Standalone verification for the rotating file logger.
#include <iostream>

#include <QFile>
#include <QTemporaryDir>

#include "core/RotatingFileLogger.h"

namespace {

bool check(bool cond, const char *what)
{
    if (!cond)
        std::cerr << "FAIL: " << what << std::endl;
    return cond;
}

} // namespace

int main()
{
    using namespace RotatingFileLogger;

    QTemporaryDir dir;
    if (!check(dir.isValid(), "temporary dir"))
        return 1;
    const QString logFile = dir.filePath(QStringLiteral("launcher.log"));

    // Small threshold forces rotation on the second write.
    if (!check(log(logFile, QStringLiteral("first line"), 8, 2), "first write"))
        return 1;
    if (!check(log(logFile, QStringLiteral("second line"), 8, 2), "second write rotates"))
        return 1;

    if (!check(QFile::exists(logFile + QStringLiteral(".1")), "rotated .1 exists"))
        return 1;

    // Retention cap: only maxFiles rotated files are kept.
    for (int i = 0; i < 5; ++i)
        log(logFile, QStringLiteral("line %1").arg(i), 1, 2);

    if (!check(QFile::exists(logFile + QStringLiteral(".1")), "retained .1"))
        return 1;
    if (!check(QFile::exists(logFile + QStringLiteral(".2")), "retained .2"))
        return 1;
    if (!check(!QFile::exists(logFile + QStringLiteral(".3")), "capped at .2"))
        return 1;

    std::cout << "OK" << std::endl;
    return 0;
}
