// SPDX-License-Identifier: GPL-3.0-only
#include "ResolveAssets.h"

#include <QDir>

#include "minecraft/MinecraftInstance.h"

ResolveAssets::ResolveAssets(LaunchTask *parent, MinecraftInstance *instance)
    : LaunchStep(parent), m_instance(instance)
{
}

void ResolveAssets::executeTask()
{
    if (!m_instance) {
        emit failed(QStringLiteral("no instance for asset resolution"));
        return;
    }
    const QString assetsDir = m_instance->assetsDir();
    if (!QDir().mkpath(assetsDir)) {
        emit failed(QStringLiteral("could not create asset directory: %1").arg(assetsDir));
        return;
    }
    emit logLine(QStringLiteral("Asset directory ready: %1").arg(assetsDir));
    emit readyForLaunch();
}
