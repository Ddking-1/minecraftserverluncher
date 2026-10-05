// SPDX-License-Identifier: GPL-3.0-only
#pragma once

#include <QObject>
#include <QString>
#include <QStringList>

class JavaDetector : public QObject
{
    Q_OBJECT
public:
    static bool isJavaAvailable();
    static QString getJavaPath();
    static int javaMajorVersion(const QString &javaExe);
    static QString detectJava24();

    // True when javaExe is a plausible, executable java binary inside a
    // Java home directory (bin/java.exe on Windows). Guards against an
    // attacker-controlled LAUNCHER_JAVA_HOME pointing at an arbitrary
    // executable that would run with the launcher's privileges.
    static bool isPlausibleJava(const QString &javaExe);

private:
    static QStringList candidateJavaHomes();
    static const int kMinJavaMajor;
};
