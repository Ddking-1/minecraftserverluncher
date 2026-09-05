// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher core. Java version parsing helpers.
#pragma once

#include <QString>

// Parsed Java version information.
struct JavaVersion
{
    int major = 0;   // e.g. 24 for Java 24
    int minor = 0;
    int patch = 0;
    QString vendor;
    QString raw;

    bool isValid() const { return major > 0; }
};

namespace JavaVersionParser {

// Parse the output of `java -version` (second line, e.g.
// "openjdk version \"24.0.1\" 2025-04-15"). Returns an invalid JavaVersion on
// failure.
JavaVersion parseFromJavaVersionOutput(const QString &output);

} // namespace JavaVersionParser
