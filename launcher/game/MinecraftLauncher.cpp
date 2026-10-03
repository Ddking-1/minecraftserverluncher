#include "MinecraftLauncher.h"
#include <QDir>
#include <QDebug>

MinecraftLauncher::MinecraftLauncher(QObject *parent)
    : QObject(parent),
      m_process(new QProcess(this))
{
    connect(m_process, &QProcess::started, this, &MinecraftLauncher::onProcessStarted);
    connect(m_process, &QProcess::finished, this, &MinecraftLauncher::onProcessFinished);
    connect(m_process, &QProcess::errorOccurred, this, &MinecraftLauncher::onProcessError);
}

void MinecraftLauncher::launchGame(const QString &javaPath, const QString &username, const QString &sessionToken, const QString &gameDir)
{
    if (javaPath.isEmpty() || gameDir.isEmpty()) {
        emit launchError("Invalid parameters");
        return;
    }

    QDir dir(gameDir);
    if (!dir.exists()) {
        emit launchError("Game directory does not exist: " + gameDir);
        return;
    }

    QStringList arguments;
    arguments << "-cp" << "net.minecraft.client.main.Main";
    arguments << "--username" << username;
    arguments << "--uuid" << sessionToken; // Using token as UUID for simplicity
    arguments << "--accessToken" << sessionToken;
    arguments << "--gameDir" << gameDir;

    m_process->setProgram(javaPath);
    m_process->setArguments(arguments);

    m_process->start();
    if (!m_process->waitForStarted(5000)) {
        emit launchError("Failed to start process");
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
        errorMsg = "Failed to start process";
        break;
    case QProcess::Crashed:
        errorMsg = "Process crashed";
        break;
    case QProcess::Timedout:
        errorMsg = "Process timed out";
        break;
    case QProcess::UnknownError:
        errorMsg = "Unknown error occurred";
        break;
    default:
        errorMsg = "Process error: " + QString::number(error);
        break;
    }
    
    emit launchError(errorMsg);
}