// SPDX-License-Identifier: GPL-3.0-only
#include "MinecraftInstance.h"

#include <QDir>
#include <QJsonObject>

#include "PackProfile.h"
#include "core/InstancePersistence.h"
#include "core/SettingsStore.h"
#include "launch/LaunchTask.h"

#include <QStandardPaths>
#include "launch/LaunchStep.h"
#include "launch/steps/ClaimAuthSession.h"
#include "launch/steps/ExtractNatives.h"
#include "launch/steps/LauncherPartLaunch.h"
#include "launch/steps/ResolveAssets.h"
#include "launch/steps/ResolveLibraries.h"
#include "launch/steps/VerifyJava.h"

MinecraftInstance::MinecraftInstance(QObject *parent)
    : BaseInstance(parent), m_packProfile(std::make_unique<PackProfile>(this, this))
{
}

MinecraftInstance::~MinecraftInstance() = default;

QStringList MinecraftInstance::traits() const
{
    QStringList t;
    t << QStringLiteral("minecraft");
    if (m_packProfile->getComponentVersion(QStringLiteral("net.fabricmc.fabric-loader")).isEmpty())
        t << QStringLiteral("vanilla");
    else
        t << QStringLiteral("fabric") << QStringLiteral("modloader");
    return t;
}

QString MinecraftInstance::modsRoot() const
{
    return gameRoot() + QStringLiteral("/mods");
}

QString MinecraftInstance::gameRoot() const
{
    return instanceRoot() + QStringLiteral("/.minecraft");
}

QString MinecraftInstance::libDir() const
{
    return gameRoot() + QStringLiteral("/libraries");
}

QString MinecraftInstance::assetsDir() const
{
    return gameRoot() + QStringLiteral("/assets");
}

QString MinecraftInstance::nativesDir() const
{
    return gameRoot() + QStringLiteral("/bin/natives");
}

PackProfile *MinecraftInstance::getPackProfile() const
{
    return m_packProfile.get();
}

bool MinecraftInstance::load()
{
    const QString metaPath = instanceRoot() + QStringLiteral("/instance.json");
    QJsonObject metaObj;
    if (InstancePersistence::readJsonFile(metaPath, metaObj)) {
        const InstanceMetadata meta = InstancePersistence::metadataFromJson(metaObj);
        if (m_id.isEmpty())
            m_id = meta.id;
        if (m_name.isEmpty())
            m_name = meta.name;
    }

    const QString settingsPath = instanceRoot() + QStringLiteral("/launcher.json");
    QJsonObject settingsObj;
    if (InstancePersistence::readJsonFile(settingsPath, settingsObj)) {
        const LauncherSettings settings = InstancePersistence::settingsFromJson(settingsObj);
        m_javaPath = settings.javaPath;
        m_maxMemoryMb = settings.maxMemoryMb;
        m_minMemoryMb = settings.minMemoryMb;
    }

    // Honor the global RAM slider default (F5) when the instance has not set
    // an explicit value. The global settings file lives in the app config dir.
    QString globalPath = QStandardPaths::writableLocation(QStandardPaths::AppConfigLocation);
    if (!globalPath.isEmpty()) {
        globalPath += QStringLiteral("/launcher.json");
        QJsonObject globalObj;
        if (InstancePersistence::readJsonFile(globalPath, globalObj)) {
            SettingsStore::Settings global;
            // reuse InstancePersistence settings fields
            LauncherSettings gs = InstancePersistence::settingsFromJson(globalObj);
            if (gs.maxMemoryMb > 0)
                m_maxMemoryMb = gs.maxMemoryMb;
            if (gs.minMemoryMb > 0)
                m_minMemoryMb = gs.minMemoryMb;
            if (!gs.javaPath.isEmpty())
                m_javaPath = gs.javaPath;
        }
    }

    m_packProfile->load();
    return true;
}

void MinecraftInstance::saveNow()
{
    // instance.json
    InstanceMetadata meta;
    meta.id = id();
    meta.name = name();
    meta.type = QStringLiteral("MinecraftInstance");
    meta.traits = traits();
    const QString metaPath = instanceRoot() + QStringLiteral("/instance.json");
    InstancePersistence::writeJsonFile(metaPath, InstancePersistence::metadataToJson(meta));

    // launcher.json
    LauncherSettings settings;
    settings.javaPath = m_javaPath;
    settings.maxMemoryMb = m_maxMemoryMb;
    settings.minMemoryMb = m_minMemoryMb;
    const QString settingsPath = instanceRoot() + QStringLiteral("/launcher.json");
    InstancePersistence::writeJsonFile(settingsPath, InstancePersistence::settingsToJson(settings));

    // version.json
    m_packProfile->save();
}

std::unique_ptr<LaunchTask> MinecraftInstance::createLaunchTask(AuthSessionPtr session,
                                                                MinecraftTargetPtr target)
{
    auto task = std::make_unique<LaunchTask>(this);
    task->appendStep(new ClaimAuthSession(task.get(), session));
    task->appendStep(new ResolveLibraries(task.get(), this));
    task->appendStep(new ResolveAssets(task.get(), this));
    task->appendStep(new ExtractNatives(task.get(), this));
    task->appendStep(new VerifyJava(task.get(), this));
    task->appendStep(new LauncherPartLaunch(task.get(), this, session, target));
    return task;
}

QString MinecraftInstance::javaPath() const
{
    return m_javaPath;
}

void MinecraftInstance::setJavaPath(const QString &path)
{
    if (m_javaPath == path)
        return;
    m_javaPath = path;
    emit propertiesChanged(this);
}

int MinecraftInstance::maxMemoryMb() const
{
    return m_maxMemoryMb;
}

void MinecraftInstance::setMaxMemoryMb(int mb)
{
    if (m_maxMemoryMb == mb)
        return;
    m_maxMemoryMb = mb;
    emit propertiesChanged(this);
}

int MinecraftInstance::minMemoryMb() const
{
    return m_minMemoryMb;
}

void MinecraftInstance::setMinMemoryMb(int mb)
{
    if (m_minMemoryMb == mb)
        return;
    m_minMemoryMb = mb;
    emit propertiesChanged(this);
}
