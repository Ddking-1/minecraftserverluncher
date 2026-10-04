// SPDX-License-Identifier: GPL-3.0-only
#include "RotatingFileLogger.h"

#include <QDir>
#include <QFile>
#include <QFileInfo>

namespace RotatingFileLogger {

void rotate(const QString &filePath, int maxFiles)
{
    if (maxFiles < 1)
        return;

    QFile::remove(filePath + QLatin1Char('.') + QString::number(maxFiles));

    for (int i = maxFiles - 1; i >= 1; --i) {
        const QString from = filePath + QLatin1Char('.') + QString::number(i);
        const QString to = filePath + QLatin1Char('.') + QString::number(i + 1);
        if (QFile::exists(from)) {
            QFile::remove(to);
            QFile::rename(from, to);
        }
    }

    if (QFile::exists(filePath))
        QFile::rename(filePath, filePath + QLatin1String(".1"));
}

bool log(const QString &filePath, const QString &line, qint64 maxBytes, int maxFiles)
{
    const QFileInfo info(filePath);
    if (info.exists() && info.size() >= maxBytes)
        rotate(filePath, maxFiles);

    const QString dir = QFileInfo(filePath).absolutePath();
    if (!dir.isEmpty())
        QDir().mkpath(dir);

    QFile file(filePath);
    if (!file.open(QIODevice::Append | QIODevice::Text))
        return false;

    const QByteArray data = line.toUtf8() + '\n';
    if (file.write(data) != data.size())
        return false;
    return file.flush();
}

} // namespace RotatingFileLogger
