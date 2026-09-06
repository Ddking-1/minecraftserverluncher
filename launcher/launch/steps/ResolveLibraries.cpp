// SPDX-License-Identifier: GPL-3.0-only
#include "ResolveLibraries.h"

#include <QDir>

#include "minecraft/MinecraftInstance.h"

ResolveLibraries::ResolveLibraries(LaunchTask *parent, MinecraftInstance *instance)
    : LaunchStep(parent), m_instance(instance)
{
}

void ResolveLibraries::executeTask()
{
    if (!m_instance) {
        emit failed(QStringLiteral("no instance for library resolution"));
        return;
    }
    const QString libDir = m_instance->libDir();
    if (!QDir().mkpath(libDir)) {
        emit failed(QStringLiteral("could not create library directory: %1").arg(libDir));
        return;
    }
    emit logLine(QStringLiteral("Library directory ready: %1").arg(libDir));
    emit readyForLaunch();
}
