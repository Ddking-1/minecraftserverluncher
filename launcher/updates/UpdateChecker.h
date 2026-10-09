// SPDX-License-Identifier: GPL-3.0-only
#pragma once

#include <QString>

#include "updates/UpdateManifest.h"

namespace UpdateChecker {

// Result of comparing the running launcher against a manifest component.
enum class Decision {
    UpToDate,
    UpdateAvailable,
    Incompatible,
};

// Compare two dotted numeric versions ("1.2.10" vs "1.2.9"). Missing
// components count as 0. Returns -1, 0 or 1.
[[nodiscard]]
int compareVersions(QStringView a, QStringView b);

// Decide whether the manifest component can be applied to currentVersion.
[[nodiscard]]
Decision check(QStringView currentVersion,
               const UpdateManifest::ManifestComponent &component);

} // namespace UpdateChecker
