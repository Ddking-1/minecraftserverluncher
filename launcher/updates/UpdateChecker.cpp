// SPDX-License-Identifier: GPL-3.0-only
#include "UpdateChecker.h"

#include <QStringList>

namespace UpdateChecker {

int compareVersions(QStringView a, QStringView b)
{
    const auto aParts = a.split(QLatin1Char('.'), Qt::SkipEmptyParts);
    const auto bParts = b.split(QLatin1Char('.'), Qt::SkipEmptyParts);

    const int count = qMax(aParts.size(), bParts.size());
    for (int i = 0; i < count; ++i) {
        const int av = i < aParts.size() ? aParts.at(i).toInt() : 0;
        const int bv = i < bParts.size() ? bParts.at(i).toInt() : 0;
        if (av < bv)
            return -1;
        if (av > bv)
            return 1;
    }
    return 0;
}

Decision check(QStringView currentVersion,
               const UpdateManifest::ManifestComponent &component)
{
    if (!component.minLauncher.isEmpty()
        && compareVersions(currentVersion, component.minLauncher) < 0) {
        return Decision::Incompatible;
    }
    if (compareVersions(currentVersion, component.version) < 0)
        return Decision::UpdateAvailable;
    return Decision::UpToDate;
}

} // namespace UpdateChecker
