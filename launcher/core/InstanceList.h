// SPDX-License-Identifier: GPL-3.0-only
#pragma once

#include <QAbstractListModel>
#include <QString>
#include <vector>
#include <functional>
#include <memory>

class BaseInstance;
struct InstanceMetadata;

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

using InstanceFactory = std::function<std::unique_ptr<BaseInstance>(const InstanceMetadata &meta,
                                                                    const QString &rootDir)>;

class InstanceList : public QAbstractListModel
{
    Q_OBJECT
public:
    explicit InstanceList(QObject *parent = nullptr);
    ~InstanceList() override;

    QString instancesDir() const;
    void setInstancesDir(const QString &dir);

    void setInstanceFactory(InstanceFactory factory);

    bool loadList();
    void saveNow();

    int rowCount(const QModelIndex &parent = QModelIndex()) const override;
    QVariant data(const QModelIndex &index, int role) const override;

    int count() const;
    BaseInstance *at(int i) const;
    BaseInstance *getInstanceById(const QString &id) const;

    void addInstance(std::unique_ptr<BaseInstance> instance);
    void removeInstance(int index);

    // Duplicate the instance at the given index into a new instance directory
    // and append it to the list. Returns the new instance pointer or nullptr.
    BaseInstance *duplicateInstance(int index);

signals:
    void instancesChanged();

private:
    QString m_instancesDir;
    InstanceFactory m_factory;
    std::vector<std::unique_ptr<BaseInstance>> m_instances;
    // C++17/20 modernization: use std::string_view for read-only string parameters
    // to avoid unnecessary QString copies in the instance list.
    void updateInstanceAt(int index, std::string_view newName);
};