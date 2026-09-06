// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher core. Shared value types used across the launcher.
#pragma once

#include <QString>
#include <memory>

// A session handed to the launch pipeline. Project-owned replacement for
// Prism's MinecraftAccount/AuthSession: the account is backend-authenticated,
// never a Mojang/MSA account.
struct AuthSession
{
    QString username;
    QString uuid;
    QString access_token;
    QString refresh_token;
    qint64 expires_at = 0; // epoch ms; 0 == unknown/offline
    bool is_online = false;
};
using AuthSessionPtr = std::shared_ptr<AuthSession>;

// Target for a launch: which server to join (optional).
struct MinecraftTarget
{
    QString address;
    quint16 port = 0;
    QString world;
};
using MinecraftTargetPtr = std::shared_ptr<MinecraftTarget>;
