// SPDX-License-Identifier: GPL-3.0-only
#include "SessionManager.h"

#include <QDateTime>
#include <QDebug>

#include "BackendAuthClient.h"
#include "TokenStorage.h"

SessionManager::SessionManager(BackendAuthClient *client, TokenStorage *storage, QObject *parent)
    : QObject(parent), m_client(client), m_storage(storage)
{
}

void SessionManager::login(const QString &username, const QString &password,
                           const QString &deviceId, const QString &deviceName, Callback cb)
{
    m_client->login(username, password, deviceId, deviceName,
                    [this, username, cb](const AuthResult &res) {
                        if (!res.success) {
                            if (cb)
                                cb(false, res.error);
                            return;
                        }
                        applyTokens(username, QString(), res.accessToken, res.refreshToken,
                                    res.expiresIn);
                        if (cb)
                            cb(true, QString());
                    });
}

void SessionManager::restore(const QString &username, Callback cb)
{
    AuthSession stored;
    if (!m_storage->load(username, stored)) {
        if (cb)
            cb(false, QStringLiteral("no stored session"));
        return;
    }

    if (!isExpired(stored)) {
        m_session = std::make_shared<AuthSession>(stored);
        // Restoring a still-valid session must also restore the HTTP client's
        // bearer token, otherwise API clients (SkinClient) stay unauthenticated.
        if (m_client != nullptr)
            m_client->setAccessToken(stored.access_token);
        if (cb)
            cb(true, QString());
        return;
    }

    // Access token expired: refresh using the stored refresh token.
    m_client->refresh(stored.refresh_token, [this, stored, cb](const AuthResult &res) {
        if (!res.success) {
            if (m_client != nullptr)
                m_client->setAccessToken(QString());
            if (cb)
                cb(false, res.error);
            return;
        }
        applyTokens(stored.username, stored.uuid, res.accessToken, res.refreshToken, res.expiresIn);
        if (cb)
            cb(true, QString());
    });
}

void SessionManager::logout(Callback cb)
{
    if (!m_session) {
        if (cb)
            cb(true, QString());
        return;
    }
    const QString username = m_session->username;
    const QString refreshToken = m_session->refresh_token;
    m_session.reset();
    // Drop the bearer token immediately so no request can use the ended
    // session, even if the server-side logout call fails.
    if (m_client != nullptr)
        m_client->setAccessToken(QString());

    m_client->logout(refreshToken, [this, username, cb](const AuthResult &res) {
        m_storage->remove(username);
        if (cb)
            cb(res.success, res.error);
    });
}

AuthSessionPtr SessionManager::currentSession() const
{
    return m_session;
}

bool SessionManager::isExpired(const AuthSession &session)
{
    if (session.access_token.isEmpty())
        return true;
    if (session.expires_at <= 0)
        return false; // unknown expiry: treat as valid
    return QDateTime::currentMSecsSinceEpoch() >= session.expires_at;
}

void SessionManager::applyTokens(const QString &username, const QString &uuid,
                                 const QString &accessToken, const QString &refreshToken,
                                 int expiresIn)
{
    auto session = std::make_shared<AuthSession>();
    session->username = username;
    session->uuid = uuid;
    session->access_token = accessToken;
    session->refresh_token = refreshToken;
    session->expires_at = QDateTime::currentMSecsSinceEpoch() + static_cast<qint64>(expiresIn) * 1000;
    session->is_online = true;
    m_session = session;
    // Keep the HTTP client's bearer token in sync so other API clients
    // (SkinClient) can authenticate without a separate token source.
    if (m_client != nullptr)
        m_client->setAccessToken(accessToken);
    // The in-memory session stays valid for this run, so a persistence failure
    // (disk full, DPAPI unavailable) must not fail the login, but it cannot be
    // silently ignored either: the next launch would ask for credentials again.
    if (!m_storage->save(*session))
        qWarning() << "Could not persist the auth session for" << username;
}
