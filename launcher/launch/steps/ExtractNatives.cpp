// SPDX-License-Identifier: GPL-3.0-only
#include "ExtractNatives.h"

#include <QDir>

#include "minecraft/MinecraftInstance.h"

ExtractNatives::ExtractNatives(LaunchTask *parent, MinecraftInstance *instance)
    : LaunchStep(parent), m_instance(instance)
{
}

void ExtractNatives::executeTask()
{
    if (!m_instance) {
        emit failed(QStringLiteral("no instance for native extraction"));
        return;
    }
    const QString nativesDir = m_instance->nativesDir();
    if (!QDir().mkpath(nativesDir)) {
        emit failed(QStringLiteral("could not create natives directory: %1").arg(nativesDir));
        return;
    }
    emit logLine(QStringLiteral("Natives directory ready: %1").arg(nativesDir));
    emit readyForLaunch();
}
