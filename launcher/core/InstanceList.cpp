// SPDX-License-Identifier: GPL-3.0-only
#include "InstanceList.h"

#include <QDir>
#include <QJsonObject>

#include "BaseInstance.h"

InstanceList::InstanceList(QObject *parent) : QAbstractListModel(parent)
{
}

InstanceList::~InstanceList() = default;

QString InstanceList::instancesDir() const
{
    return m_instancesDir;
}

void InstanceList::setInstancesDir(const QString &dir)
{
    m_instancesDir = dir;
}

void InstanceList::setInstanceFactory(InstanceFactory factory)
{
    m_factory = std::move(factory);
}

bool InstanceList::loadList()
{
    beginResetModel();
    m_instances.clear();

    QDir dir(m_instancesDir);
    if (!dir.exists()) {
        endResetModel();
        return false;
    }

    const QStringList subdirs = dir.entryList(QDir::Dirs | QDir::NoDotAndDotDot, QDir::Name);
    for (const QString &sub : subdirs) {
        const QString metaPath = dir.filePath(sub + QStringLiteral("/instance.json"));
        QJsonObject obj;
        if (!InstancePersistence::readJsonFile(metaPath, obj))
            continue; // not a valid instance folder

        const InstanceMetadata meta = InstancePersistence::metadataFromJson(obj);
        if (meta.id.isEmpty())
            continue;

        if (!m_factory) {
            endResetModel();
            return false;
        }
        std::unique_ptr<BaseInstance> inst = m_factory(meta, dir.filePath(sub));
        if (!inst)
            continue;
        m_instances.push_back(std::move(inst));
    }

    endResetModel();
    emit instancesChanged();
    return true;
}

void InstanceList::saveNow()
{
    for (auto &inst : m_instances)
        inst->saveNow();
}

int InstanceList::rowCount(const QModelIndex &parent) const
{
    if (parent.isValid())
        return 0;
    return m_instances.size();
}

QVariant InstanceList::data(const QModelIndex &index, int role) const
{
    if (!index.isValid() || index.row() < 0 || index.row() >= m_instances.size())
        return QVariant();

    BaseInstance *inst = m_instances.at(index.row()).get();
    switch (role) {
    case InstanceRoles::InstancePointerRole:
        return QVariant::fromValue(inst);
    case InstanceRoles::InstanceIDRole:
        return inst->id();
    case InstanceRoles::InstanceNameRole:
        return inst->name();
    case InstanceRoles::InstanceTypeRole:
        return inst->traits().join(QLatin1Char(','));
    case InstanceRoles::InstanceRunningRole:
        return inst->isRunning();
    default:
        return QVariant();
    }
}

int InstanceList::count() const
{
    return m_instances.size();
}

BaseInstance *InstanceList::at(int i) const
{
    if (i < 0 || i >= m_instances.size())
        return nullptr;
    return m_instances.at(i).get();
}

BaseInstance *InstanceList::getInstanceById(const QString &id) const
{
    for (const auto &inst : m_instances) {
        if (inst->id() == id)
            return inst.get();
    }
    return nullptr;
}

void InstanceList::addInstance(std::unique_ptr<BaseInstance> instance)
{
    beginInsertRows(QModelIndex(), m_instances.size(), m_instances.size());
    m_instances.push_back(std::move(instance));
    endInsertRows();
    emit instancesChanged();
}
