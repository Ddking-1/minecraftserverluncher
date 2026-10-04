// SPDX-License-Identifier: GPL-3.0-only
#pragma once

#include <QList>
#include <QString>

namespace ModScanner {

// A mod jar discovered in an instance mods/ directory.
struct ModEntry
{
    QString fileName;    // on-disk file name, including any .disabled suffix
    QString displayName; // human-readable name without .jar/.disabled
    bool enabled = true; // false for *.jar.disabled
};

// Scan modsDir for *.jar and *.jar.disabled files, sorted by display name
// (case-insensitive). A missing directory yields an empty list.
QList<ModEntry> scan(const QString &modsDir);

} // namespace ModScanner
