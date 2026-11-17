// SPDX-License-Identifier: GPL-3.0-only
#include "LauncherPartLaunch.h"

#include <QDir>
#include <QProcess>
#include <QtGlobal>

#include "minecraft/MinecraftInstance.h"

LauncherPartLaunch::LauncherPartLaunch(LaunchTask *parent, MinecraftInstance *instance,
                                       AuthSessionPtr session, MinecraftTargetPtr target)
    : LaunchStep(parent), m_instance(instance), m_session(std::move(session)),
      m_target(std::move(target))
{
}

void LauncherPartLaunch::executeTask()
{
    if (!m_instance) {
        emit failed(QStringLiteral("no instance to launch"));
        return;
    }
    const QString javaPath = m_instance->javaPath();
    if (javaPath.isEmpty()) {
        emit failed(QStringLiteral("no Java runtime configured"));
        return;
    }

    QStringList args = buildJvmArgs();
    args << QStringLiteral("-cp") << QStringLiteral("__CLASSPATH__"); // filled by library resolver
    args << QStringLiteral("net.minecraft.launcher.Main");
    args << buildGameArgs();

    m_process = new QProcess(this);
    m_process->setWorkingDirectory(m_instance->gameRoot());
    m_process->start(javaPath, args);

    if (!m_process->waitForStarted(10000)) {
        emit failed(QStringLiteral("failed to start the game process"));
        return;
    }
    emit logLine(QStringLiteral("Launched game process (pid %1)").arg(m_process->processId()));
    emit readyForLaunch();
}

void LauncherPartLaunch::abort()
{
    if (m_process)
        m_process->kill();
}

QStringList LauncherPartLaunch::buildJvmArgs() const
{
    // Defensive clamping: the JVM refuses to start when -Xms is larger than
    // -Xmx or when a heap value is zero/negative, so never pass raw values.
    int maxMb = m_instance->maxMemoryMb();
    int minMb = m_instance->minMemoryMb();
    maxMb = qBound(512, maxMb, 65536);
    minMb = qBound(256, minMb, maxMb);

    QStringList args;
    args << QStringLiteral("-Xmx%1M").arg(maxMb);
    args << QStringLiteral("-Xms%1M").arg(minMb);
    args << QStringLiteral("-XX:+UseG1GC");
    args << QStringLiteral("-Djava.library.path=%1").arg(QDir::toNativeSeparators(m_instance->nativesDir()));
    return args;
}

QStringList LauncherPartLaunch::buildGameArgs() const
{
    QStringList args;
    args << QStringLiteral("--username") << (m_session ? m_session->username : QStringLiteral("Player"));
    args << QStringLiteral("--uuid") << (m_session ? m_session->uuid : QString());
    args << QStringLiteral("--accessToken") << (m_session ? m_session->access_token : QString());
    args << QStringLiteral("--version") << QStringLiteral("1.21.11");
    args << QStringLiteral("--gameDir") << QDir::toNativeSeparators(m_instance->gameRoot());
    args << QStringLiteral("--assetsDir") << QDir::toNativeSeparators(m_instance->assetsDir());
    if (m_target && !m_target->address.isEmpty())
        args << QStringLiteral("--server") << m_target->address << QStringLiteral("--port")
             << QString::number(m_target->port);
    return args;
}
