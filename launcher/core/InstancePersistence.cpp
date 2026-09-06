// SPDX-License-Identifier: GPL-3.0-only
#include "InstancePersistence.h"

#include <QFile>
#include <QJsonArray>
#include <QJsonDocument>
#include <QSaveFile>

namespace InstancePersistence {

QJsonObject metadataToJson(const InstanceMetadata &meta)
{
    QJsonObject obj;
    obj.insert(QStringLiteral("id"), meta.id);
    obj.insert(QStringLiteral("uuid"), meta.uuid);
    obj.insert(QStringLiteral("name"), meta.name);
    obj.insert(QStringLiteral("type"), meta.type);

    QJsonArray traits;
    for (const QString &t : meta.traits)
        traits.append(t);
    obj.insert(QStringLiteral("traits"), traits);

    obj.insert(QStringLiteral("last_launch"), meta.lastLaunchMs);
    obj.insert(QStringLiteral("total_time_played"), meta.totalTimePlayedMs);
    return obj;
}

InstanceMetadata metadataFromJson(const QJsonObject &obj)
{
    InstanceMetadata meta;
    meta.id = obj.value(QStringLiteral("id")).toString();
    meta.uuid = obj.value(QStringLiteral("uuid")).toString();
    meta.name = obj.value(QStringLiteral("name")).toString();
    meta.type = obj.value(QStringLiteral("type")).toString();

    const QJsonArray traits = obj.value(QStringLiteral("traits")).toArray();
    for (const QJsonValue &v : traits)
        meta.traits.append(v.toString());

    meta.lastLaunchMs = static_cast<qint64>(obj.value(QStringLiteral("last_launch")).toDouble());
    meta.totalTimePlayedMs = static_cast<qint64>(obj.value(QStringLiteral("total_time_played")).toDouble());
    return meta;
}

QJsonObject settingsToJson(const LauncherSettings &settings)
{
    QJsonObject obj;
    obj.insert(QStringLiteral("java_path"), settings.javaPath);
    obj.insert(QStringLiteral("max_memory_mb"), settings.maxMemoryMb);
    obj.insert(QStringLiteral("min_memory_mb"), settings.minMemoryMb);

    QJsonArray jvmArgs;
    for (const QString &a : settings.jvmArgs)
        jvmArgs.append(a);
    obj.insert(QStringLiteral("jvm_args"), jvmArgs);
    return obj;
}

LauncherSettings settingsFromJson(const QJsonObject &obj)
{
    LauncherSettings settings;
    settings.javaPath = obj.value(QStringLiteral("java_path")).toString();
    settings.maxMemoryMb = obj.value(QStringLiteral("max_memory_mb")).toInt(4096);
    settings.minMemoryMb = obj.value(QStringLiteral("min_memory_mb")).toInt(2048);

    const QJsonArray jvmArgs = obj.value(QStringLiteral("jvm_args")).toArray();
    for (const QJsonValue &v : jvmArgs)
        settings.jvmArgs.append(v.toString());
    return settings;
}

bool readJsonFile(const QString &path, QJsonObject &out)
{
    QFile file(path);
    if (!file.open(QIODevice::ReadOnly))
        return false;
    const QByteArray data = file.readAll();
    file.close();

    QJsonParseError err{};
    const QJsonDocument doc = QJsonDocument::fromJson(data, &err);
    if (err.error != QJsonParseError::NoError || !doc.isObject())
        return false;
    out = doc.object();
    return true;
}

bool writeJsonFile(const QString &path, const QJsonObject &obj)
{
    QSaveFile file(path);
    if (!file.open(QIODevice::WriteOnly))
        return false;
    const QByteArray data = QJsonDocument(obj).toJson(QJsonDocument::Indented);
    if (file.write(data) != data.size())
        return false;
    return file.commit();
}

} // namespace InstancePersistence
