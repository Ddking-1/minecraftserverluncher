// SPDX-License-Identifier: GPL-3.0-only
#include <QApplication>
#include "core/EditorConfig.h"
#include "gui/EditorSettingsWidget.h"

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
    QApplication app(argc, argv);

    EditorConfig cfg = EditorConfig::defaults();
    cfg.title = QStringLiteral("MyServer");
    cfg.accentColor = QStringLiteral("#3FB950");
    cfg.showNewsTab = false;
    cfg.showSettingsTab = true;

    EditorSettingsWidget w;
    w.load(cfg);

    EditorConfig out = w.config();
    check(out.title == QStringLiteral("MyServer"), "title not preserved");
    check(out.accentColor == QStringLiteral("#3FB950"), "accent not preserved");
    check(out.showNewsTab == false, "news tab not toggled");
    check(out.showSettingsTab == true, "settings tab not toggled");
    check(EditorSettingsWidget::accentColorOf(out) == QStringLiteral("#3FB950"), "accentColorOf wrong");

    if (failures == 0)
        std::printf("editor_settings_widget: all checks passed\n");
    return failures == 0 ? 0 : 1;
}
