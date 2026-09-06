// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher core. Launch step that verifies the Java runtime.
#pragma once

#include "launch/LaunchStep.h"

class MinecraftInstance;
class QProcess;

// Verifies the instance's configured Java runtime exists and reports its
// version. The project targets Java 24. No external Microsoft checks.
class VerifyJava : public LaunchStep
{
    Q_OBJECT
public:
    VerifyJava(LaunchTask *parent, MinecraftInstance *instance);

    void executeTask() override;
    void abort() override;

private:
    void onProcessFinished(int exitCode);

    MinecraftInstance *m_instance;
    QProcess *m_process = nullptr;
};
