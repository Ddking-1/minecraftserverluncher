// SPDX-License-Identifier: GPL-3.0-only
#include "SkinCache.h"

#include <QCryptographicHash>
#include <QDir>
#include <QFile>
#include <QSaveFile>

namespace SkinCache {

QString cacheFileName(const SkinModel::SkinRef &skin)
{
    return skin.id + QLatin1Char('_') + skin.sha256.toLower() + QStringLiteral(".png");
}

QString cachedPath(const QString &cacheDir, const SkinModel::SkinRef &skin)
{
    return QDir(cacheDir).filePath(cacheFileName(skin));
}

namespace {

QString fileSha256(const QString &path)
{
    QFile file(path);
    if (!file.open(QIODevice::ReadOnly))
        return QString();
    QCryptographicHash hash(QCryptographicHash::Sha256);
    hash.addData(&file);
    return QString::fromLatin1(hash.result().toHex());
}

} // namespace

bool store(const QString &cacheDir, const SkinModel::SkinRef &skin,
           const QByteArray &data, QString *outPath)
{
    const QString computed =
        QString::fromLatin1(QCryptographicHash::hash(data, QCryptographicHash::Sha256).toHex());
    if (computed.compare(skin.sha256, Qt::CaseInsensitive) != 0)
        return false;

    if (!QDir().mkpath(cacheDir))
        return false;

    const QString path = cachedPath(cacheDir, skin);
    QSaveFile file(path);
    if (!file.open(QIODevice::WriteOnly))
        return false;
    if (file.write(data) != data.size())
        return false;
    if (!file.commit())
        return false;

    if (outPath)
        *outPath = path;
    return true;
}

bool contains(const QString &cacheDir, const SkinModel::SkinRef &skin)
{
    const QString path = cachedPath(cacheDir, skin);
    if (!QFile::exists(path))
        return false;
    const QString actual = fileSha256(path);
    return actual.compare(skin.sha256, Qt::CaseInsensitive) == 0;
}

} // namespace SkinCache
