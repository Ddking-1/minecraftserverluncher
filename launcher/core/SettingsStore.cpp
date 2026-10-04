// SPDX-License-Identifier: GPL-3.0-only
#include "SettingsStore.h"

#include <QFile>
#include <QJsonArray>
#include <QJsonDocument>
#include <QJsonObject>
#include <QJsonValue>
#include <QSaveFile>

namespace SettingsStore {

Settings defaults()
{
    Settings s;
    s.minMemoryMb = 2048;
    s.maxMemoryMb = 4096;
    return s;
}

void applyDefaults(Settings &s)
{
    if (s.minMemoryMb <= 0)
        s.minMemoryMb = 2048;
    if (s.maxMemoryMb <= 0)
        s.maxMemoryMb = 4096;
}

void clamp(Settings &s)
{
    if (s.minMemoryMb < 512)
        s.minMemoryMb = 512;
    if (s.maxMemoryMb < s.minMemoryMb)
        s.maxMemoryMb = s.minMemoryMb;
    if (s.maxMemoryMb > 65536)
        s.maxMemoryMb = 65536;
    if (s.minMemoryMb > s.maxMemoryMb)
        s.minMemoryMb = s.maxMemoryMb;
}

bool load(const QString &path, Settings &out)
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

    const QJsonObject obj = doc.object();
    Settings s;
    s.javaPath = obj.value(QStringLiteral("java_path")).toString();
    s.minMemoryMb = obj.value(QStringLiteral("min_memory_mb")).toInt(2048);
    s.maxMemoryMb = obj.value(QStringLiteral("max_memory_mb")).toInt(4096);

    const QJsonArray args = obj.value(QStringLiteral("jvm_args")).toArray();
    for (const QJsonValue &v : args)
        s.jvmArgs.append(v.toString());

    applyDefaults(s);
    clamp(s);
    out = s;
    return true;
}

bool save(const QString &path, const Settings &in)
{
    Settings s = in;
    applyDefaults(s);
    clamp(s);

    QJsonObject obj;
    obj.insert(QStringLiteral("java_path"), s.javaPath);
    obj.insert(QStringLiteral("min_memory_mb"), s.minMemoryMb);
    obj.insert(QStringLiteral("max_memory_mb"), s.maxMemoryMb);

    QJsonArray args;
    for (const QString &a : s.jvmArgs)
        args.append(a);
    obj.insert(QStringLiteral("jvm_args"), args);

    QSaveFile file(path);
    if (!file.open(QIODevice::WriteOnly))
        return false;
    const QByteArray data = QJsonDocument(obj).toJson(QJsonDocument::Indented);
    if (file.write(data) != data.size())
        return false;
    return file.commit();
}

} // namespace SettingsStore
