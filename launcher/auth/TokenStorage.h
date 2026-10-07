// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher core. Persistent storage for backend auth tokens.
#pragma once

#include <QString>

#include "core/InstanceTypes.h"

// Persists a session's tokens to disk so the user stays logged in across
// launches. On Windows the token blob is protected with the current user's
// DPAPI credentials (CryptProtectData); on other platforms, or if DPAPI is
// unavailable, it falls back to plaintext JSON (documented limitation,
// docs/AUTH.md).
class TokenStorage
{
public:
    // Directory that will hold the token file. Created on demand.
    explicit TokenStorage(const QString &dataDir);

    // Save the session's tokens. Returns false on write failure.
    [[nodiscard]]
    bool save(const AuthSession &session) const;

    // Load a stored session for the given username. Returns false if absent.
    [[nodiscard]]
    bool load(const QString &username, AuthSession &out) const;

    // Remove the stored session for the given username.
    void remove(const QString &username) const;

    // Path of the token file for a username.
    [[nodiscard]]
    QString filePathFor(const QString &username) const;

private:
    QString m_dataDir;
};
