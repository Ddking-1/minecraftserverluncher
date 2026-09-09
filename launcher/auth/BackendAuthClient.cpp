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
} // namespace

BackendAuthClient::BackendAuthClient(QObject *parent)
    : QObject(parent), m_network(new QNetworkAccessManager(this))
{
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

    if (reply->error() != QNetworkReply::NoError) {
        result.error = reply->errorString();
        if (cb)
            cb(result);
        return;
    }

    // Handle 403 Forbidden response for banned users
    if (status == 403) {
        result.error = QStringLiteral("Access denied: Account is banned");
        if (cb)
            cb(result);
        return;
    }

    QJsonParseError err{};
    const QJsonDocument doc = QJsonDocument::fromJson(data, &err);
    if (err.error != QJsonParseError::NoError || !doc.isObject()) {
        result.error = QStringLiteral("invalid response from server");
        if (cb)
            cb(result);
        return;
    }

    const QJsonObject obj = doc.object();
    if (status >= 200 && status < 300) {
        result.success = true;
        result.accessToken = obj.value(QStringLiteral("access_token")).toString();
        result.refreshToken = obj.value(QStringLiteral("refresh_token")).toString();
        result.expiresIn = obj.value(QStringLiteral("expires_in")).toInt();
    } else {
        result.error = obj.value(QStringLiteral("error")).toString();
        if (result.error.isEmpty())
            result.error = QStringLiteral("request failed (HTTP %1)").arg(status);
    }

    if (cb)
        cb(result);
}
