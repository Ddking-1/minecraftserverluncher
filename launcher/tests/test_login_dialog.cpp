// SPDX-License-Identifier: GPL-3.0-only
#include <QTest>
#include <QStandardPaths>

#include "gui/LoginDialog.h"

class TestLoginDialog : public QObject
{
    Q_OBJECT
private slots:
    void initTestCase()
    {
        QStandardPaths::setTestMode(true);
    }

    void testInitialState()
    {
        LoginDialog dlg;
        QVERIFY(dlg.username().isEmpty());
        QVERIFY(dlg.password().isEmpty());
        QVERIFY(!dlg.windowTitle().isEmpty());
    }
};

QTEST_MAIN(TestLoginDialog)
#include "test_login_dialog.moc"
