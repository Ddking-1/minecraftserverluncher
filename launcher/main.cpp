// SPDX-License-Identifier: GPL-3.0-only
// Minimal launcher entry point. UI is not yet implemented; this exercises the
// core instance list so the module graph links.
#include <QCoreApplication>
#include <QDebug>
#include <QDir>
#include <QStringList>

#include "core/InstancePersistence.h"
#include "core/BaseInstance.h"
#include "core/InstanceList.h"
#include "core/Version.h"
#include "minecraft/MinecraftInstance.h"
#include "game/MinecraftLauncher.h"

namespace {

struct CliOptions
{
    QString instancesDir;
    QString filter;
    bool showVersion = false;
};

CliOptions parseArgs(const QStringList &args)
{
    CliOptions options;
    options.instancesDir = QStringLiteral("instances");

    for (int i = 0; i < args.size(); ++i) {
        const QString arg = args.at(i);
        if (arg == QLatin1String("--version") || arg == QLatin1String("-v")
            || arg == QLatin1String("--help") || arg == QLatin1String("-h")) {
            options.showVersion = true;
        } else if (arg == QLatin1String("--filter") || arg == QLatin1String("-f")) {
            if (i + 1 < args.size())
                options.filter = args.at(++i);
        } else if (arg.startsWith(QLatin1String("--filter="))) {
            options.filter = arg.mid(static_cast<int>(qstrlen("--filter=")));
        } else if (!arg.startsWith(QLatin1Char('-'))) {
            options.instancesDir = arg;
        }
    }
    return options;
}

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

    const CliOptions options = parseArgs(app.arguments().mid(1));

    if (options.showVersion) {
        qInfo().noquote() << LauncherVersion::kName << LauncherVersion::kVersion
                          << "(" << LauncherVersion::kBuildInfo << ")";
        return 0;
    }

    InstanceList list;
    list.setInstancesDir(options.instancesDir);
    list.setInstanceFactory(&makeInstance);

    if (!list.loadList()) {
        qWarning() << "Could not load instances from" << options.instancesDir;
        return 1;
    }

    int shown = 0;
    for (int i = 0; i < list.count(); ++i) {
        BaseInstance *inst = list.at(i);
        if (!options.filter.isEmpty()
            && !inst->id().contains(options.filter, Qt::CaseInsensitive)
            && !inst->name().contains(options.filter, Qt::CaseInsensitive)) {
            continue;
        }
        qInfo() << " -" << inst->id() << ":" << inst->name();
        ++shown;
    }

    qInfo() << "Loaded" << list.count() << "instance(s) from" << options.instancesDir
            << "- showing" << shown;
    return 0;
}
