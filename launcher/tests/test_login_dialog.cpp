// SPDX-License-Identifier: GPL-3.0-only
#include <QTest>
#include <QStandardPaths>

#include <QLabel>
#include <QLineEdit>
#include <QPushButton>

#include "gui/LoginDialog.h"

class TestLoginDialog : public QObject
{
    Q_OBJECT
private slots:
    void initTestCase()
    {
        QStandardPaths::setTestModeEnabled(true);
    }

    void testInitialState()
    {
        LoginDialog dlg;
        QVERIFY(dlg.username().isEmpty());
        QVERIFY(dlg.password().isEmpty());
        QVERIFY(!dlg.windowTitle().isEmpty());
    }

    void testUsernamePrefill()
    {
        LoginDialog dlg;
        dlg.setUsername(QStringLiteral("alice"));
        QCOMPARE(dlg.username(), QStringLiteral("alice"));
    }

    void testEmptyFieldsAreRejected()
    {
        LoginDialog dlg;
        auto *ok = dlg.findChild<QPushButton *>(QStringLiteral("loginOkButton"));
        auto *error = dlg.findChild<QLabel *>(QStringLiteral("loginErrorLabel"));
        QVERIFY(ok != nullptr);
        QVERIFY(error != nullptr);

        ok->click();

        QCOMPARE(dlg.result(), static_cast<int>(QDialog::Rejected));
        QVERIFY(!error->text().isEmpty());
    }

    void testShortPasswordIsRejected()
    {
        LoginDialog dlg;
        auto *username = dlg.findChild<QLineEdit *>(QStringLiteral("usernameEdit"));
        auto *password = dlg.findChild<QLineEdit *>(QStringLiteral("passwordEdit"));
        auto *ok = dlg.findChild<QPushButton *>(QStringLiteral("loginOkButton"));
        QVERIFY(username != nullptr && password != nullptr && ok != nullptr);

        username->setText(QStringLiteral("alice"));
        password->setText(QStringLiteral("short"));
        ok->click();

        QCOMPARE(dlg.result(), static_cast<int>(QDialog::Rejected));
        QVERIFY(password->text().size() < 8);
    }

    void testValidCredentialsAccepted()
    {
        LoginDialog dlg;
        dlg.setUsername(QStringLiteral("alice"));
        auto *password = dlg.findChild<QLineEdit *>(QStringLiteral("passwordEdit"));
        auto *ok = dlg.findChild<QPushButton *>(QStringLiteral("loginOkButton"));
        QVERIFY(password != nullptr && ok != nullptr);

        password->setText(QStringLiteral("supersecret"));
        ok->click();

        QCOMPARE(dlg.result(), static_cast<int>(QDialog::Accepted));
        QCOMPARE(dlg.username(), QStringLiteral("alice"));
        QCOMPARE(dlg.password(), QStringLiteral("supersecret"));
    }
};

QTEST_MAIN(TestLoginDialog)
#include "test_login_dialog.moc"
