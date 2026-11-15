// SPDX-License-Identifier: GPL-3.0-only
#include "TokenStorage.h"

#include <QDir>
#include <QFile>
#include <QFileInfo>
#include <QJsonDocument>
#include <QJsonObject>
#include <QSaveFile>

#ifdef Q_OS_WIN
#include <windows.h>
#include <wincrypt.h>
#endif

namespace {

// Marker prefix for DPAPI-protected files (version 1).
constexpr char kDpapiPrefix[] = "DPAPI1:";

// Usernames become file names, so they must not contain path separators or
// traversal sequences. This mirrors the backend's username rule.
[[nodiscard]]
bool isSafeUsername(const QString &username)
{
    if (username.isEmpty() || username.size() > 64)
        return false;
    for (const QChar &c : username) {
        const ushort code = c.unicode();
        const bool allowed = (code >= 'a' && code <= 'z') || (code >= 'A' && code <= 'Z')
            || (code >= '0' && code <= '9') || code == '-' || code == '_';
        if (!allowed)
            return false;
    }
    return true;
}

#ifdef Q_OS_WIN
// Protect plaintext with the current user's DPAPI credentials.
[[nodiscard]]
QByteArray dpapiProtect(const QByteArray &plain)
{
    DATA_BLOB in;
    in.pbData = reinterpret_cast<BYTE *>(const_cast<char *>(plain.constData()));
    in.cbData = static_cast<DWORD>(plain.size());

    DATA_BLOB out{};
    if (!CryptProtectData(&in, L"launcher-auth-token", nullptr, nullptr, nullptr,
                          CRYPTPROTECT_UI_FORBIDDEN, &out)) {
        return {};
    }
    QByteArray result(reinterpret_cast<const char *>(out.pbData),
                      static_cast<int>(out.cbData));
    LocalFree(out.pbData);
    return result;
}

// Unprotect a DPAPI blob produced by dpapiProtect.
[[nodiscard]]
QByteArray dpapiUnprotect(const QByteArray &blob)
{
    DATA_BLOB in;
    in.pbData = reinterpret_cast<BYTE *>(const_cast<char *>(blob.constData()));
    in.cbData = static_cast<DWORD>(blob.size());

    DATA_BLOB out{};
    if (!CryptUnprotectData(&in, nullptr, nullptr, nullptr, nullptr,
                            CRYPTPROTECT_UI_FORBIDDEN, &out)) {
        return {};
    }
    QByteArray result(reinterpret_cast<const char *>(out.pbData),
                      static_cast<int>(out.cbData));
    LocalFree(out.pbData);
    return result;
}
#endif // Q_OS_WIN

} // namespace

TokenStorage::TokenStorage(const QString &dataDir) : m_dataDir(dataDir)
{
}

QString TokenStorage::filePathFor(const QString &username) const
{
    if (!isSafeUsername(username))
        return QString();
    return m_dataDir + QStringLiteral("/tokens/") + username + QStringLiteral(".json");
}

bool TokenStorage::save(const AuthSession &session) const
{
    const QString path = filePathFor(session.username);
    if (path.isEmpty())
        return false;
    QDir().mkpath(QFileInfo(path).absolutePath());

    QJsonObject obj;
    obj.insert(QStringLiteral("username"), session.username);
    obj.insert(QStringLiteral("uuid"), session.uuid);
    obj.insert(QStringLiteral("access_token"), session.access_token);
    obj.insert(QStringLiteral("refresh_token"), session.refresh_token);
    obj.insert(QStringLiteral("expires_at"), session.expires_at);
    obj.insert(QStringLiteral("is_online"), session.is_online);

    const QByteArray json = QJsonDocument(obj).toJson(QJsonDocument::Compact);

    // Protect the token blob with the OS credential store (Windows DPAPI).
    // If DPAPI is unavailable, fall back to plaintext so the launcher still
    // works (documented limitation; docs/AUTH.md).
    QByteArray data = json;
#ifdef Q_OS_WIN
    const QByteArray protectedBlob = dpapiProtect(json);
    if (!protectedBlob.isEmpty()) {
        data = QByteArray(kDpapiPrefix) + protectedBlob.toBase64();
    }
#endif

    QSaveFile file(path);
    if (!file.open(QIODevice::WriteOnly))
        return false;
    if (file.write(data) != data.size())
        return false;
    return file.commit();
}

bool TokenStorage::load(const QString &username, AuthSession &out) const
{
    const QString path = filePathFor(username);
    if (path.isEmpty())
        return false;
    QFile file(path);
    if (!file.open(QIODevice::ReadOnly))
        return false;
    const QByteArray raw = file.readAll();
    file.close();

    QByteArray data = raw;
    if (raw.startsWith(kDpapiPrefix)) {
#ifdef Q_OS_WIN
        const QByteArray blob = QByteArray::fromBase64(raw.mid(static_cast<int>(sizeof(kDpapiPrefix)) - 1));
        data = dpapiUnprotect(blob);
        if (data.isEmpty())
            return false;
#else
        // DPAPI-protected file on a non-Windows host cannot be read.
        return false;
#endif
    }

    QJsonParseError err{};
    const QJsonDocument doc = QJsonDocument::fromJson(data, &err);
    if (err.error != QJsonParseError::NoError || !doc.isObject())
        return false;

    const QJsonObject obj = doc.object();
    out.username = obj.value(QStringLiteral("username")).toString();
    out.uuid = obj.value(QStringLiteral("uuid")).toString();
    out.access_token = obj.value(QStringLiteral("access_token")).toString();
    out.refresh_token = obj.value(QStringLiteral("refresh_token")).toString();
    out.expires_at = static_cast<qint64>(obj.value(QStringLiteral("expires_at")).toDouble());
    out.is_online = obj.value(QStringLiteral("is_online")).toBool();
    return !out.username.isEmpty();
}

void TokenStorage::remove(const QString &username) const
{
    const QString path = filePathFor(username);
    if (!path.isEmpty())
        QFile::remove(path);
}
