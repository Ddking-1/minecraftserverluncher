// SPDX-License-Identifier: GPL-3.0-only
// Standalone verification for the mod directory scanner.
#include <iostream>

#include <QDir>
#include <QFile>
#include <QTemporaryDir>

#include "minecraft/ModScanner.h"

namespace {

bool check(bool cond, const char *what)
{
    if (!cond)
        std::cerr << "FAIL: " << what << std::endl;
    return cond;
}

void touch(const QString &path)
{
    QFile f(path);
    f.open(QIODevice::WriteOnly);
    f.write("x");
    f.close();
}

} // namespace

int main()
{
    using namespace ModScanner;

    QTemporaryDir dir;
    if (!check(dir.isValid(), "temporary dir"))
        return 1;

    // Missing directory yields an empty list.
    if (!check(scan(dir.filePath(QStringLiteral("nope"))).isEmpty(), "missing dir empty"))
        return 1;

    const QString mods = dir.filePath(QStringLiteral("mods"));
    QDir().mkpath(mods);
    touch(mods + QStringLiteral("/zeta.jar"));
    touch(mods + QStringLiteral("/Alpha.jar.disabled"));
    touch(mods + QStringLiteral("/beta.jar"));
    touch(mods + QStringLiteral("/notes.txt"));

    const QList<ModEntry> entries = scan(mods);
    if (!check(entries.size() == 3, "three mods found"))
        return 1;

    if (!check(entries.at(0).displayName == QStringLiteral("Alpha"), "sorted first Alpha")
        || !check(!entries.at(0).enabled, "Alpha disabled")
        || !check(entries.at(1).displayName == QStringLiteral("beta"), "sorted second beta")
        || !check(entries.at(1).enabled, "beta enabled")
        || !check(entries.at(2).displayName == QStringLiteral("zeta"), "sorted third zeta")
        || !check(entries.at(2).enabled, "zeta enabled"))
        return 1;

    if (!check(entries.at(0).fileName == QStringLiteral("Alpha.jar.disabled"), "disabled file name"))
        return 1;

    std::cout << "OK" << std::endl;
    return 0;
}
