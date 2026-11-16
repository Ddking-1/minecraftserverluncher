// SPDX-License-Identifier: GPL-3.0-only
#include "SkinClient.h"

#include "auth/BackendAuthClient.h"

#include <QDebug>
#include <QFile>
#include <QFileInfo>
#include <QHttpMultiPart>
#include <QHttpPart>
#include <QJsonDocument>
#include <QJsonObject>
#include <QNetworkAccessManager>
#include <QNetworkReply>
#include <QNetworkRequest>
#include <QTimer>

#if defined(QT_GUI_LIB)
#include <QPixmap>
#endif

#include <functional>
#include <memory>

#if !defined(QT_GUI_LIB)
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
           static_cast<unsigned char>(data.at(4)) == 0x0DU &&
           static_cast<unsigned char>(data.at(5)) == 0x0AU &&
           static_cast<unsigned char>(data.at(6)) == 0x1AU &&
           static_cast<unsigned char>(data.at(7)) == 0x0AU;
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
    if (m_client == nullptr || m_client->backendUrl().isEmpty()) {
        invokeCallback(std::move(callback), false, QString(),
                       QStringLiteral("no backend configured"));
        return;
    }
    const QString token = m_client->accessToken();
    if (token.isEmpty()) {
        invokeCallback(std::move(callback), false, QString(),
                       QStringLiteral("not logged in"));
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

    QNetworkRequest request(QUrl(m_client->backendUrl() +
                                 QStringLiteral("/api/v1/skin-files")));
    request.setRawHeader("Authorization", "Bearer " + token.toUtf8());

    // QNetworkAccessManager takes ownership of the QHttpMultiPart together with
    // the reply, so it must live on the heap: a stack object would be destroyed
    // while the request body is still being read.
    auto *multiPart = new QHttpMultiPart(QHttpMultiPart::FormDataType);
    QHttpPart skinPart;
    const QString fileName = QFileInfo(filePath).fileName();
    skinPart.setHeader(
        QNetworkRequest::ContentDispositionHeader,
        QVariant(QStringLiteral("form-data; name=\"skin\"; filename=\"%1\"").arg(fileName)));
    skinPart.setHeader(QNetworkRequest::ContentTypeHeader, QStringLiteral("image/png"));
    skinPart.setBody(data);
    multiPart->append(skinPart);

    QNetworkReply *reply = m_client->networkAccessManager()->post(request, multiPart);
    if (reply == nullptr) {
        delete multiPart;
        invokeCallback(std::move(callback), false, QString(),
                       QStringLiteral("could not start upload"));
        return;
    }

    // Guards against delivering both the timeout and the finished callback for
    // the same request (abort() also emits finished()).
    auto settled = std::make_shared<bool>(false);

    QTimer *timeoutTimer = new QTimer(this);
    timeoutTimer->setSingleShot(true);
    timeoutTimer->connect(timeoutTimer, &QTimer::timeout, this,
                          [this, reply, timeoutTimer, callback, settled]() {
                              if (*settled)
                                  return;
                              *settled = true;
                              reply->abort();
                              timeoutTimer->deleteLater();
                              logNetworkError(reply);
                              invokeCallback(std::move(callback), false, QString(),
                                             QStringLiteral("request timed out (5s)"));
                          });
    timeoutTimer->start(5000);

    reply->connect(reply, &QNetworkReply::finished, this,
                   [this, reply, timeoutTimer, callback, settled]() {
                       timeoutTimer->stop();
                       timeoutTimer->deleteLater();
                       if (*settled) {
                           reply->deleteLater();
                           return;
                       }
                       *settled = true;
                       logNetworkError(reply);

                       const int status =
                           reply->attribute(QNetworkRequest::HttpStatusCodeAttribute)
                               .toInt();
                       const QByteArray body = reply->readAll();
                       if (reply->error() == QNetworkReply::NoError &&
                           status >= 200 && status < 300) {
                           QString skinId;
                           const QJsonDocument doc = QJsonDocument::fromJson(body);
                           if (doc.isObject())
                               skinId = doc.object().value(QStringLiteral("skin_id")).toString();
                           invokeCallback(std::move(callback), true, skinId, QString());
                       } else {
                           invokeCallback(std::move(callback), false, QString(),
                                          QStringLiteral("upload failed (HTTP %1)").arg(status));
                       }

                       reply->deleteLater();
                   });
}

void SkinClient::downloadAvatar(
    const QString &skinId,
    std::function<void(bool, QPixmap)> callback)
{
    if (m_client == nullptr || m_client->backendUrl().isEmpty()) {
        invokeCallback(std::move(callback), false, QPixmap());
        return;
    }
    const QString token = m_client->accessToken();
    if (token.isEmpty()) {
        invokeCallback(std::move(callback), false, QPixmap());
        return;
    }
    if (skinId.isEmpty()) {
        invokeCallback(std::move(callback), false, QPixmap());
        return;
    }

    QNetworkRequest request(QUrl(m_client->backendUrl() +
                                 QStringLiteral("/api/v1/skin-files/") + skinId +
                                 QStringLiteral("/avatar")));
    request.setRawHeader("Authorization", "Bearer " + token.toUtf8());

    QNetworkReply *reply = m_client->networkAccessManager()->get(request);
    if (reply == nullptr) {
        invokeCallback(std::move(callback), false, QPixmap());
        return;
    }

    auto settled = std::make_shared<bool>(false);

    QTimer *timeoutTimer = new QTimer(this);
    timeoutTimer->setSingleShot(true);
    timeoutTimer->connect(timeoutTimer, &QTimer::timeout, this,
                          [this, reply, timeoutTimer, callback, settled]() {
                              if (*settled)
                                  return;
                              *settled = true;
                              reply->abort();
                              timeoutTimer->deleteLater();
                              logNetworkError(reply);
                              invokeCallback(std::move(callback), false, QPixmap());
                          });
    timeoutTimer->start(5000);

    reply->connect(reply, &QNetworkReply::finished, this,
                   [this, reply, timeoutTimer, callback, settled]() {
                       timeoutTimer->stop();
                       timeoutTimer->deleteLater();
                       if (*settled) {
                           reply->deleteLater();
                           return;
                       }
                       *settled = true;
                       logNetworkError(reply);

                       const int status =
                           reply->attribute(QNetworkRequest::HttpStatusCodeAttribute)
                               .toInt();
                       const bool networkOk = reply->error() == QNetworkReply::NoError;
                       const QByteArray body = reply->readAll();
                       reply->deleteLater();

                       if (!networkOk || status < 200 || status >= 300) {
                           invokeCallback(std::move(callback), false, QPixmap());
                           return;
                       }

#if defined(QT_GUI_LIB)
                       QPixmap pixmap;
                       if (!pixmap.loadFromData(body, "PNG") || pixmap.isNull()) {
                           invokeCallback(std::move(callback), false, QPixmap());
                           return;
                       }
                       invokeCallback(std::move(callback), true, pixmap);
#else
                       Q_UNUSED(body);
                       invokeCallback(std::move(callback), false, QPixmap());
#endif
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
