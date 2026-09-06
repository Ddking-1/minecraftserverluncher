// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher core. Model of a detected Java runtime.
#pragma once

#include <QString>

#include "JavaVersion.h"

// A Java runtime discovered on the system or configured for an instance.
struct JavaInstall
{
    QString path;      // full path to java.exe
    QString home;      // JAVA_HOME directory
    JavaVersion version;
    bool is64bit = true;

    bool isValid() const { return !path.isEmpty() && version.isValid(); }
};
