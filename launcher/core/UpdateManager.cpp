// SPDX-License-Identifier: GPL-3.0-only
#include "UpdateManager.h"

#include "UpdateManagerHMAC.h"

#include "FriendlyError.h"

#include <QDebug>
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
    m_backend = backendBase;
    if (m_backend.endsWith(QLatin1Char('/')))
        m_backend.chop(1);
    m_current = currentVersion;
}

void UpdateManager::setSigningKey(QStringView key)
{
    m_signingKey = key.toString();
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
            emit checkFailed(FriendlyError::explain(reply->errorString()));
            reply->deleteLater();
            return;
        }
        const QByteArray data = reply->readAll();
        reply->deleteLater();

        QJsonParseError err{};
        const QJsonDocument doc = QJsonDocument::fromJson(data, &err);
        if (err.error != QJsonParseError::NoError || !doc.isObject()) {
            emit checkFailed(FriendlyError::explain(QStringLiteral("invalid manifest JSON")));
            return;
        }
        const QJsonObject root = doc.object();
        const QJsonValue u = root.value(QStringLiteral("update"));
        if (!u.isObject()) {
            emit checkFailed(FriendlyError::explain(QStringLiteral("missing update object")));
            return;
        }
        const QJsonObject obj = u.toObject();
        const QString version = obj.value(QStringLiteral("version")).toString();
        const QString url = obj.value(QStringLiteral("url")).toString();
        const QString sha256 = obj.value(QStringLiteral("sha256")).toString();
        const bool mandatory = obj.value(QStringLiteral("mandatory")).toBool(false);
        const QString signature = obj.value(QStringLiteral("signature")).toString();

        // Verify the manifest HMAC-SHA256 signature before trusting it.
        if (!verifyManifestSignature(m_signingKey, version, url, sha256, mandatory, signature)) {
            // Never tell the player to install an unverified update: log the
            // technical reason and explain the safe action instead.
            qWarning() << "UpdateManager: manifest signature verification failed";
            emit checkFailed(
                FriendlyError::explain(QStringLiteral("manifest signature verification failed")));
            return;
        }

        // The caller decides whether an update is needed; this always reports
        // whatever the backend says (even if equal to the current version).
        emit updateInfo(version, url, sha256, mandatory);
    });
}