// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher core. Persistent storage for backend auth tokens.
#pragma once

#include <QString>

#include "core/InstanceTypes.h"

// Persists a session's tokens to disk so the user stays logged in across
// launches. Tokens are only saved when OS-protected encryption is available.
// Unsafe usernames are rejected to prevent path traversal.
class TokenStorage
{
public:
    // Directory that will hold the token file. Created on demand.
    explicit TokenStorage(const QString &dataDir);

    // Save the session's tokens. Returns false if secure encryption or write
    // fails.
    bool save(const AuthSession &session) const;

    // Load a stored session for the given username. Returns false if absent.
    bool load(const QString &username, AuthSession &out) const;

    // Remove the stored session for the given username.
    void remove(const QString &username) const;

    // Path of the token file for a username.
    QString filePathFor(const QString &username) const;

private:
    QString m_dataDir;
};
