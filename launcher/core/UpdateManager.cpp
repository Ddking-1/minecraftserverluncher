// SPDX-License-Identifier: GPL-3.0-only
#include "UpdateManager.h"

#include <QJsonDocument>
#include <QJsonObject>
#include <QNetworkReply>
#include <QNetworkRequest>
#include <QUrl>

UpdateManager::UpdateManager(QObject *parent)
    : QObject(parent), m_nam(new QNetworkAccessManager(this))
{
    m_nam->setTransferTimeout(10000);
}

void UpdateManager::setEndpoints(const QString &backendBase, const QString &currentVersion)
{
    m_backend = backendBase.endsWith(QLatin1Char('/')) ? backendBase.left(backendBase.length() - 1)
                                                       : backendBase;
    m_current = currentVersion;
}

QString UpdateManager::currentVersion() const
{
    return m_current;
}

void UpdateManager::checkForUpdate()
{
    const QUrl url(m_backend + QStringLiteral("/api/v1/updates/manifest"));
    QNetworkRequest req(url);
    QNetworkReply *reply = m_nam->get(req);

    connect(reply, &QNetworkReply::finished, this, [this, reply]() {
        if (reply->error() != QNetworkReply::NoError) {
            emit checkFailed(reply->errorString());
            reply->deleteLater();
            return;
        }
        const QByteArray data = reply->readAll();
        reply->deleteLater();

        QJsonParseError err{};
        const QJsonDocument doc = QJsonDocument::fromJson(data, &err);
        if (err.error != QJsonParseError::NoError || !doc.isObject()) {
            emit checkFailed(QStringLiteral("invalid manifest JSON"));
            return;
        }
        const QJsonObject root = doc.object();
        const QJsonValue u = root.value(QStringLiteral("update"));
        if (!u.isObject()) {
            emit checkFailed(QStringLiteral("missing update object"));
            return;
        }
        const QJsonObject obj = u.toObject();
        const QString version = obj.value(QStringLiteral("version")).toString();
        const QString url = obj.value(QStringLiteral("url")).toString();
        const QString sha256 = obj.value(QStringLiteral("sha256")).toString();
        const bool mandatory = obj.value(QStringLiteral("mandatory")).toBool(false);

        // The caller decides whether an update is needed; this always reports
        // whatever the backend says (even if equal to the current version).
        emit updateInfo(version, url, sha256, mandatory);
    });
}