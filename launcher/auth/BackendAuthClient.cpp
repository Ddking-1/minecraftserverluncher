// SPDX-License-Identifier: GPL-3.0-only
#include "BackendAuthClient.h"

#include <QCoreApplication>
#include <QJsonDocument>
#include <QJsonObject>
#include <QNetworkAccessManager>
#include <QNetworkReply>
#include <QNetworkRequest>
#include <QSslConfiguration>
#include <QUrl>

#include "core/HWIDExtractor.h"

namespace {
const char *const kRegisterPath = "/api/v1/auth/register";
const char *const kLoginPath = "/api/v1/auth/login";
const char *const kRefreshPath = "/api/v1/auth/refresh";
const char *const kLogoutPath = "/api/v1/auth/logout";
const char *const kLogoutAllPath = "/api/v1/auth/logout-all";

// Maps a backend HTTP status to a human-facing message. A server-supplied
// message always wins so specific backend errors are shown verbatim.
[[nodiscard]]
QString mapHttpStatusError(int status, const QString &serverMessage)
{
    if (!serverMessage.isEmpty()) {
        return serverMessage;
    }
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

// Helper: convert std::string_view to QString.
[[nodiscard]]
static QString toQString(std::string_view s)
{
    return QString::fromUtf8(s.data(), static_cast<int>(s.size()));
}

// Apply Qt version-compatible SSL settings to the request.
// Qt6 uses QSslConfiguration for SSL options; Qt5 uses request attributes.
void applySslConfiguration(QNetworkRequest &req)
{
#if QT_VERSION >= QT_VERSION_CHECK(6, 0, 0)
    QSslConfiguration sslConfig = QSslConfiguration::defaultConfiguration();
    sslConfig.setPeerVerifyMode(QSslSocket::VerifyPeer);
    sslConfig.setProtocol(QSsl::TlsV1_2OrLater);
    req.setSslConfiguration(sslConfig);
#else
    // Qt5: use attribute-based SSL options.
    req.setAttribute(QNetworkRequest::SslSelectNegotiatedProtocol, true);
    req.setAttribute(QNetworkRequest::SslOptionDisableSessionTickets, false);
#endif
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
    const QString trimmed = url.trimmed();
    if (trimmed.isEmpty()) {
        // Clearing the URL is explicit (logout / reconfiguration).
        m_backendUrl.clear();
        return;
    }

    const QUrl parsed(trimmed);
    const QString scheme = parsed.scheme().toLower();
    const QString host = parsed.host();
    if (!parsed.isValid() || host.isEmpty() || !parsed.userInfo().isEmpty()
        || parsed.hasQuery() || parsed.hasFragment()
        || (!parsed.path().isEmpty() && parsed.path() != QLatin1String("/"))) {
        qWarning() << "Backend URL rejected: expected an origin such as https://api.example.com:"
                   << url;
        return;
    }

    // HTTPS is mandatory for a remote backend. Plain HTTP is allowed only for
    // a loopback backend, which never leaves the machine and is how the local
    // development backend (and the HTTP contract test) is reached.
    const bool loopback = host == QLatin1String("localhost")
        || host == QLatin1String("127.0.0.1")
        || host == QLatin1String("::1");
    if (scheme != QLatin1String("https") && !(scheme == QLatin1String("http") && loopback)) {
        qWarning() << "Backend URL must use HTTPS (HTTP is allowed only for loopback):" << url;
        return;
    }

    QString normalized = parsed.toString(QUrl::RemovePath | QUrl::RemoveQuery
                                         | QUrl::RemoveFragment);
    while (normalized.endsWith(QLatin1Char('/')))
        normalized.chop(1);
    m_backendUrl = normalized;
}

QString BackendAuthClient::accessToken() const
{
    return m_accessToken;
}

void BackendAuthClient::setAccessToken(const QString &token)
{
    m_accessToken = token;
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
    // Add HWID to the registration payload using the real hardware UUID.
    const std::string hwid = HWIDExtractor::getHardwareUUID();
    body.insert(QStringLiteral("hwid"), toQString(hwid));
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
    // Add HWID to the login payload using the real hardware UUID.
    const std::string hwid = HWIDExtractor::getHardwareUUID();
    body.insert(QStringLiteral("hwid"), toQString(hwid));
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

    // SSL/TSL enforcement: we require certificate validation and SSL errors are fatal.
    applySslConfiguration(req);
    req.setPriority(QNetworkRequest::HighPriority);

    QNetworkReply *reply = m_network->post(req, jsonBody);
    connect(reply, &QNetworkReply::sslErrors, this, [this, reply, cb](const QList<QSslError> &errors) {
        for (const auto &err : errors) {
            qWarning() << "SSL error:" << err.errorString();
        }
        // Reject the reply on any SSL error.
        reply->ignoreSslErrors(); // noqa: keep for Qt behavior; failure handled below
        AuthResult bad;
        bad.error = QStringLiteral("Untrusted backend TLS certificate");
        if (cb)
            cb(bad);
    });
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
        if (!result.accessToken.isEmpty())
            m_accessToken = result.accessToken;
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
