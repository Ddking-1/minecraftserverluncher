// SPDX-License-Identifier: GPL-3.0-only
#include <iostream>

#include <QFile>
#include <QFileInfo>
#include <QTemporaryDir>

#include "auth/TokenStorage.h"

namespace {

bool check(bool cond, const char *what)
{
    if (!cond)
        std::cerr << "FAIL: " << what << std::endl;
    return cond;
}

AuthSession sampleSession(const QString &username)
{
    AuthSession s;
    s.username = username;
    s.uuid = QStringLiteral("uuid");
    s.access_token = QStringLiteral("ACCESS_SECRET_TOKEN");
    s.refresh_token = QStringLiteral("REFRESH_SECRET_TOKEN");
    s.expires_at = 12345;
    s.is_online = true;
    return s;
}

} // namespace

int main()
{
    QTemporaryDir dir;
    if (!check(dir.isValid(), "temporary dir"))
        return 1;

    TokenStorage storage(dir.path());

    // Secure storage only: a failed save must not leave a plaintext token file.
    const AuthSession good = sampleSession(QStringLiteral("player1"));
    const bool saved = storage.save(good);
    const QString goodPath = storage.filePathFor(good.username);
    if (!check(!goodPath.isEmpty(), "valid username path"))
        return 1;

    if (saved) {
        QFile file(goodPath);
        if (!check(file.open(QIODevice::ReadOnly), "open saved token file"))
            return 1;
        const QByteArray raw = file.readAll();
        if (!check(raw.startsWith("DPAPI1:"), "saved token must be DPAPI-protected")
            || !check(!raw.contains(good.access_token.toUtf8()), "no plaintext access token")
            || !check(!raw.contains(good.refresh_token.toUtf8()), "no plaintext refresh token"))
            return 1;
    } else {
        if (!check(!QFileInfo::exists(goodPath), "failed save must not create token file"))
            return 1;
    }

    // Path traversal and separator injection are rejected.
    const QStringList invalidUsers = {
        QStringLiteral("../evil"),
        QStringLiteral("..\\evil"),
        QStringLiteral("a/b"),
        QStringLiteral("a\\b"),
        QStringLiteral(".."),
    };
    for (const QString &username : invalidUsers) {
        const QString path = storage.filePathFor(username);
        if (!check(path.isEmpty(), "unsafe username path rejected"))
            return 1;
        if (!check(!storage.save(sampleSession(username)), "unsafe username save rejected"))
            return 1;

        AuthSession out;
        if (!check(!storage.load(username, out), "unsafe username load rejected"))
            return 1;
    }

    std::cout << "OK" << std::endl;
    return 0;
}
