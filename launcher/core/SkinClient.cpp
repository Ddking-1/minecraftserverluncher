// SPDX-License-Identifier: GPL-3.0-only
#include "SkinClient.h"

#include "auth/BackendAuthClient.h"

#include <QFile>
#include <QFileInfo>
#include <QNetworkAccessManager>
#include <QNetworkReply>
#include <QNetworkRequest>
#include <QUrl>

SkinClient::SkinClient(BackendAuthClient *client, QObject *parent)
    : QObject(parent), m_client(client)
{
}

bool SkinClient::isSupportedSize(int width, int height)
{
    return (width == 64 && height == 64) || (width == 64 && height == 32);
}

void SkinClient::uploadSkin(const QString &accessToken, const QString &filePath,
                            const QString &variant)
{
    if (m_busy) {
        emit uploadFailed(QStringLiteral("an upload is already in progress"));
        return;
    }
    if (!m_client) {
        emit uploadFailed(QStringLiteral("no backend client configured"));
        return;
    }
    if (accessToken.isEmpty()) {
        emit uploadFailed(QStringLiteral("not logged in"));
        return;
    }

    QFile file(filePath);
    if (!file.open(QIODevice::ReadOnly)) {
        emit uploadFailed(QStringLiteral("could not open skin file"));
        return;
    }
    const QByteArray data = file.readAll();
    file.close();
    if (data.isEmpty()) {
        emit uploadFailed(QStringLiteral("skin file is empty"));
        return;
    }

    // The backend sniffs the real bytes; the client sends a stable type and
    // lets the server reject anything that is not a valid PNG.
    QNetworkRequest req(QUrl(m_client->backendUrl() + QStringLiteral("/api/v1/skins/") + variant));
    req.setHeader(QNetworkRequest::ContentTypeHeader, QStringLiteral("image/png"));
    req.setRawHeader("Authorization", "Bearer " + accessToken.toUtf8());

    m_busy = true;
    QNetworkAccessManager *net = m_client->findChild<QNetworkAccessManager *>();
    if (!net) {
        net = new QNetworkAccessManager(m_client);
    }
    QNetworkReply *reply = net->put(req, data);
    connect(reply, &QNetworkReply::finished, this, [this, reply, variant]() {
        reply->deleteLater();
        m_busy = false;
        const int status = reply->attribute(QNetworkRequest::HttpStatusCodeAttribute).toInt();
        if (reply->error() == QNetworkReply::NoError && status >= 200 && status < 300) {
            emit uploadSucceeded(variant);
        } else if (status == 429) {
            emit uploadFailed(QStringLiteral("upload limit reached (max 2 per 3 hours)"));
        } else if (status == 400 || status == 415) {
            emit uploadFailed(QStringLiteral("invalid skin (must be a 64x64 or 64x32 PNG)"));
        } else {
            emit uploadFailed(QStringLiteral("upload failed (HTTP %1)").arg(status));
        }
    });
}
