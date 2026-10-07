// SPDX-License-Identifier: GPL-3.0-only
#pragma once

#include <QString>

namespace FileIntegrity {

// Lowercase hex SHA-256 digest of the file at path, or an empty string if the
// file cannot be read.
[[nodiscard]]
QString sha256File(const QString &path);

// True when the file exists, can be read, and its SHA-256 equals
// expectedSha256 (case-insensitive).
[[nodiscard]]
bool verify(const QString &path, const QString &expectedSha256);

} // namespace FileIntegrity
