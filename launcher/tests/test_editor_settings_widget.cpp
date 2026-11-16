// SPDX-License-Identifier: GPL-3.0-only
#include <QTest>
#include <QStandardPaths>

#include "core/EditorConfig.h"
#include "gui/EditorSettingsWidget.h"

class TestEditorSettingsWidget : public QObject
{
    Q_OBJECT
private slots:
    void initTestCase()
    {
        QStandardPaths::setTestModeEnabled(true);
    }

    void testConfigPersistence()
    {
        EditorConfig cfg = EditorConfig::defaults();
        cfg.title = QStringLiteral("MyServer");
        cfg.accentColor = QStringLiteral("#3FB950");
        cfg.showNewsTab = false;
        cfg.showSettingsTab = true;

        EditorSettingsWidget w;
        w.load(cfg);

        EditorConfig out = w.config();
        QCOMPARE(out.title, QStringLiteral("MyServer"));
        QCOMPARE(out.accentColor, QStringLiteral("#3FB950"));
        QCOMPARE(out.showNewsTab, false);
        QCOMPARE(out.showSettingsTab, true);
        QCOMPARE(EditorSettingsWidget::accentColorOf(out), QStringLiteral("#3FB950"));
    }
};

QTEST_MAIN(TestEditorSettingsWidget)
#include "test_editor_settings_widget.moc"
