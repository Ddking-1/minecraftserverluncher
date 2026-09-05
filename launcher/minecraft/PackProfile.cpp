// SPDX-License-Identifier: GPL-3.0-only
#include "PackProfile.h"

#include <QDir>
#include <QJsonArray>
#include <QJsonObject>

#include "MinecraftInstance.h"
#include "core/InstancePersistence.h"

PackProfile::PackProfile(MinecraftInstance *instance, QObject *parent)
    : QAbstractListModel(parent), m_instance(instance)
{
}

int PackProfile::rowCount(const QModelIndex &parent) const
{
    if (parent.isValid())
        return 0;
    return m_components.size();
}

QVariant PackProfile::data(const QModelIndex &index, int role) const
{
    if (!index.isValid() || index.row() < 0 || index.row() >= m_components.size())
        return QVariant();

    const Component &c = m_components.at(index.row());
    switch (role) {
    case PackProfileRoles::ComponentUidRole:
        return c.uid;
    case PackProfileRoles::ComponentVersionRole:
        return c.version;
    case PackProfileRoles::ComponentNameRole:
        return c.name;
    default:
        return QVariant();
    }
}

void PackProfile::setComponents(const QVector<Component> &components)
{
    beginResetModel();
    m_components = components;
    endResetModel();
}

const QVector<Component> &PackProfile::components() const
{
    return m_components;
}

QString PackProfile::getComponentVersion(const QString &uid) const
{
    for (const Component &c : m_components) {
        if (c.uid == uid)
            return c.version;
    }
    return QString();
}

void PackProfile::setComponentVersion(const QString &uid, const QString &version)
{
    for (Component &c : m_components) {
        if (c.uid == uid) {
            c.version = version;
            emit dataChanged(index(0), index(rowCount() - 1));
            return;
        }
    }
    Component c;
    c.uid = uid;
    c.version = version;
    c.name = uid;
    beginInsertRows(QModelIndex(), m_components.size(), m_components.size());
    m_components.push_back(c);
    endInsertRows();
}

bool PackProfile::load()
{
    const QString path = m_instance->instanceRoot() + QStringLiteral("/version.json");
    QJsonObject obj;
    if (!InstancePersistence::readJsonFile(path, obj))
        return false;

    QVector<Component> components;
    const QJsonArray arr = obj.value(QStringLiteral("components")).toArray();
    for (const QJsonValue &v : arr) {
        const QJsonObject co = v.toObject();
        Component c;
        c.uid = co.value(QStringLiteral("uid")).toString();
        c.version = co.value(QStringLiteral("version")).toString();
        c.name = co.value(QStringLiteral("name")).toString(c.uid);
        if (!c.uid.isEmpty())
            components.push_back(c);
    }
    setComponents(components);
    return true;
}

bool PackProfile::save() const
{
    QJsonArray arr;
    for (const Component &c : m_components) {
        QJsonObject co;
        co.insert(QStringLiteral("uid"), c.uid);
        co.insert(QStringLiteral("version"), c.version);
        co.insert(QStringLiteral("name"), c.name);
        arr.append(co);
    }
    QJsonObject root;
    root.insert(QStringLiteral("components"), arr);

    const QString path = m_instance->instanceRoot() + QStringLiteral("/version.json");
    return InstancePersistence::writeJsonFile(path, root);
}
