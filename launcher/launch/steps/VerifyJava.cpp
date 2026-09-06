// SPDX-License-Identifier: GPL-3.0-only
#include "VerifyJava.h"

#include <QFileInfo>
#include <QProcess>

#include "launch/java/JavaVersion.h"
#include "minecraft/MinecraftInstance.h"

VerifyJava::VerifyJava(LaunchTask *parent, MinecraftInstance *instance)
    : LaunchStep(parent), m_instance(instance)
{
}

void VerifyJava::executeTask()
{
    if (!m_instance) {
        emit failed(QStringLiteral("no instance for Java verification"));
        return;
    }
    const QString javaPath = m_instance->javaPath();
    if (javaPath.isEmpty()) {
        emit failed(QStringLiteral("no Java runtime configured for this instance"));
        return;
    }
    if (!QFileInfo::exists(javaPath)) {
        emit failed(QStringLiteral("configured Java runtime not found: %1").arg(javaPath));
        return;
    }

    m_process = new QProcess(this);
    connect(m_process, QOverload<int, QProcess::ExitStatus>::of(&QProcess::finished), this,
            &VerifyJava::onProcessFinished);
    m_process->start(javaPath, { QStringLiteral("-version") });
}

void VerifyJava::abort()
{
    if (m_process)
        m_process->kill();
}

void VerifyJava::onProcessFinished(int exitCode)
{
    const QString output = QString::fromUtf8(m_process->readAllStandardError());
    m_process->deleteLater();
    m_process = nullptr;

    if (exitCode != 0) {
        emit failed(QStringLiteral("could not query Java version (exit %1)").arg(exitCode));
        return;
    }

    const JavaVersion v = JavaVersionParser::parseFromJavaVersionOutput(output);
    if (!v.isValid()) {
        emit failed(QStringLiteral("could not parse Java version output"));
        return;
    }
    emit logLine(QStringLiteral("Detected Java %1 (%2)").arg(v.major).arg(m_instance->javaPath()));
    emit readyForLaunch();
}
