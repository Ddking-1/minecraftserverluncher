// SPDX-License-Identifier: GPL-3.0-only
#pragma once

#include <QString>

namespace RotatingFileLogger {

// Rotate filePath to filePath.1, filePath.1 to filePath.2, ... keeping at most
// maxFiles rotated files. maxFiles < 1 disables rotation.
void rotate(const QString &filePath, int maxFiles);

// Append line + newline to filePath, rotating first when the current file has
// reached maxBytes. Returns false if the file cannot be written.
bool log(const QString &filePath, const QString &line,
         qint64 maxBytes = 1048576, int maxFiles = 3);

} // namespace RotatingFileLogger
