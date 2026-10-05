// SPDX-License-Identifier: GPL-3.0-only
#include "GamePaths.h"

#include <QDir>
#include <QFileInfo>

namespace GamePaths {

QString buildClasspath(const QString &librariesDir, const QStringList &jarPaths)
{
    const QString canonicalLibDir = QFileInfo(librariesDir).absoluteFilePath();
    QStringList classpath;
    for (const QString &jar : jarPaths) {
        // Only accept existing files ending in .jar that live inside the
        // libraries directory (no absolute escapes, no arbitrary jars).
        const QFileInfo info(jar);
        if (!info.exists() || info.isDir())
            continue;
        if (!info.fileName().endsWith(QStringLiteral(".jar"), Qt::CaseInsensitive))
            continue;
        const QString canonical = info.absoluteFilePath();
        if (!canonical.startsWith(canonicalLibDir + QLatin1Char('/')))
            continue;
        classpath << canonical;
    }
    return classpath.join(QLatin1Char(';'));
}

QString safeNativesDir(const QString &baseDir, const QString &subDir)
{
    const QString canonicalBase = QFileInfo(baseDir).absoluteFilePath();
    // Reject subDir that is absolute or escapes the base via .. /
    // canonicalization.
    QFileInfo info(subDir);
    if (info.isAbsolute())
        return QString();
    const QString candidate = canonicalBase + QLatin1Char('/') + subDir;
    const QString canonical = QFileInfo(candidate).absoluteFilePath();
    if (!canonical.startsWith(canonicalBase + QLatin1Char('/')))
        return QString();
    return canonical;
}

} // namespace GamePaths
