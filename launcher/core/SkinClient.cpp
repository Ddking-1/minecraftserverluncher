// SPDX-License-Identifier: GPL-3.0-only
#include "SkinClient.h"

#include "auth/BackendAuthClient.h"

#include <QDebug>
#include <QDir>
#include <QFile>
#include <QHttpMultiPart>
#include <QHttpPart>
#include <QNetworkAccessManager>
#include <QNetworkReply>
#include <QNetworkRequest>
#include <QTimer>

#if defined(QT_GUI_LIB) || defined(QTWidgets)
#include <QPixmap>
#endif

#include <functional>

#if !defined(QT_GUI_LIB) && !defined(QTWidgets)
class QPixmap {
public:
    QPixmap() = default;
    QPixmap(QPixmap &&) noexcept = default;
    QPixmap &operator=(QPixmap &&) noexcept = default;
    QPixmap(const QPixmap &) = delete;
    QPixmap &operator=(const QPixmap &) = delete;

    bool isNull() const { return true; }
    int width() const { return 0; }
    int height() const { return 0; }
};
#endif

namespace {

bool isValidSkinByteSize(const QByteArray &data)
{
    if (data.size() < 8)
        return false;

    return static_cast<unsigned char>(data.at(0)) == 0x89U &&
           data.at(1) == 'P' &&
           data.at(2) == 'N' &&
           data.at(3) == 'G' &&
           static_cast<const unsigned char>(data.at(4)) == 0x0DU &&
           static_cast<const unsigned char>(data.at(5)) == 0x0AU &&
           static_cast<const unsigned char>(data.at(6)) == 0x1AU &&
           static_cast<const unsigned char>(data.at(7)) == 0x0AU;
}

void logNetworkError(QNetworkReply *reply)
{
    if (!reply)
        return;

    const int status = reply->attribute(QNetworkRequest::HttpStatusCodeAttribute).toInt();
    if (reply->error() != QNetworkReply::NoError) {
        qWarning() << "SkinClient network error:" << reply->errorString();
    } else if (status >= 400) {
        qWarning() << "SkinClient HTTP error status:" << status;
    }
}

} // namespace

bool SkinClient::isSupportedSize(int width, int height)
{
    return (width == 64 && height == 64) || (width == 64 && height == 32);
}

SkinClient::SkinClient(BackendAuthClient *client, QObject *parent)
    : QObject(parent), m_client(client)
{
}

void SkinClient::uploadSkin(
    const QString &filePath,
    std::function<void(bool, QString, QString)> callback)
{
    if (m_client->backendUrl().isEmpty()) {
        invokeCallback(std::move(callback), false, QString(),
                       QStringLiteral("no backend configured"));
        return;
    }

    QFile file(filePath);
    if (!file.open(QIODevice::ReadOnly)) {
        invokeCallback(std::move(callback), false, QString(),
                       QStringLiteral("failed to open skin file"));
        return;
    }

    const QByteArray data = file.readAll();
    file.close();

    if (data.isEmpty() || !isValidSkinByteSize(data)) {
        invokeCallback(std::move(callback), false, QString(),
                       QStringLiteral("invalid skin PNG"));
        return;
    }

    QNetworkRequest request(QUrl(m_client->backendUrl() + "/api/v1/skins/upload"));
    request.setHeader(QNetworkRequest::ContentTypeHeader,
                      QStringLiteral("multipart/form-data"));

    QHttpMultiPart multiPart(QHttpMultiPart::FormDataType);
    QHttpPart skinPart;
    skinPart.setHeader(QNetworkRequest::ContentDispositionHeader,
                       QVariant("form-data; name=\"file\"; filename=\""
                                + filePath.split(QDir::separator()).last()
                                      .toUtf8()
                                      + "\""));
    skinPart.setHeader(
        QNetworkRequest::ContentTypeHeader, QStringLiteral("image/png"));
    skinPart.setBody(data);
    multiPart.append(skinPart);

    QNetworkReply *reply = m_client->m_network->post(request, &multiPart);

    QTimer *timeoutTimer = new QTimer(this);
    timeoutTimer->setSingleShot(true);
    timeoutTimer->connect(timeoutTimer, &QTimer::timeout, this,
                          [this, reply, timeoutTimer, callback]() {
                              reply->abort();
                              timeoutTimer->deleteLater();
                              logNetworkError(reply);
                              invokeCallback(std::move(callback), false, QString(),
                                             QStringLiteral("request timed out (5s)"));
                          });
    timeoutTimer->start(5000);

    reply->connect(reply, &QNetworkReply::finished, this,
                   [this, reply, timeoutTimer, callback]() {
                       timeoutTimer->stop();
                       timeoutTimer->deleteLater();
                       logNetworkError(reply);

                       const int status =
                           reply->attribute(QNetworkRequest::HttpStatusCodeAttribute)
                               .toInt();
                       if (reply->error() == QNetworkReply::NoError &&
                           status >= 200 && status < 300) {
                           invokeCallback(std::move(callback), true, QString(),
                                         QString());
                       } else {
                           invokeCallback(std::move(callback), false, QString(),
                                         QString("upload failed (HTTP %1)").arg(status));
                       }

                       reply->deleteLater();
                   });
}

void SkinClient::downloadAvatar(
    const QString &skinId,
    std::function<void(bool, QPixmap)> callback)
{
    if (m_client->backendUrl().isEmpty()) {
        invokeCallback(std::move(callback), false, QPixmap());
        return;
    }

    QNetworkRequest request(QUrl(m_client->backendUrl() +
                                 QStringLiteral("/api/v1/skins/") + skinId +
                                 QStringLiteral("/avatar")));
    request.setHeader(QNetworkRequest::ContentTypeHeader,
                      QStringLiteral("application/json"));

    QNetworkReply *reply = m_client->m_network->get(request);

    QTimer *timeoutTimer = new QTimer(this);
    timeoutTimer->setSingleShot(true);
    timeoutTimer->connect(timeoutTimer, &QTimer::timeout, this,
                          [this, reply, timeoutTimer, callback]() {
                              reply->abort();
                              timeoutTimer->deleteLater();
                              logNetworkError(reply);
                              invokeCallback(std::move(callback), false, QPixmap());
                          });
    timeoutTimer->start(5000);

    reply->connect(reply, &QNetworkReply::finished, this,
                   [this, reply, timeoutTimer, callback]() {
                       timeoutTimer->stop();
                       timeoutTimer->deleteLater();
                       logNetworkError(reply);

                       const int status =
                           reply->attribute(QNetworkRequest::HttpStatusCodeAttribute)
                               .toInt();

                       if (reply->error() != QNetworkReply::NoError ||
                           status < 200 || status >= 300) {
                           reply->deleteLater();
                           invokeCallback(std::move(callback), false, QPixmap());
                           return;
                       }

                       reply->deleteLater();
                       invokeCallback(std::move(callback), false, QPixmap());
                   });
}

void SkinClient::invokeCallback(
    std::function<void(bool, QString, QString)> cb, bool success,
    QString skinId, QString error)
{
    if (cb) {
        cb(success, std::move(skinId), std::move(error));
    }
}

void SkinClient::invokeCallback(
    std::function<void(bool, QPixmap)> cb,
    bool success, QPixmap avatar)
{
    if (cb) {
        cb(success, std::move(avatar));
    }
}
