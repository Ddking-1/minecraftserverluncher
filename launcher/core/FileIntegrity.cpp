// SPDX-License-Identifier: GPL-3.0-only
#include "FileIntegrity.h"

#include <QCryptographicHash>
#include <QFile>

namespace FileIntegrity {

QString sha256File(const QString &path)
{
    QFile file(path);
    if (!file.open(QIODevice::ReadOnly))
        return QString();

    QCryptographicHash hash(QCryptographicHash::Sha256);
    if (!hash.addData(&file))
        return QString();

    return QString::fromLatin1(hash.result().toHex());
}

bool verify(const QString &path, const QString &expectedSha256)
{
    const QString actual = sha256File(path);
    if (actual.isEmpty())
        return false;
    return actual.compare(expectedSha256, Qt::CaseInsensitive) == 0;
}

} // namespace FileIntegrity
