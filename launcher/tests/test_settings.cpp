// SPDX-License-Identifier: GPL-3.0-only
// Standalone verification for InstancePersistence round-tripping.
#include <iostream>

#include <QJsonObject>
#include <QString>
#include <QTemporaryDir>

#include "core/InstancePersistence.h"

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
    // Settings round-trip.
    LauncherSettings settings;
    settings.javaPath = QStringLiteral("/usr/bin/java");
    settings.maxMemoryMb = 8192;
    settings.minMemoryMb = 1024;
    settings.jvmArgs = {QStringLiteral("-Xmx8G"), QStringLiteral("-XX:+UseG1GC")};

    const LauncherSettings loaded = InstancePersistence::settingsFromJson(
        InstancePersistence::settingsToJson(settings));
    if (!check(loaded.javaPath == settings.javaPath, "settings javaPath")
        || !check(loaded.maxMemoryMb == settings.maxMemoryMb, "settings maxMemoryMb")
        || !check(loaded.minMemoryMb == settings.minMemoryMb, "settings minMemoryMb")
        || !check(loaded.jvmArgs == settings.jvmArgs, "settings jvmArgs"))
        return 1;

    // Metadata round-trip.
    InstanceMetadata meta;
    meta.id = QStringLiteral("inst-1");
    meta.uuid = QStringLiteral("uuid-1");
    meta.name = QStringLiteral("My Instance");
    meta.type = QStringLiteral("MinecraftInstance");
    meta.traits = {QStringLiteral("a"), QStringLiteral("b")};
    meta.lastLaunchMs = 123;
    meta.totalTimePlayedMs = 456;

    const InstanceMetadata loadedMeta = InstancePersistence::metadataFromJson(
        InstancePersistence::metadataToJson(meta));
    if (!check(loadedMeta.id == meta.id, "metadata id")
        || !check(loadedMeta.uuid == meta.uuid, "metadata uuid")
        || !check(loadedMeta.name == meta.name, "metadata name")
        || !check(loadedMeta.type == meta.type, "metadata type")
        || !check(loadedMeta.traits == meta.traits, "metadata traits")
        || !check(loadedMeta.lastLaunchMs == meta.lastLaunchMs, "metadata lastLaunchMs")
        || !check(loadedMeta.totalTimePlayedMs == meta.totalTimePlayedMs, "metadata totalTimePlayedMs"))
        return 1;

    // File round-trip.
    QTemporaryDir dir;
    if (!check(dir.isValid(), "temporary dir"))
        return 1;
    const QString path = dir.filePath(QStringLiteral("settings.json"));
    if (!check(InstancePersistence::writeJsonFile(
                   path, InstancePersistence::settingsToJson(settings)),
               "writeJsonFile"))
        return 1;

    QJsonObject out;
    if (!check(InstancePersistence::readJsonFile(path, out), "readJsonFile"))
        return 1;
    const LauncherSettings fromFile = InstancePersistence::settingsFromJson(out);
    if (!check(fromFile.maxMemoryMb == 8192, "file maxMemoryMb"))
        return 1;

    std::cout << "OK" << std::endl;
    return 0;
}
