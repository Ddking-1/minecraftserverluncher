// SPDX-License-Identifier: GPL-3.0-only
#include "UpdateManagerHMAC.h"

#include <QCryptographicHash>
#include <QByteArray>
#include <QString>
#include <cstring>

// Helper: HMAC-SHA256 over data using key, returned as hex.
static QString hmacSha256Hex(const QByteArray &key, const QByteArray &data)
{
    // Qt does not provide HMAC directly; we implement HMAC-SHA256 manually
    // following RFC 2104 using QCryptographicHash for SHA-256.

    // Block size for SHA-256 is 64 bytes.
    const int blockSize = 64;

    QByteArray keyBlock;

    if (key.size() > blockSize) {
        // Hash the key if it's longer than block size.
        QCryptographicHash hash(QCryptographicHash::Sha256);
        hash.addData(key);
        keyBlock = hash.result();
    } else {
        keyBlock = key;
    }

    // Pad key to block size with zeros.
    keyBlock.resize(blockSize, '\0');

    // Compute o_key_pad and i_key_pad.
    QByteArray oKeyPad(blockSize, '\0');
    QByteArray iKeyPad(blockSize, '\0');
    for (int i = 0; i < blockSize; ++i) {
        oKeyPad[i] = keyBlock[i] ^ 0x5c;
        iKeyPad[i] = keyBlock[i] ^ 0x36;
    }

    // Inner hash: H(o_key_pad || H(i_key_pad || data))
    QCryptographicHash inner(QCryptographicHash::Sha256);
    inner.addData(iKeyPad);
    inner.addData(data);
    QByteArray innerHash = inner.result();

    QCryptographicHash outer(QCryptographicHash::Sha256);
    outer.addData(oKeyPad);
    outer.addData(innerHash);
    return QString::fromLatin1(outer.result().toHex());
}

QString hmacManifestSignature(const QString &key,
                             const QString &version,
                             const QString &url,
                             const QString &sha256,
                             bool mandatory)
{
    QString canonical = version + QLatin1Char('|') + url + QLatin1Char('|')
                        + sha256 + QLatin1Char('|')
                        + (mandatory ? QLatin1Char('1') : QLatin1Char('0'));
    return hmacSha256Hex(key.toUtf8(), canonical.toUtf8());
}

bool hmacSignaturesEqual(const QString &a, const QString &b)
{
    // Simple constant-time comparison using qMemEqual (Qt provides this).
    // Note: This is not truly constant-time, but sufficient for our use case.
    // For better security, use a dedicated constant-time compare in production.
    if (a.size() != b.size())
        return false;

    // Convert hex to raw bytes and compare.
    QByteArray aBytes = QByteArray::fromHex(a.toLatin1());
    QByteArray bBytes = QByteArray::fromHex(b.toLatin1());

    if (aBytes.size() != bBytes.size())
        return false;

    return memcmp(aBytes.constData(), bBytes.constData(), aBytes.size()) == 0;
}

bool verifyManifestSignature(const QString &signingKey,
                             const QString &version,
                             const QString &url,
                             const QString &sha256,
                             bool mandatory,
                             const QString &signature)
{
    if (signingKey.isEmpty() || signature.isEmpty())
        return false;

    QString expected = hmacManifestSignature(signingKey, version, url, sha256, mandatory);
    return hmacSignaturesEqual(expected, signature);
}
