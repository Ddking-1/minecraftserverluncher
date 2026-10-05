// SPDX-License-Identifier: GPL-3.0-only
#include <QCoreApplication>
#include <QDir>
#include <QFile>
#include <QTemporaryDir>

#include "core/GamePaths.h"

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
    QTemporaryDir dir;
    check(dir.isValid(), "temp dir");

    const QString libDir = dir.path() + QStringLiteral("/libraries");
    check(QDir().mkpath(libDir), "mkpath libraries");

    const QString goodJar = libDir + QStringLiteral("/mod-1.0.jar");
    {
        QFile f(goodJar);
        f.open(QIODevice::WriteOnly);
        f.write("x");
    }

    // A jar inside libraries is accepted.
    const QString cp = GamePaths::buildClasspath(libDir, QStringList() << goodJar);
    check(cp.contains(goodJar), "in-dir jar included");

    // An outside jar is rejected (path escape).
    const QString outsideJar = dir.path() + QStringLiteral("/outside.jar");
    {
        QFile f(outsideJar);
        f.open(QIODevice::WriteOnly);
        f.write("x");
    }
    const QString cp2 = GamePaths::buildClasspath(libDir, QStringList() << outsideJar);
    check(!cp2.contains(outsideJar), "outside jar excluded");

    // Non-jar and missing entries are rejected.
    const QString txt = libDir + QStringLiteral("/readme.txt");
    QFile f2(txt); f2.open(QIODevice::WriteOnly); f2.write("x");
    const QString cp3 = GamePaths::buildClasspath(libDir, QStringList() << txt << (libDir + QStringLiteral("/missing.jar")));
    check(cp3.isEmpty() || !cp3.contains(txt), "non-jar/missing excluded");

    // Natives dir canonicalization rejects escapes.
    check(GamePaths::safeNativesDir(libDir, QStringLiteral("bin/natives")) == (libDir + QStringLiteral("/bin/natives")),
          "valid natives dir resolved");
    check(GamePaths::safeNativesDir(libDir, QStringLiteral("../escape")).isEmpty(), ".. escape rejected");
    check(GamePaths::safeNativesDir(libDir, QStringLiteral("C:/absolute")).isEmpty(), "absolute rejected");

    if (failures == 0)
        std::printf("game_paths: all checks passed\n");
    return failures == 0 ? 0 : 1;
}
