// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher core. Launch step that starts the Minecraft JVM.
#pragma once

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

private:
    QStringList buildJvmArgs() const;
    QStringList buildGameArgs() const;

    MinecraftInstance *m_instance;
    AuthSessionPtr m_session;
    MinecraftTargetPtr m_target;
    QProcess *m_process = nullptr;
};
