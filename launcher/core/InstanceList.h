// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher core. QAbstractListModel over all managed instances.
#pragma once

#include <QAbstractListModel>
#include <QString>
#include <QVector>
#include <functional>
#include <memory>

#include "InstancePersistence.h"

class BaseInstance;

// Model roles exposed to the UI.
namespace InstanceRoles {
enum Role
{
    InstancePointerRole = Qt::UserRole + 1,
    InstanceIDRole,
    InstanceNameRole,
    InstanceTypeRole,
    InstanceRunningRole,
};
}

// Factory used to construct concrete instance types from on-disk metadata.
// Set by the application so that launcher/core does not depend on
// launcher/minecraft. The factory is responsible for setting the instance root
// and loading persisted state. Returns nullptr if the type is unknown.
using InstanceFactory = std::function<std::unique_ptr<BaseInstance>(const InstanceMetadata &meta,
                                                                    const QString &rootDir)>;

// Manages the set of instances on disk and exposes them to the UI as a
// QAbstractListModel. Loads/persists instance metadata from a root directory.
class InstanceList : public QAbstractListModel
{
    Q_OBJECT
public:
    explicit InstanceList(QObject *parent = nullptr);
    ~InstanceList() override;

    // Root directory that contains one sub-folder per instance.
    QString instancesDir() const;
    void setInstancesDir(const QString &dir);

    // Factory used to build concrete instances during loadList().
    void setInstanceFactory(InstanceFactory factory);

    // Discover and deserialize instances from instancesDir(). Returns false
    // if the directory cannot be read.
    bool loadList();

    // Persist all instance state to disk.
    void saveNow();

    int rowCount(const QModelIndex &parent = QModelIndex()) const override;
    QVariant data(const QModelIndex &index, int role) const override;

    int count() const;
    BaseInstance *at(int i) const;
    BaseInstance *getInstanceById(const QString &id) const;

    // Add an already-constructed instance to the list (takes ownership).
    void addInstance(std::unique_ptr<BaseInstance> instance);

signals:
    void instancesChanged();

private:
    QString m_instancesDir;
    InstanceFactory m_factory;
    QVector<std::unique_ptr<BaseInstance>> m_instances;
};
