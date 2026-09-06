// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher core. Version/loader composition for a Minecraft
// instance (minimal adaptation of Prism's PackProfile).
#pragma once

#include <QAbstractListModel>
#include <QString>
#include <QVector>

class MinecraftInstance;

// One versioned component of an instance (e.g. the game jar, a loader).
struct Component
{
    QString uid;      // e.g. "net.minecraft" or "net.fabricmc.fabric-loader"
    QString version;  // e.g. "1.21.11" or "0.19.2"
    QString name;     // human-readable name
};

namespace PackProfileRoles {
enum Role
{
    ComponentUidRole = Qt::UserRole + 1,
    ComponentVersionRole,
    ComponentNameRole,
};
}

// QAbstractListModel of the version components that make up an instance.
// Persisted as version.json in the instance folder.
class PackProfile : public QAbstractListModel
{
    Q_OBJECT
public:
    explicit PackProfile(MinecraftInstance *instance, QObject *parent = nullptr);

    int rowCount(const QModelIndex &parent = QModelIndex()) const override;
    QVariant data(const QModelIndex &index, int role) const override;

    // Replace the full component list (e.g. after loading version.json).
    void setComponents(const QVector<Component> &components);
    const QVector<Component> &components() const;

    // Version of a given component uid, or empty if absent.
    QString getComponentVersion(const QString &uid) const;
    // Set (or add) the version of a component uid.
    void setComponentVersion(const QString &uid, const QString &version);

    // Load components from the instance's version.json. Returns false on error.
    bool load();
    // Persist components to the instance's version.json.
    bool save() const;

private:
    MinecraftInstance *m_instance;
    QVector<Component> m_components;
};
