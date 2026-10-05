// SPDX-License-Identifier: GPL-3.0-only
#include "GameLauncher.h"

#include "game/JavaDetector.h"

#include <QProcess>

namespace {
const char *const kTokenEnv = "MCSERVER_SESSION_TOKEN";
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
