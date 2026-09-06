// SPDX-License-Identifier: GPL-3.0-only
#include "JavaVersion.h"

#include <QRegularExpression>

namespace JavaVersionParser {

JavaVersion parseFromJavaVersionOutput(const QString &output)
{
    JavaVersion v;
    v.raw = output.trimmed();

    // Match a version string like "24.0.1" or "1.8.0_401" anywhere in the line.
    static const QRegularExpression re(QStringLiteral("(\\d+)(?:\\.(\\d+))?(?:\\.(\\d+))?"));
    const QRegularExpressionMatch m = re.match(output);
    if (!m.hasMatch())
        return v;

    const int first = m.captured(1).toInt();
    if (first == 1) {
        // Legacy scheme: "1.8.0_401" -> major 8.
        v.major = m.captured(2).toInt();
        v.minor = m.captured(3).toInt();
    } else {
        // Modern scheme: "24.0.1" -> major 24.
        v.major = first;
        v.minor = m.captured(2).toInt();
        v.patch = m.captured(3).toInt();
    }
    return v;
}

} // namespace JavaVersionParser
