// SPDX-License-Identifier: GPL-3.0-only
// Qt Test for the MainWindow GUI wiring: backend URL persistence, login gating
// and restoration of a stored session into the shared BackendAuthClient.
#include <QtTest>

#include <QDateTime>
#include <QFile>
#include <QLabel>
#include <QLineEdit>
#include <QPushButton>
#include <QStandardPaths>
#include <QTemporaryDir>

#include "auth/BackendAuthClient.h"
#include "auth/SessionManager.h"
#include "auth/TokenStorage.h"
#include "core/SettingsStore.h"
#include "gui/MainWindow.h"
#include "ui/SkinPreviewWidget.h"

class TestMainWindow : public QObject
{
    Q_OBJECT

private slots:
    void initTestCase()
    {
        QStandardPaths::setTestModeEnabled(true);
    }

    void init()
    {
        QFile::remove(MainWindow::defaultSettingsPath());
    }

    void cleanup()
    {
        QFile::remove(MainWindow::defaultSettingsPath());
    }

    void backendUrlValidationAndPersistence();
    void loginRequiresBackendUrl();
    void sessionRestorePropagatesToken();
};

void TestMainWindow::backendUrlValidationAndPersistence()
{
    MainWindow window;
    auto *edit = window.findChild<QLineEdit *>(QStringLiteral("backendUrlEdit"));
    auto *save = window.findChild<QPushButton *>(QStringLiteral("saveBackendButton"));
    auto *status = window.findChild<QLabel *>(QStringLiteral("statusLabel"));
    QVERIFY(edit != nullptr);
    QVERIFY(save != nullptr);
    QVERIFY(status != nullptr);
    QVERIFY(window.backendUrl().isEmpty());

    // Remote plaintext HTTP must be rejected by the GUI.
    edit->setText(QStringLiteral("http://example.com"));
    save->click();
    QVERIFY(window.backendUrl().isEmpty());
    QVERIFY(status->text().contains(QStringLiteral("Érvénytelen")));

    // A trailing slash is normalized away and the value is persisted.
    edit->setText(QStringLiteral("https://api.example.com/"));
    save->click();
    QCOMPARE(window.backendUrl(), QStringLiteral("https://api.example.com"));
    QVERIFY(QFile::exists(window.settingsPath()));

    SettingsStore::Settings stored;
    QVERIFY(SettingsStore::load(window.settingsPath(), stored));
    QCOMPARE(stored.backendUrl, QStringLiteral("https://api.example.com"));

    // A new window loads the saved backend URL from launcher.json.
    MainWindow reloaded;
    auto *reloadedEdit = reloaded.findChild<QLineEdit *>(QStringLiteral("backendUrlEdit"));
    QVERIFY(reloadedEdit != nullptr);
    QCOMPARE(reloadedEdit->text(), QStringLiteral("https://api.example.com"));
    QCOMPARE(reloaded.backendUrl(), QStringLiteral("https://api.example.com"));

    // A local backend may use plain HTTP (loopback only).
    edit->setText(QStringLiteral("http://127.0.0.1:8080"));
    save->click();
    QCOMPARE(window.backendUrl(), QStringLiteral("http://127.0.0.1:8080"));
}

void TestMainWindow::loginRequiresBackendUrl()
{
    MainWindow window;
    auto *edit = window.findChild<QLineEdit *>(QStringLiteral("backendUrlEdit"));
    auto *save = window.findChild<QPushButton *>(QStringLiteral("saveBackendButton"));
    auto *login = window.findChild<QPushButton *>(QStringLiteral("loginButton"));
    auto *authStatus = window.findChild<QLabel *>(QStringLiteral("authStatusLabel"));
    auto *skin = window.findChild<SkinPreviewWidget *>(QStringLiteral("skinPreview"));
    QVERIFY(edit != nullptr && save != nullptr && login != nullptr);
    QVERIFY(authStatus != nullptr && skin != nullptr);

    QVERIFY(!login->isEnabled());
    QVERIFY(!skin->isEnabled());
    QVERIFY(authStatus->text().contains(QStringLiteral("Nincs bejelentkezve")));

    edit->setText(QStringLiteral("https://api.example.com"));
    save->click();

    QVERIFY(login->isEnabled());
    // Uploading still needs a session, not just a backend URL.
    QVERIFY(!skin->isEnabled());
    QVERIFY(!window.isLoggedIn());
}

void TestMainWindow::sessionRestorePropagatesToken()
{
    QTemporaryDir dir;
    QVERIFY(dir.isValid());

    TokenStorage storage(dir.path());
    AuthSession session;
    session.username = QStringLiteral("alice");
    session.access_token = QStringLiteral("token-abc");
    session.refresh_token = QStringLiteral("refresh-abc");
    session.expires_at = QDateTime::currentMSecsSinceEpoch() + 3600 * 1000;
    session.is_online = true;
    QVERIFY(storage.save(session));

    BackendAuthClient client;
    SessionManager sessions(&client, &storage);

    bool called = false;
    bool restored = false;
    QString error;
    sessions.restore(QStringLiteral("alice"), [&](bool ok, const QString &err) {
        called = true;
        restored = ok;
        error = err;
    });

    QVERIFY2(called, qPrintable(error));
    QVERIFY(restored);
    QVERIFY(sessions.currentSession() != nullptr);
    QCOMPARE(sessions.currentSession()->username, QStringLiteral("alice"));
    // The HTTP client must carry the restored token, otherwise the skin upload
    // from the GUI would be unauthenticated.
    QCOMPARE(client.accessToken(), QStringLiteral("token-abc"));

    // Usernames are used as file names; traversal attempts must be refused.
    AuthSession unsafe;
    unsafe.username = QStringLiteral("../escape");
    unsafe.access_token = QStringLiteral("x");
    QVERIFY(!storage.save(unsafe));
    AuthSession loaded;
    QVERIFY(!storage.load(QStringLiteral("../escape"), loaded));
}

QTEST_MAIN(TestMainWindow)
#include "test_main_window.moc"
