// SPDX-License-Identifier: GPL-3.0-only
// Minimal launcher entry point. UI is not yet implemented; this exercises the
// core instance list so the module graph links.
#include <QCoreApplication>
#include <QDebug>
#include <QDir>

#include "core/BaseInstance.h"
#include "core/InstanceList.h"
#include "minecraft/MinecraftInstance.h"

namespace {
std::unique_ptr<BaseInstance> makeInstance(const InstanceMetadata &meta, const QString &rootDir)
{
    auto inst = std::make_unique<MinecraftInstance>();
    inst->setId(meta.id);
    inst->setName(meta.name);
    inst->setInstanceRoot(rootDir);
    inst->load();
    return inst;
}
} // namespace

int main(int argc, char *argv[])
{
    QCoreApplication app(argc, argv);

    const QString instancesDir = argc > 1 ? QString::fromLocal8Bit(argv[1])
                                          : QStringLiteral("instances");

    InstanceList list;
    list.setInstancesDir(instancesDir);
    list.setInstanceFactory(&makeInstance);

    if (!list.loadList()) {
        qWarning() << "Could not load instances from" << instancesDir;
        return 1;
    }

    qInfo() << "Loaded" << list.count() << "instance(s) from" << instancesDir;
    for (int i = 0; i < list.count(); ++i) {
        BaseInstance *inst = list.at(i);
        qInfo() << " -" << inst->id() << ":" << inst->name();
    }
    return 0;
}
