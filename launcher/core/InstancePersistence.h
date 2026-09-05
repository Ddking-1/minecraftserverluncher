// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher core. JSON persistence for instance metadata.
#pragma once

#include <QJsonObject>
#include <QString>
#include <QStringList>

// Plain data mirroring the on-disk instance.json / launcher.json files.
// Kept free of QObject so it can be serialized/deserialized in isolation.
struct InstanceMetadata
{
    QString id;
    QString uuid;
    QString name;
    QString type; // e.g. "MinecraftInstance"
    QStringList traits;
    qint64 lastLaunchMs = 0;
    qint64 totalTimePlayedMs = 0;
};

struct LauncherSettings
{
    QString javaPath;
    int maxMemoryMb = 4096;
    int minMemoryMb = 2048;
    QStringList jvmArgs;
};

namespace InstancePersistence {

// Serialize / deserialize instance metadata.
QJsonObject metadataToJson(const InstanceMetadata &meta);
InstanceMetadata metadataFromJson(const QJsonObject &obj);

// Serialize / deserialize launcher settings.
QJsonObject settingsToJson(const LauncherSettings &settings);
LauncherSettings settingsFromJson(const QJsonObject &obj);

// Read a JSON object from a file. Returns false if the file is missing or
// not valid JSON.
bool readJsonFile(const QString &path, QJsonObject &out);

// Atomically write a JSON object to a file (temp file + rename).
bool writeJsonFile(const QString &path, const QJsonObject &obj);

} // namespace InstancePersistence
