#include "MinecraftLauncher.h"

#include "core/FriendlyError.h"

#include <QDebug>
#include <QDir>

MinecraftLauncher::MinecraftLauncher(QObject *parent)
    : QObject(parent),
      m_process(new QProcess(this))
{
    connect(m_process, &QProcess::started, this, &MinecraftLauncher::onProcessStarted);
    connect(m_process, &QProcess::finished, this, &MinecraftLauncher::onProcessFinished);
    connect(m_process, &QProcess::errorOccurred, this, &MinecraftLauncher::onProcessError);
}

void MinecraftLauncher::launchGame(const QString &javaPath, const QString &username, const QString &uuid, const QString &accessToken, const QString &gameDir)
{
    if (javaPath.isEmpty() || gameDir.isEmpty()) {
        emit launchError(FriendlyError::explain(QStringLiteral("Invalid parameters")));
        return;
    }

    QDir dir(gameDir);
    if (!dir.exists()) {
        emit launchError(FriendlyError::explain(QStringLiteral("Game directory does not exist: ") +
                                                gameDir));
        return;
    }

    QStringList arguments;
    arguments << "-cp" << "net.minecraft.client.main.Main";
    arguments << "--username" << username;
    arguments << "--uuid" << uuid;
    arguments << "--accessToken" << accessToken;
    arguments << "--gameDir" << gameDir;

    m_process->setProgram(javaPath);
    m_process->setArguments(arguments);

    m_process->start();
    if (!m_process->waitForStarted(5000)) {
        emit launchError(FriendlyError::explain(QStringLiteral("Failed to start process")));
        return;
    }
}

void MinecraftLauncher::onProcessStarted()
{
    emit gameStarted();
}

void MinecraftLauncher::onProcessFinished(int exitCode, QProcess::ExitStatus exitStatus)
{
    Q_UNUSED(exitStatus);
    emit gameFinished(exitCode);
}

void MinecraftLauncher::onProcessError(QProcess::ProcessError error)
{
    QString errorMsg;
    switch (error) {
    case QProcess::FailedToStart:
        errorMsg = QStringLiteral("Failed to start process");
        break;
    case QProcess::Crashed:
        errorMsg = QStringLiteral("Process crashed");
        break;
    case QProcess::Timedout:
        errorMsg = QStringLiteral("Process timed out");
        break;
    case QProcess::UnknownError:
        errorMsg = QStringLiteral("Unknown error occurred");
        break;
    default:
        errorMsg = QStringLiteral("Process error: %1").arg(int(error));
        break;
    }

    // The technical wording goes to the log; the player gets a plain sentence.
    qWarning() << "MinecraftLauncher process error:" << errorMsg;
    emit launchError(FriendlyError::explain(errorMsg));
}