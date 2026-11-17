// SPDX-License-Identifier: GPL-3.0-only
// Regression test: hostile/garbage memory values must never reach the JVM as
// an unlaunchable -Xmx/-Xms pair.
#include <QCoreApplication>

#include "launch/steps/LauncherPartLaunch.h"
#include "minecraft/MinecraftInstance.h"

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

    MinecraftInstance instance;

    // Zero / negative values are clamped to a usable minimum.
    instance.setMaxMemoryMb(0);
    check(instance.maxMemoryMb() >= 512, "zero max clamped up");
    instance.setMaxMemoryMb(-4096);
    check(instance.maxMemoryMb() >= 512, "negative max clamped up");

    // Absurd values are clamped down.
    instance.setMaxMemoryMb(99999999);
    check(instance.maxMemoryMb() == 65536, "huge max clamped to 65536");

    // Minimum can never exceed maximum.
    instance.setMaxMemoryMb(2048);
    instance.setMinMemoryMb(8192);
    check(instance.minMemoryMb() <= instance.maxMemoryMb(), "min <= max enforced");
    instance.setMinMemoryMb(0);
    check(instance.minMemoryMb() == 256, "zero min clamped to 256");

    // The assembled JVM arguments carry the clamped values and stay coherent.
    LauncherPartLaunch step(nullptr, &instance, AuthSessionPtr(), MinecraftTargetPtr());
    const QStringList args = step.buildJvmArgs();
    const QString xmx = args.filter(QStringLiteral("-Xmx")).value(0);
    const QString xms = args.filter(QStringLiteral("-Xms")).value(0);
    check(!xmx.isEmpty(), "-Xmx present");
    check(!xms.isEmpty(), "-Xms present");

    bool okMax = false;
    bool okMin = false;
    const int maxMb = xmx.mid(4).chopped(1).toInt(&okMax);
    const int minMb = xms.mid(4).chopped(1).toInt(&okMin);
    check(okMax && okMin, "-Xmx/-Xms values are numeric");
    check(okMax && maxMb >= 512 && maxMb <= 65536, "-Xmx in range");
    check(okMin && minMb >= 256, "-Xms at least 256");
    check(minMb <= maxMb, "-Xms <= -Xmx");

    // A deliberately inconsistent instance must still produce a launchable pair.
    instance.setMaxMemoryMb(1024);
    instance.setMinMemoryMb(1024);
    const QStringList args2 = step.buildJvmArgs();
    const QString xmx2 = args2.filter(QStringLiteral("-Xmx")).value(0);
    const QString xms2 = args2.filter(QStringLiteral("-Xms")).value(0);
    check(xmx2 == QStringLiteral("-Xmx1024M"), "-Xmx reflects instance value");
    check(xms2 == QStringLiteral("-Xms1024M"), "-Xms reflects instance value");

    if (failures == 0)
        std::printf("memory_clamp: all checks passed\n");
    return failures == 0 ? 0 : 1;
}
