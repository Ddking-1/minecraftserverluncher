// SPDX-License-Identifier: GPL-3.0-only
#include "ModScanner.h"

#include <QDir>
#include <QFileInfo>
#include <QStringList>

#include <algorithm>

namespace ModScanner {

QList<ModEntry> scan(const QString &modsDir)
{
    QList<ModEntry> result;

    QDir dir(modsDir);
    if (!dir.exists())
        return result;

    const QStringList nameFilters{QStringLiteral("*.jar"), QStringLiteral("*.jar.disabled")};
    const QFileInfoList files = dir.entryInfoList(nameFilters, QDir::Files, QDir::Name);

    for (const QFileInfo &file : files) {
        ModEntry entry;
        entry.fileName = file.fileName();
        entry.enabled = !file.fileName().endsWith(QLatin1String(".disabled"));

        QString name = file.fileName();
        if (!entry.enabled)
            name.chop(static_cast<int>(qstrlen(".disabled")));
        if (name.endsWith(QLatin1String(".jar")))
            name.chop(4);
        entry.displayName = name;

        result.append(entry);
    }

    std::sort(result.begin(), result.end(), [](const ModEntry &a, const ModEntry &b) {
        return a.displayName.compare(b.displayName, Qt::CaseInsensitive) < 0;
    });

    return result;
}

} // namespace ModScanner
