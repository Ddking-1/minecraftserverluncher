// SPDX-License-Identifier: GPL-3.0-only
#include "GameLauncher.h"

#include "game/JavaDetector.h"

#include <QProcess>

namespace {
const char *const kTokenEnv = "MCSERVER_SESSION_TOKEN";

// Reject JVM/game arguments that could be used for injection: embedded
// newlines/NULs or dangerous agent/error-hook switches.
bool isSafeArg(const QString &arg)
{
    if (arg.isEmpty())
        return false;
    if (arg.contains(QChar('\n')) || arg.contains(QChar('\r')) || arg.contains(QChar('\0')))
        return false;
    const QString lower = arg.toLower();
    static const QStringList dangerous = {
        QStringLiteral("-javaagent"),
        QStringLiteral("-agentlib"),
        QStringLiteral("-agentpath"),
        QStringLiteral("-xx:onerror"),
        QStringLiteral("-xx:onoutofmemoryerror"),
    };
    for (const QString &d : dangerous)
        if (lower.startsWith(d))
            return false;
    return true;
}
} // namespace

GameLauncher::GameLauncher(QObject *parent)
    : QObject(parent), m_sessionTokenEnvName(QLatin1String(kTokenEnv))
{
}

bool GameLauncher::isRunning() const
{
    return m_process && m_process->state() != QProcess::NotRunning;
}

QStringList GameLauncher::preparedArguments(const LaunchRequest &req) const
{
    QStringList args;
    if (req.minMemoryMb > 0)
        args << QStringLiteral("-Xms%1M").arg(req.minMemoryMb);
    if (req.maxMemoryMb > 0)
        args << QStringLiteral("-Xmx%1M").arg(req.maxMemoryMb);
    args << req.extraJvmArgs;
    if (!req.nativesDir.isEmpty())
        args << QStringLiteral("-Djava.library.path=%1").arg(req.nativesDir);
    args << QStringLiteral("-cp") << req.classpath;
    args << req.mainClass;
    args << req.gameArgs;
    return args;
}

QProcessEnvironment GameLauncher::preparedEnvironment(const LaunchRequest &req) const
{
    QProcessEnvironment env = QProcessEnvironment::systemEnvironment();
    if (!req.sessionToken.isEmpty())
        env.insert(m_sessionTokenEnvName, req.sessionToken);
    return env;
}

bool GameLauncher::start(const LaunchRequest &req)
{
    // JVM arg / injection validation.
    if (req.minMemoryMb < 0 || req.maxMemoryMb < 0 || req.minMemoryMb > 65536 || req.maxMemoryMb > 65536) {
        emit launchFailed(QStringLiteral("implausible memory settings"));
        return false;
    }
    if (req.minMemoryMb > 0 && req.maxMemoryMb > 0 && req.minMemoryMb > req.maxMemoryMb) {
        emit launchFailed(QStringLiteral("min memory exceeds max memory"));
        return false;
    }
    if (req.mainClass.contains(QChar('\n')) || req.mainClass.contains(QChar('\r')) || req.mainClass.contains(QChar('\0'))) {
        emit launchFailed(QStringLiteral("unsafe main class"));
        return false;
    }
    if (req.classpath.contains(QChar('\n')) || req.classpath.contains(QChar('\r')) || req.classpath.contains(QChar('\0'))) {
        emit launchFailed(QStringLiteral("unsafe classpath"));
        return false;
    }
    for (const QString &a : req.extraJvmArgs) {
        if (!isSafeArg(a)) {
            emit launchFailed(QStringLiteral("unsafe jvm argument"));
            return false;
        }
    }
    for (const QString &a : req.gameArgs) {
        if (!isSafeArg(a)) {
            emit launchFailed(QStringLiteral("unsafe game argument"));
            return false;
        }
    }
    if (req.javaPath.isEmpty() || !JavaDetector::isPlausibleJava(req.javaPath)) {
        emit launchFailed(QStringLiteral("invalid or implausible java executable"));
        return false;
    }
    if (req.classpath.isEmpty()) {
        emit launchFailed(QStringLiteral("empty classpath"));
        return false;
    }
    if (req.mainClass.isEmpty()) {
        emit launchFailed(QStringLiteral("empty main class"));
        return false;
    }
    if (m_process) {
        emit launchFailed(QStringLiteral("a launch is already running"));
        return false;
    }

    m_process = new QProcess(this);
    m_process->setProgram(req.javaPath);
    m_process->setArguments(preparedArguments(req));
    m_process->setProcessEnvironment(preparedEnvironment(req));
    if (!req.workingDir.isEmpty())
        m_process->setWorkingDirectory(req.workingDir);

    connect(m_process, &QProcess::started, this, [this]() {
        emit gameStarted(m_process->processId());
    });
    connect(m_process, QOverload<int, QProcess::ExitStatus>::of(&QProcess::finished), this,
            [this](int code, QProcess::ExitStatus) {
                emit gameFinished(code);
                m_process->deleteLater();
                m_process = nullptr;
            });

    m_process->start();
    if (m_process->state() == QProcess::NotRunning) {
        const QString err = m_process->errorString();
        emit launchFailed(err.isEmpty() ? QStringLiteral("failed to start process") : err);
        m_process->deleteLater();
        m_process = nullptr;
        return false;
    }
    return true;
}

void GameLauncher::stop()
{
    if (m_process && m_process->state() != QProcess::NotRunning)
        m_process->terminate();
}
