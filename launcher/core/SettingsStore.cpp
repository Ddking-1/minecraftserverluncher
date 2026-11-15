// SPDX-License-Identifier: GPL-3.0-only
#include "SettingsStore.h"

#include <QChar>
#include <QDir>
#include <QFile>
#include <QFileInfo>
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

bool isSafeJvmArg(QStringView arg)
{
    if (arg.isEmpty() || arg.size() > 1024)
        return false;
    if (arg != arg.trimmed())
        return false;
    for (const QChar &c : arg) {
        if (c.isNull() || c.category() == QChar::Other_Control)
            return false;
    }
    return true;
}

bool validate(const Settings &s, QString *error)
{
    if (s.minMemoryMb < 512) {
        if (error)
            *error = QStringLiteral("minimum memory must be at least 512 MB");
        return false;
    }
    if (s.maxMemoryMb < s.minMemoryMb) {
        if (error)
            *error = QStringLiteral("maximum memory must not be below minimum memory");
        return false;
    }
    if (s.maxMemoryMb > 65536) {
        if (error)
            *error = QStringLiteral("maximum memory must not exceed 65536 MB");
        return false;
    }
    for (const QString &arg : s.jvmArgs) {
        if (!isSafeJvmArg(arg)) {
            if (error)
                *error = QStringLiteral("unsafe JVM argument: %1").arg(arg);
            return false;
        }
    }
    return true;
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
    s.backendUrl = obj.value(QStringLiteral("backend_url")).toString();
    s.username = obj.value(QStringLiteral("username")).toString();

    const QJsonArray args = obj.value(QStringLiteral("jvm_args")).toArray();
    for (const QJsonValue &v : args) {
        const QString arg = v.toString();
        if (isSafeJvmArg(arg))
            s.jvmArgs.append(arg);
    }

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
    obj.insert(QStringLiteral("backend_url"), s.backendUrl);
    obj.insert(QStringLiteral("username"), s.username);

    QJsonArray args;
    for (const QString &a : s.jvmArgs)
        args.append(a);
    obj.insert(QStringLiteral("jvm_args"), args);

    const QString parentDir = QFileInfo(path).absolutePath();
    if (!parentDir.isEmpty())
        QDir().mkpath(parentDir);

    QSaveFile file(path);
    if (!file.open(QIODevice::WriteOnly))
        return false;
    const QByteArray data = QJsonDocument(obj).toJson(QJsonDocument::Indented);
    if (file.write(data) != data.size())
        return false;
    return file.commit();
}

} // namespace SettingsStore
