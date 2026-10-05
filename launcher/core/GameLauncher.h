// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher core. Game process launcher (JVM bootstrap).
#pragma once

#include <QObject>
#include <QString>
#include <QStringList>
#include <QProcessEnvironment>

class QProcess;

// LaunchRequest describes a single JVM launch.
struct LaunchRequest
{
    QString javaPath;
    QString classpath;
    QString mainClass;
    QString nativesDir;
    QString workingDir;
    int maxMemoryMb = 0;
    int minMemoryMb = 0;
    // The backend session token is passed to the JVM via an environment
    // variable (not a command-line argument) so it does not appear in process
    // listings.
    QString sessionToken;
    QStringList extraJvmArgs;
    QStringList gameArgs;
};

// GameLauncher starts the Minecraft JVM via QProcess. The session token is
// delivered through the environment only; it is never placed on the command
// line where another local process could read it.
class GameLauncher : public QObject
{
    Q_OBJECT
public:
    explicit GameLauncher(QObject *parent = nullptr);

    // Validate the request and start the process. Returns false and emits
    // launchFailed if the request is not safe (no java path, no classpath,
    // empty main class, or an implausible java executable).
    bool start(const LaunchRequest &req);

    bool isRunning() const;
    void stop();

    // Test hook: expose the prepared environment so tests can assert the
    // token is only in the environment, never in the argument list.
    QProcessEnvironment preparedEnvironment(const LaunchRequest &req) const;
    QStringList preparedArguments(const LaunchRequest &req) const;

signals:
    void gameStarted(qint64 pid);
    void gameFinished(int exitCode);
    void launchFailed(const QString &reason);

private:
    QProcess *m_process = nullptr;
    QString m_sessionTokenEnvName;
};
