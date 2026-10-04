// SPDX-License-Identifier: GPL-3.0-only
// Standalone verification for the launcher settings store.
#include <iostream>

#include <QFile>
#include <QTemporaryDir>

#include "core/SettingsStore.h"

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
    using namespace SettingsStore;

    const Settings def = defaults();
    if (!check(def.minMemoryMb == 2048, "default min")
        || !check(def.maxMemoryMb == 4096, "default max"))
        return 1;

    {
        Settings s;
        s.minMemoryMb = 100;
        s.maxMemoryMb = 50;
        clamp(s);
        if (!check(s.minMemoryMb == 512, "min clamped to 512")
            || !check(s.maxMemoryMb == 512, "max raised to min"))
            return 1;
    }

    QTemporaryDir dir;
    if (!check(dir.isValid(), "temporary dir"))
        return 1;
    const QString path = dir.filePath(QStringLiteral("launcher.json"));

    Settings out;
    out.javaPath = QStringLiteral("C:/java/bin/java.exe");
    out.minMemoryMb = 1024;
    out.maxMemoryMb = 8192;
    out.jvmArgs = QStringList{QStringLiteral("-XX:+UseG1GC"), QStringLiteral("-Dfoo=bar")};

    if (!check(save(path, out), "save succeeds"))
        return 1;

    Settings in;
    if (!check(load(path, in), "load succeeds"))
        return 1;
    if (!check(in.javaPath == out.javaPath, "round-trip javaPath")
        || !check(in.minMemoryMb == out.minMemoryMb, "round-trip min")
        || !check(in.maxMemoryMb == out.maxMemoryMb, "round-trip max")
        || !check(in.jvmArgs == out.jvmArgs, "round-trip jvmArgs"))
        return 1;

    if (!check(!load(dir.filePath(QStringLiteral("missing.json")), in), "missing file fails"))
        return 1;

    {
        QFile bad(path);
        if (!check(bad.open(QIODevice::WriteOnly), "open bad json"))
            return 1;
        bad.write("{not valid json");
        bad.close();
        if (!check(!load(path, in), "malformed file fails"))
            return 1;
    }

    std::cout << "OK" << std::endl;
    return 0;
}
