// SPDX-License-Identifier: GPL-3.0-only
#include "TokenStorage.h"

#include <QDir>
#include <QFile>
#include <QFileInfo>
#include <QJsonDocument>
#include <QJsonObject>
#include <QSaveFile>

TokenStorage::TokenStorage(const QString &dataDir) : m_dataDir(dataDir)
{
}

QString TokenStorage::filePathFor(const QString &username) const
{
    return m_dataDir + QStringLiteral("/tokens/") + username + QStringLiteral(".json");
}

bool TokenStorage::save(const AuthSession &session) const
{
    const QString path = filePathFor(session.username);
    QDir().mkpath(QFileInfo(path).absolutePath());

    QJsonObject obj;
    obj.insert(QStringLiteral("username"), session.username);
    obj.insert(QStringLiteral("uuid"), session.uuid);
    obj.insert(QStringLiteral("access_token"), session.access_token);
    obj.insert(QStringLiteral("refresh_token"), session.refresh_token);
    obj.insert(QStringLiteral("expires_at"), session.expires_at);
    obj.insert(QStringLiteral("is_online"), session.is_online);

    QSaveFile file(path);
    if (!file.open(QIODevice::WriteOnly))
        return false;
    const QByteArray data = QJsonDocument(obj).toJson(QJsonDocument::Indented);
    if (file.write(data) != data.size())
        return false;
    return file.commit();
}

bool TokenStorage::load(const QString &username, AuthSession &out) const
{
    QFile file(filePathFor(username));
    if (!file.open(QIODevice::ReadOnly))
        return false;
    const QByteArray data = file.readAll();
    file.close();

    QJsonParseError err{};
    const QJsonDocument doc = QJsonDocument::fromJson(data, &err);
    if (err.error != QJsonParseError::NoError || !doc.isObject())
        return false;

    const QJsonObject obj = doc.object();
    out.username = obj.value(QStringLiteral("username")).toString();
    out.uuid = obj.value(QStringLiteral("uuid")).toString();
    out.access_token = obj.value(QStringLiteral("access_token")).toString();
    out.refresh_token = obj.value(QStringLiteral("refresh_token")).toString();
    out.expires_at = static_cast<qint64>(obj.value(QStringLiteral("expires_at")).toDouble());
    out.is_online = obj.value(QStringLiteral("is_online")).toBool();
    return !out.username.isEmpty();
}

void TokenStorage::remove(const QString &username) const
{
    QFile::remove(filePathFor(username));
}
