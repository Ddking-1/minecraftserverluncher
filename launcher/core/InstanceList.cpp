// SPDX-License-Identifier: GPL-3.0-only
#include "InstanceList.h"
#include "BaseInstance.h"
#include "InstancePersistence.h"

#include <QDir>
#include <QDirIterator>
#include <QFile>
#include <QFileInfo>
#include <QJsonObject>

InstanceList::InstanceList(QObject *parent)
    : QAbstractListModel(parent)
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
    m_factory = factory;
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

    QStringList subDirs = dir.entryList(QDir::Dirs | QDir::NoDotAndDotDot);
    for (const QString &subDir : subDirs) {
        QString instancePath = dir.absoluteFilePath(subDir);
        QString metaPath = instancePath + "/instance.json";
        
        QJsonObject obj;
        if (!InstancePersistence::readJsonFile(metaPath, obj) || obj.isEmpty())
            continue;

        InstanceMetadata meta = InstancePersistence::metadataFromJson(obj);
        if (!m_factory)
            continue;

        auto inst = m_factory(meta, instancePath);
        if (inst) {
            m_instances.push_back(std::move(inst));
        }
    }

    endResetModel();
    emit instancesChanged();
    return true;
}

void InstanceList::saveNow()
{
    for (auto &inst : m_instances) {
        inst->saveNow();
    }
}

int InstanceList::rowCount(const QModelIndex &parent) const
{
    if (parent.isValid())
        return 0;
    return static_cast<int>(m_instances.size());
}

QVariant InstanceList::data(const QModelIndex &index, int role) const
{
    if (!index.isValid() || index.row() < 0 || index.row() >= static_cast<int>(m_instances.size()))
        return QVariant();

    const auto &inst = m_instances[index.row()];

    switch (role) {
    case InstanceRoles::InstancePointerRole:
        return QVariant::fromValue<BaseInstance *>(inst.get());
    case InstanceRoles::InstanceIDRole:
        return inst->id();
    case InstanceRoles::InstanceNameRole:
        return inst->name();
    case InstanceRoles::InstanceTypeRole:
        return QString("minecraft");
    case InstanceRoles::InstanceRunningRole:
        return inst->isRunning();
    default:
        return QVariant();
    }
}

int InstanceList::count() const
{
    return static_cast<int>(m_instances.size());
}

BaseInstance *InstanceList::at(int i) const
{
    if (i < 0 || i >= static_cast<int>(m_instances.size()))
        return nullptr;
    return m_instances[i].get();
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
    int idx = static_cast<int>(m_instances.size());
    beginInsertRows(QModelIndex(), idx, idx);
    m_instances.push_back(std::move(instance));
    endInsertRows();
    emit instancesChanged();
}

void InstanceList::removeInstance(int index)
{
    if (index < 0 || index >= static_cast<int>(m_instances.size()))
        return;
    beginRemoveRows(QModelIndex(), index, index);
    m_instances.erase(m_instances.begin() + index);
    endRemoveRows();
    emit instancesChanged();
}

BaseInstance *InstanceList::duplicateInstance(int index)
{
    if (index < 0 || index >= static_cast<int>(m_instances.size()))
        return nullptr;
    if (!m_factory || m_instancesDir.isEmpty())
        return nullptr;

    BaseInstance *src = m_instances[index].get();
    if (!src)
        return nullptr;

    // Build a new unique id.
    QString newId = src->id() + QStringLiteral("-copy");
    int suffix = 1;
    while (QDir(m_instancesDir + "/" + newId).exists()) {
        newId = src->id() + QStringLiteral("-copy%1").arg(suffix++);
    }

    QString srcDir = src->instanceRoot();
    QString newDir = m_instancesDir + "/" + newId;
    if (!QDir().mkpath(newDir))
        return nullptr;

    // Recursively copy the instance directory.
    QDirIterator it(srcDir, QDir::Files | QDir::Dirs | QDir::NoDotAndDotDot,
                    QDirIterator::Subdirectories);
    while (it.hasNext()) {
        it.next();
        const QString rel = srcDir.length() < it.filePath().length()
                                ? it.filePath().mid(srcDir.length() + 1)
                                : QString();
        if (rel.isEmpty())
            continue;
        const QString target = newDir + "/" + rel;
        if (it.fileInfo().isDir()) {
            QDir().mkpath(target);
        } else {
            QDir().mkpath(QFileInfo(target).absolutePath());
            QFile::copy(it.filePath(), target);
        }
    }

    // Update instance.json id if present.
    QString metaPath = newDir + "/instance.json";
    QJsonObject obj;
    if (InstancePersistence::readJsonFile(metaPath, obj)) {
        obj["id"] = newId;
        InstancePersistence::writeJsonFile(metaPath, obj);
    }

    InstanceMetadata meta;
    if (InstancePersistence::readJsonFile(metaPath, obj))
        meta = InstancePersistence::metadataFromJson(obj);
    meta.id = newId;

    auto inst = m_factory(meta, newDir);
    if (!inst)
        return nullptr;
    inst->setId(newId);
    inst->setInstanceRoot(newDir);

    int idx = static_cast<int>(m_instances.size());
    beginInsertRows(QModelIndex(), idx, idx);
    m_instances.push_back(std::move(inst));
    endInsertRows();
    emit instancesChanged();
    return m_instances.back().get();
}

void InstanceList::updateInstanceAt(int index, std::string_view newName)
{
    if (index < 0 || index >= static_cast<int>(m_instances.size()))
        return;
    auto *inst = m_instances[index].get();
    if (inst)
        inst->setName(QString::fromStdString(std::string(newName)));
    emit dataChanged(createIndex(index, 0), createIndex(index, 0));
}