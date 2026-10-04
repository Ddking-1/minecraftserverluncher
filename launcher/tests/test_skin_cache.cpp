// SPDX-License-Identifier: GPL-3.0-only
// Standalone verification for the content-addressed skin cache.
#include <iostream>

#include <QCryptographicHash>
#include <QFile>
#include <QTemporaryDir>

#include "core/SkinCache.h"

namespace {

bool check(bool cond, const char *what)
{
    if (!cond)
        std::cerr << "FAIL: " << what << std::endl;
    return cond;
}

SkinModel::SkinRef makeSkin(const QString &id, const QByteArray &data)
{
    SkinModel::SkinRef skin;
    skin.id = id;
    skin.sha256 = QString::fromLatin1(
        QCryptographicHash::hash(data, QCryptographicHash::Sha256).toHex());
    skin.url = QStringLiteral("https://example.com/skin.png");
    return skin;
}

} // namespace

int main()
{
    using namespace SkinCache;

    QTemporaryDir dir;
    if (!check(dir.isValid(), "temporary dir"))
        return 1;
    const QString cacheDir = dir.filePath(QStringLiteral("skincache"));

    const QByteArray data("fake-png-bytes");
    const SkinModel::SkinRef skin = makeSkin(QStringLiteral("steve-01"), data);

    QString outPath;
    if (!check(store(cacheDir, skin, data, &outPath), "store matching hash"))
        return 1;
    if (!check(QFile::exists(outPath), "stored file exists"))
        return 1;
    if (!check(contains(cacheDir, skin), "contains after store"))
        return 1;

    if (!check(!contains(cacheDir, makeSkin(QStringLiteral("steve-01"), QByteArray("other"))),
                "mismatched hash not present"))
        return 1;

    if (!check(!store(cacheDir, skin, QByteArray("tampered")), "store rejects tampered data"))
        return 1;

    std::cout << "OK" << std::endl;
    return 0;
}
