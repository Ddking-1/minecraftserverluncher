// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher core. HTTP client for the backend auth API.
// Replaces Prism's Mojang/MSA auth flow entirely.
#pragma once

#include <QObject>
#include <QString>
#include <functional>

#include "core/InstanceTypes.h"

class QNetworkAccessManager;
class QNetworkReply;

// Result of an auth request against the backend.
struct AuthResult {
    bool success = false;
    QString error; // human-readable error message on failure
    QString accessToken;
    QString refreshToken;
    int expiresIn = 0; // seconds until the access token expires
};

// Async client for the backend auth endpoints (server/api/handlers.go):
//   POST /api/v1/auth/register
//   POST /api/v1/auth/login
//   POST /api/v1/auth/refresh
//   POST /api/v1/auth/logout
//   POST /api/v1/auth/logout-all
class BackendAuthClient : public QObject
{
    Q_OBJECT
public:
    using Callback = std::function<void(const AuthResult &)>;

    explicit BackendAuthClient(QObject *parent = nullptr);

    // Base URL of the backend, e.g. "https://api.example.com".
    [[nodiscard]]
    QString backendUrl() const;
    void setBackendUrl(const QString &url);

    void registerAccount(const QString &username, const QString &password,
                         const QString &deviceId, const QString &deviceName, Callback cb);
    void login(const QString &username, const QString &password,
               const QString &deviceId, const QString &deviceName, Callback cb);
    void refresh(const QString &refreshToken, Callback cb);
    void logout(const QString &refreshToken, Callback cb);
    void logoutAll(const QString &accessToken, Callback cb);
    // Removed startServer as it's no longer used in the player launcher client.

private:
    void post(const QString &path, const QByteArray &jsonBody, const QString &bearerToken,
              Callback cb);
    void handleReply(QNetworkReply *reply, Callback cb);

    QString m_backendUrl;
    QNetworkAccessManager *m_network;
};
