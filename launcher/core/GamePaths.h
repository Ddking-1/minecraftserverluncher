// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher core. Classpath construction and native library
// layout helpers for the game launch.
#pragma once

#include <QString>
#include <QStringList>

namespace GamePaths {

// Build the JVM classpath from the list of absolute library jar paths. Each
// path is canonicalized and verified to be a .jar inside the libraries dir
// before it is included, so an attacker-controlled manifest cannot add an
// arbitrary local jar.
QString buildClasspath(const QString &librariesDir, const QStringList &jarPaths);

// Resolve a natives directory path, ensuring it is absolute and does not
// escape a given base directory via .. or symlinks.
QString safeNativesDir(const QString &baseDir, const QString &subDir);

} // namespace GamePaths
