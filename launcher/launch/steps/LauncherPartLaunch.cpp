// SPDX-License-Identifier: GPL-3.0-only
#include "LauncherPartLaunch.h"

#include <QDir>
#include <QProcess>
#include <QtGlobal>

#include "minecraft/MinecraftInstance.h"
#include "core/SystemMemory.h"
#include "core/GameExitDiagnostics.h"

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

    // Refuse an impossible heap up front: the JVM would either fail with an
    // opaque error or push the machine into swapping. The numbers are part of
    // the message so the user sees *why* the setting is rejected.
    const int totalMb = SystemMemory::totalPhysicalMb();
    const int allowedMb = SystemMemory::allowedMaxMb(totalMb);
    const int requestedMb =
        SystemMemory::clampToAllowed(m_instance->maxMemoryMb(), SystemMemory::kAbsoluteMaxMb);
    const QString memoryHint = totalMb > 0
        ? QStringLiteral("%1 MB installed, at most %2 MB allowed (%3 MB kept for the system)")
              .arg(totalMb)
              .arg(allowedMb)
              .arg(SystemMemory::kReservedMb)
        : QStringLiteral("at most %1 MB allowed").arg(allowedMb);

    if (requestedMb > allowedMb) {
        emit failed(QStringLiteral("not enough memory: this instance is set to %1 MB, but %2. "
                                   "Lower the RAM setting and try again.")
                        .arg(requestedMb)
                        .arg(memoryHint));
        return;
    }

    QStringList args = buildJvmArgs();
    args << QStringLiteral("-cp") << QStringLiteral("__CLASSPATH__"); // filled by library resolver
    args << QStringLiteral("net.minecraft.launcher.Main");
    args << buildGameArgs();

    // Remember the numbers now: the settings can change while the game runs.
    m_requestedMb = requestedMb;
    m_totalMb = totalMb;
    m_allowedMb = allowedMb;
    m_exitReported = false;

    m_process = new QProcess(this);
    m_process->setWorkingDirectory(m_instance->gameRoot());
    // The JVM explains its own death on stderr; we need those words to tell
    // "out of memory" apart from an ordinary crash.
    m_process->setProcessChannelMode(QProcess::MergedChannels);
    connect(m_process, &QProcess::readyRead, this, [this]() { drainProcessOutput(); });
    connect(m_process, QOverload<int, QProcess::ExitStatus>::of(&QProcess::finished), this,
            [this](int exitCode, QProcess::ExitStatus status) {
                reportProcessExit(exitCode, status);
            });
    m_process->start(javaPath, args);

    if (!m_process->waitForStarted(10000)) {
        emit failed(QStringLiteral("the game did not start with %1 MB of memory (%2). "
                                   "Lower the RAM setting or close other programs, then try again.")
                        .arg(requestedMb)
                        .arg(memoryHint));
        return;
    }
    emit logLine(QStringLiteral("Launched game process (pid %1)").arg(m_process->processId()));
    emit readyForLaunch();
}

void LauncherPartLaunch::abort()
{
    if (m_process) {
        // A user-initiated stop is not a crash: no diagnosis for it.
        m_exitReported = true;
        m_process->kill();
    }
}

void LauncherPartLaunch::drainProcessOutput()
{
    if (!m_process)
        return;
    constexpr int kMaxOutputTail = 16 * 1024;
    m_outputTail += QString::fromLocal8Bit(m_process->readAll());
    if (m_outputTail.size() > kMaxOutputTail)
        m_outputTail = m_outputTail.right(kMaxOutputTail);
}

void LauncherPartLaunch::reportProcessExit(int exitCode, QProcess::ExitStatus status)
{
    if (m_exitReported)
        return;
    m_exitReported = true;

    // Pick up anything the process wrote just before it died.
    drainProcessOutput();

    if (status == QProcess::NormalExit && exitCode == 0)
        return; // the player closed the game: nothing to report

    const QString memoryMessage = GameExitDiagnostics::diagnoseExit(
        exitCode, m_outputTail, m_requestedMb, m_totalMb, m_allowedMb);
    if (!memoryMessage.isEmpty()) {
        // Same wording and numbers as the RAM panel, instead of a raw JVM trace.
        emit logLine(memoryMessage);
        emit failed(memoryMessage);
        return;
    }

    emit logLine(QStringLiteral("The game process exited (code %1)%2")
                     .arg(exitCode)
                     .arg(status == QProcess::CrashExit ? QStringLiteral(" after a crash")
                                                        : QString()));
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
