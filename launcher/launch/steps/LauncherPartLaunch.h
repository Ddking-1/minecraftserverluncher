// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher core. Launch step that starts the Minecraft JVM.
#pragma once

#include <QProcess>
#include <QString>
#include <QStringList>

#include "launch/LaunchStep.h"
#include "core/InstanceTypes.h"

class MinecraftInstance;
class QProcess;

// Assembles the JVM command line and launches the game process. The game is
// started with the backend session identity (no Mojang/MSA tokens). This step
// does not wait for the game to exit; it reports readyForLaunch() once the
// process has been started successfully.
class LauncherPartLaunch : public LaunchStep
{
    Q_OBJECT
public:
    LauncherPartLaunch(LaunchTask *parent, MinecraftInstance *instance, AuthSessionPtr session,
                       MinecraftTargetPtr target);

    void executeTask() override;
    void abort() override;

    // Pure helpers: exposed so the assembled command line can be regression
    // tested without starting a JVM.
    [[nodiscard]] QStringList buildJvmArgs() const;
    [[nodiscard]] QStringList buildGameArgs() const;

private:
    // Keep the tail of the JVM output so an exit can be diagnosed (bounded, the
    // game may print a lot).
    void drainProcessOutput();
    // Report a process that started and then ended. Clean exits stay silent;
    // memory exhaustion gets the RAM panel's numbers, everything else keeps the
    // generic message.
    void reportProcessExit(int exitCode, QProcess::ExitStatus status);

    MinecraftInstance *m_instance;
    AuthSessionPtr m_session;
    MinecraftTargetPtr m_target;
    QProcess *m_process = nullptr;

    // Memory facts captured at launch time, used to explain a later exit.
    int m_requestedMb = 0;
    int m_totalMb = 0;
    int m_allowedMb = 0;
    QString m_outputTail;
    bool m_exitReported = false;
};
