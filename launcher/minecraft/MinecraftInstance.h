// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher core. Concrete Minecraft instance type.
#pragma once

#include <QString>
#include <QStringList>
#include <memory>

#include "core/BaseInstance.h"

class PackProfile;
class LaunchTask;

// A Minecraft instance: manages the game root, version/loader composition and
// the launch pipeline for a local Minecraft 1.21.11 + Fabric instance.
class MinecraftInstance : public BaseInstance
{
    Q_OBJECT
public:
    explicit MinecraftInstance(QObject *parent = nullptr);
    ~MinecraftInstance() override;

    QStringList traits() const override;
    QString modsRoot() const override;

    // The .minecraft game root (instanceRoot()/.minecraft).
    QString gameRoot() const;
    // Directory for downloaded libraries (gameRoot()/libraries).
    QString libDir() const;
    // Directory for downloaded assets (gameRoot()/assets).
    QString assetsDir() const;
    // Directory for extracted natives (gameRoot()/bin/natives).
    QString nativesDir() const;

    // Version/loader composition for this instance.
    PackProfile *getPackProfile() const;

    // Load persisted metadata (instance.json, launcher.json, version.json)
    // from instanceRoot(). Returns false on error.
    bool load();

    // Persist instance metadata (instance.json, launcher.json, version.json).
    void saveNow() override;

    // Build the launch pipeline. Returns nullptr on failure.
    std::unique_ptr<LaunchTask> createLaunchTask(AuthSessionPtr session,
                                                 MinecraftTargetPtr target) override;

    // Resolved Java executable path (from launcher.json or auto-detected).
    QString javaPath() const;
    void setJavaPath(const QString &path);

    // Memory settings (MB).
    int maxMemoryMb() const;
    void setMaxMemoryMb(int mb);
    int minMemoryMb() const;
    void setMinMemoryMb(int mb);

private:
    std::unique_ptr<PackProfile> m_packProfile;
    QString m_javaPath;
    int m_maxMemoryMb = 4096;
    int m_minMemoryMb = 2048;
};
