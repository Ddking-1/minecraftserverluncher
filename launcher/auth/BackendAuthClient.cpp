// SPDX-License-Identifier: GPL-3.0-only
#include "BackendAuthClient.h"

#include <QJsonDocument>
#include <QJsonObject>
#include <QNetworkAccessManager>
#include <QNetworkReply>
#include <QNetworkRequest>
#include <QUrl>

namespace {
const char *const kRegisterPath = "/api/v1/auth/register";
const char *const kLoginPath = "/api/v1/auth/login";
const char *const kRefreshPath = "/api/v1/auth/refresh";
const char *const kLogoutPath = "/api/v1/auth/logout";
const char *const kLogoutAllPath = "/api/v1/auth/logout-all";

// Maps a backend HTTP status to a user-facing message. A server-supplied
// message always wins so specific backend errors are shown verbatim.
QString mapHttpStatusError(int status, const QString &serverMessage)
{
    if (!serverMessage.isEmpty())
        return serverMessage;
    switch (status) {
    case 400:
        return QStringLiteral("Invalid request");
    case 401:
        return QStringLiteral("Invalid or expired credentials");
    case 403:
        return QStringLiteral("Access denied");
    case 404:
        return QStringLiteral("Not found");
    case 409:
        return QStringLiteral("That username is already taken");
    case 429:
        return QStringLiteral("Too many attempts, please wait and try again");
    default:
        if (status >= 500 && status <= 599)
            return QStringLiteral("Server error, please try again later");
        return QStringLiteral("Request failed (HTTP %1)").arg(status);
    }
}
} // namespace

BackendAuthClient::BackendAuthClient(QObject *parent)
    : QObject(parent), m_network(new QNetworkAccessManager(this))
{
    m_network->setTransferTimeout(15000);
}

QString BackendAuthClient::backendUrl() const
{
    return m_backendUrl;
}

void BackendAuthClient::setBackendUrl(const QString &url)
{
    m_backendUrl = url;
    if (m_backendUrl.endsWith(QLatin1Char('/')))
        m_backendUrl.chop(1);
}

void BackendAuthClient::registerAccount(const QString &username, const QString &password,
                                        const QString &deviceId, const QString &deviceName,
                                        Callback cb)
{
    QJsonObject body;
    body.insert(QStringLiteral("username"), username);
    body.insert(QStringLiteral("password"), password);
    body.insert(QStringLiteral("device_id"), deviceId);
    body.insert(QStringLiteral("device_name"), deviceName);
    // Add HWID to the registration payload
    body.insert(QStringLiteral("hwid"), QString(QCoreApplication::applicationName() + "_" + deviceId));
    post(QLatin1String(kRegisterPath), QJsonDocument(body).toJson(QJsonDocument::Compact), {}, cb);
}

void BackendAuthClient::login(const QString &username, const QString &password,
                              const QString &deviceId, const QString &deviceName, Callback cb)
{
    QJsonObject body;
    body.insert(QStringLiteral("username"), username);
    body.insert(QStringLiteral("password"), password);
    body.insert(QStringLiteral("device_id"), deviceId);
    body.insert(QStringLiteral("device_name"), deviceName);
    // Add HWID to the login payload
    body.insert(QStringLiteral("hwid"), QString(QCoreApplication::applicationName() + "_" + deviceId));
    post(QLatin1String(kLoginPath), QJsonDocument(body).toJson(QJsonDocument::Compact), {}, cb);
}

void BackendAuthClient::refresh(const QString &refreshToken, Callback cb)
{
    QJsonObject body;
    body.insert(QStringLiteral("refresh_token"), refreshToken);
    post(QLatin1String(kRefreshPath), QJsonDocument(body).toJson(QJsonDocument::Compact), {}, cb);
}

void BackendAuthClient::logout(const QString &refreshToken, Callback cb)
{
    QJsonObject body;
    body.insert(QStringLiteral("refresh_token"), refreshToken);
    post(QLatin1String(kLogoutPath), QJsonDocument(body).toJson(QJsonDocument::Compact), {}, cb);
}

void BackendAuthClient::logoutAll(const QString &accessToken, Callback cb)
{
    post(QLatin1String(kLogoutAllPath), QByteArray(), accessToken, cb);
}



void BackendAuthClient::post(const QString &path, const QByteArray &jsonBody,
                             const QString &bearerToken, Callback cb)
{
    QNetworkRequest req(QUrl(m_backendUrl + path));
    req.setHeader(QNetworkRequest::ContentTypeHeader, QStringLiteral("application/json"));
    if (!bearerToken.isEmpty())
        req.setRawHeader("Authorization", "Bearer " + bearerToken.toUtf8());

    QNetworkReply *reply = m_network->post(req, jsonBody);
    connect(reply, &QNetworkReply::finished, this, [this, reply, cb]() { handleReply(reply, cb); });
}

void BackendAuthClient::handleReply(QNetworkReply *reply, Callback cb)
{
    reply->deleteLater();

    AuthResult result;
    const int status = reply->attribute(QNetworkRequest::HttpStatusCodeAttribute).toInt();
    const QByteArray data = reply->readAll();

    // Parse the body once; error responses carry a JSON "error" field that we
    // surface verbatim.
    QJsonParseError err{};
    const QJsonDocument doc = QJsonDocument::fromJson(data, &err);
    const bool validObject = err.error == QJsonParseError::NoError && doc.isObject();

    if (reply->error() == QNetworkReply::NoError && status >= 200 && status < 300) {
        if (!validObject) {
            result.error = QStringLiteral("invalid response from server");
            if (cb)
                cb(result);
            return;
        }
        const QJsonObject obj = doc.object();
        result.success = true;
        result.accessToken = obj.value(QStringLiteral("access_token")).toString();
        result.refreshToken = obj.value(QStringLiteral("refresh_token")).toString();
        result.expiresIn = obj.value(QStringLiteral("expires_in")).toInt();
    } else if (status == 0) {
        // Transport-level failure: no HTTP status was received.
        result.error = reply->errorString();
    } else {
        QString serverMessage;
        if (validObject)
            serverMessage = doc.object().value(QStringLiteral("error")).toString();
        result.error = mapHttpStatusError(status, serverMessage);
    }

    if (cb)
        cb(result);
}
