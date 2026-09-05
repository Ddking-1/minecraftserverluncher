// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher core. Abstract base class for a managed instance.
#pragma once

#include <QObject>
#include <QString>
#include <QStringList>
#include <memory>

#include "InstanceTypes.h"

class LaunchTask;

// Abstract base for every instance type managed by the launcher.
// Concrete types (e.g. MinecraftInstance) live in launcher/minecraft.
class BaseInstance : public QObject
{
    Q_OBJECT
public:
    explicit BaseInstance(QObject *parent = nullptr);
    ~BaseInstance() override;

    // Unique, stable identifier for this instance (folder name).
    QString id() const;
    void setId(const QString &id);

    // Root directory that holds this instance's data.
    QString instanceRoot() const;
    void setInstanceRoot(const QString &root);

    // Human-readable display name.
    QString name() const;
    void setName(const QString &name);

    // Traits describing this instance (e.g. "fabric", "modloader").
    virtual QStringList traits() const = 0;

    // Directory that holds loader-specific mods.
    virtual QString modsRoot() const = 0;

    // Whether a launch is currently in progress for this instance.
    bool isRunning() const;
    void setRunning(bool running);

    // Persist instance state to disk.
    virtual void saveNow() = 0;

    // Build the launch pipeline for this instance. Returns nullptr on failure.
    virtual std::unique_ptr<LaunchTask> createLaunchTask(AuthSessionPtr session,
                                                         MinecraftTargetPtr target) = 0;

signals:
    // Emitted when any user-visible property changed.
    void propertiesChanged(BaseInstance *instance);
    // Emitted when the running state changed.
    void runningStatusChanged(bool running);

protected:
    QString m_id;
    QString m_rootDir;
    QString m_name;
    bool m_isRunning = false;
};
