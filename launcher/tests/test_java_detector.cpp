// SPDX-License-Identifier: GPL-3.0-only
#include <QCoreApplication>
#include <QTemporaryDir>

#include "game/JavaDetector.h"

#include <cstdio>

static int failures = 0;
static void check(bool cond, const char *msg)
{
    if (!cond) {
        std::fprintf(stderr, "FAIL: %s\n", msg);
        ++failures;
    }
}

int main(int argc, char **argv)
{
    QCoreApplication app(argc, argv);

    // Reject empty / implausible paths.
    check(!JavaDetector::isPlausibleJava(QString()), "empty path rejected");
    check(!JavaDetector::isPlausibleJava(QStringLiteral("C:/Windows/System32/cmd.exe")), "cmd.exe rejected");
    check(!JavaDetector::isPlausibleJava(QStringLiteral("/tmp/bin/notjava.exe")), "non-java name rejected");
    check(!JavaDetector::isPlausibleJava(QStringLiteral("/tmp/other/java.exe")), "non-bin dir rejected");

    // A real JDK on this machine should be plausible.
    const QString javaHome = qEnvironmentVariable("JAVA_HOME");
    if (!javaHome.isEmpty()) {
        const QString exe = javaHome + QStringLiteral("/bin/java.exe");
        if (QFile::exists(exe)) {
            check(JavaDetector::isPlausibleJava(exe), "real JDK java is plausible");
        }
    }

    // javaMajorVersion on a real java returns >= 0.
    const QString path = JavaDetector::getJavaPath();
    if (!path.isEmpty()) {
        const int v = JavaDetector::javaMajorVersion(path);
        check(v >= 21, "detected major version >= 21");
    }

    if (failures == 0)
        std::printf("java_detector: all checks passed\n");
    return failures == 0 ? 0 : 1;
}
