// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher core. Orchestrates session lifecycle: login, token
// refresh on expiry, and logout, backed by TokenStorage.
#pragma once

#include <QObject>
#include <QString>
#include <memory>
#include <functional>

#include "core/InstanceTypes.h"

class BackendAuthClient;
class TokenStorage;

// Manages the lifecycle of a backend-authenticated session. On login it stores
// the token pair; on later launches it restores and refreshes the access token
// when it has expired. Replaces Prism's AccountList/AuthFlow.
class SessionManager : public QObject
{
    Q_OBJECT
public:
    using Callback = std::function<void(bool ok, const QString &error)>;

    SessionManager(BackendAuthClient *client, TokenStorage *storage, QObject *parent = nullptr);

    // Login with username/password. On success the session is stored and
    // currentSession() is updated.
    void login(const QString &username, const QString &password,
               const QString &deviceId, const QString &deviceName, Callback cb);

    // Restore a stored session for username, refreshing the access token if
    // it has expired. On success currentSession() is updated.
    void restore(const QString &username, Callback cb);

    // Log out the current session (revokes the refresh token server-side and
    // clears local storage).
    void logout(Callback cb);

    // The current valid session, or nullptr if none.
    [[nodiscard]]
    AuthSessionPtr currentSession() const;

    // True if the access token is missing or expired.
    [[nodiscard]]
    static bool isExpired(const AuthSession &session);

private:
    void applyTokens(const QString &username, const QString &uuid,
                     const QString &accessToken, const QString &refreshToken, int expiresIn);

    BackendAuthClient *m_client;
    TokenStorage *m_storage;
    AuthSessionPtr m_session;
};
