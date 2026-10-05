// SPDX-License-Identifier: GPL-3.0-only
#include <QCoreApplication>
#include <QJsonObject>
#include <QTemporaryDir>

#include "core/EditorConfig.h"
#include "core/InstancePersistence.h"

#include <cstdio>

static int failures = 0;

static void check(bool cond, const char *msg)
{
    if (!cond) {
        std::fprintf(stderr, "FAIL: %s\n", msg);
        ++failures;
    }
}

int main(int argc, char **argv)
{
    QCoreApplication app(argc, argv);

    // Defaults round-trip through JSON.
    EditorConfig def = EditorConfig::defaults();
    EditorConfig fromJson = EditorConfig::fromJson(def.toJson());
    check(fromJson.title == def.title, "title round-trip");
    check(fromJson.accentColor == def.accentColor, "accent round-trip");
    check(fromJson.showNewsTab == def.showNewsTab, "news flag round-trip");

    // Persistence round-trip to a local file.
    QTemporaryDir dir;
    check(dir.isValid(), "temp dir");
    const QString path = dir.path() + QStringLiteral("/editor.json");

    EditorConfig cfg = EditorConfig::defaults();
    cfg.title = QStringLiteral("Custom Title");
    cfg.accentColor = QStringLiteral("#FF8800");
    cfg.showSettingsTab = false;
    check(cfg.save(path), "save config");

    EditorConfig loaded;
    check(EditorConfig::load(path, loaded), "load config");
    check(loaded.title == QStringLiteral("Custom Title"), "loaded title");
    check(loaded.accentColor == QStringLiteral("#FF8800"), "loaded accent");
    check(loaded.showSettingsTab == false, "loaded settings flag");

    // Missing file fails cleanly.
    EditorConfig missing;
    check(!EditorConfig::load(dir.path() + QStringLiteral("/nope.json"), missing), "missing file rejected");

    if (failures == 0)
        std::printf("editor_config: all checks passed\n");
    return failures == 0 ? 0 : 1;
}
