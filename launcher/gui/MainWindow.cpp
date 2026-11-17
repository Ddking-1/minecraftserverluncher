// SPDX-License-Identifier: GPL-3.0-only
#include "MainWindow.h"

#include "auth/BackendAuthClient.h"
#include "auth/SessionManager.h"
#include "auth/TokenStorage.h"
#include "gui/CrashReportDialog.h"
#include "gui/LoginDialog.h"
#include "ui/SkinPreviewWidget.h"

#include <QCheckBox>
#include <QFileDialog>
#include <QFile>
#include <QJsonDocument>
#include <QJsonObject>
#include <QNetworkAccessManager>
#include <QNetworkReply>
#include <QNetworkRequest>
#include <QSlider>

#include <QDir>
#include <QGroupBox>
#include <QHBoxLayout>
#include <QLabel>
#include <QLineEdit>
#include <QPushButton>
#include <QStandardPaths>
#include <QStatusBar>
#include <QSysInfo>
#include <QVBoxLayout>

namespace {

QString appDataDirectory()
{
    QString dir = QStandardPaths::writableLocation(QStandardPaths::AppDataLocation);
    if (dir.isEmpty())
        dir = QDir::homePath() + QStringLiteral("/.minecraftserverlauncher");
    return dir;
}

} // namespace

QString MainWindow::defaultSettingsPath()
{
    QString dir = QStandardPaths::writableLocation(QStandardPaths::AppConfigLocation);
    if (dir.isEmpty())
        dir = QDir::homePath() + QStringLiteral("/.minecraftserverlauncher");
    return dir + QStringLiteral("/launcher.json");
}

MainWindow::MainWindow(QWidget *parent)
    : QMainWindow(parent)
    , m_settingsPath(defaultSettingsPath())
    , m_tokenStorage(std::make_unique<TokenStorage>(appDataDirectory()))
    , m_authClient(new BackendAuthClient(this))
    , m_sessions(new SessionManager(m_authClient, m_tokenStorage.get(), this))
{
    buildUi();
    loadSettings();
    m_skinPreview->setAuthClient(m_authClient);
    updateAuthUi();
    attemptSessionRestore();
}

MainWindow::~MainWindow() = default;

void MainWindow::buildUi()
{
    auto *central = new QWidget(this);
    auto *layout = new QVBoxLayout(central);

    auto *backendBox = new QGroupBox(tr("Backend"), central);
    auto *backendLayout = new QHBoxLayout(backendBox);
    m_backendUrlEdit = new QLineEdit(backendBox);
    m_backendUrlEdit->setObjectName(QStringLiteral("backendUrlEdit"));
    m_backendUrlEdit->setPlaceholderText(QStringLiteral("https://api.example.com"));
    m_backendUrlEdit->setClearButtonEnabled(true);
    m_saveBackendButton = new QPushButton(tr("Mentés"), backendBox);
    m_saveBackendButton->setObjectName(QStringLiteral("saveBackendButton"));
    backendLayout->addWidget(m_backendUrlEdit, 1);
    backendLayout->addWidget(m_saveBackendButton);
    layout->addWidget(backendBox);

    auto *accountBox = new QGroupBox(tr("Fiók"), central);
    auto *accountLayout = new QHBoxLayout(accountBox);
    m_authStatusLabel = new QLabel(tr("Nincs bejelentkezve"), accountBox);
    m_authStatusLabel->setObjectName(QStringLiteral("authStatusLabel"));
    m_loginButton = new QPushButton(tr("Bejelentkezés"), accountBox);
    m_loginButton->setObjectName(QStringLiteral("loginButton"));
    m_logoutButton = new QPushButton(tr("Kijelentkezés"), accountBox);
    m_logoutButton->setObjectName(QStringLiteral("logoutButton"));
    accountLayout->addWidget(m_authStatusLabel, 1);
    accountLayout->addWidget(m_loginButton);
    accountLayout->addWidget(m_logoutButton);
    layout->addWidget(accountBox);

    // RAM slider (F5)
    auto *ramBox = new QGroupBox(tr("RAM (MB)"), central);
    auto *ramLayout = new QHBoxLayout(ramBox);
    m_ramLabel = new QLabel(tr("4096"), ramBox);
    m_ramSlider = new QSlider(Qt::Horizontal, ramBox);
    m_ramSlider->setObjectName(QStringLiteral("ramSlider"));
    m_ramSlider->setRange(512, 65536);
    m_ramSlider->setValue(m_settings.maxMemoryMb);
    m_ramSlider->setSingleStep(256);
    ramLayout->addWidget(m_ramSlider, 1);
    ramLayout->addWidget(m_ramLabel);
    layout->addWidget(ramBox);

    // Offline mode (F6)
    m_offlineModeCheckBox = new QCheckBox(tr("Offline mód"), central);
    m_offlineModeCheckBox->setObjectName(QStringLiteral("offlineModeCheckBox"));
    layout->addWidget(m_offlineModeCheckBox);

    // Action buttons (F2, F3, F10)
    auto *actionLayout = new QHBoxLayout();
    m_newsButton = new QPushButton(tr("Hírek"), central);
    m_newsButton->setObjectName(QStringLiteral("newsButton"));
    m_crashButton = new QPushButton(tr("Crash jelentés"), central);
    m_crashButton->setObjectName(QStringLiteral("crashButton"));
    m_importModpackButton = new QPushButton(tr("Modpack import"), central);
    m_importModpackButton->setObjectName(QStringLiteral("importModpackButton"));
    actionLayout->addWidget(m_newsButton);
    actionLayout->addWidget(m_crashButton);
    actionLayout->addWidget(m_importModpackButton);
    layout->addLayout(actionLayout);

    m_skinPreview = new SkinPreviewWidget(central);
    m_skinPreview->setObjectName(QStringLiteral("skinPreview"));
    layout->addWidget(m_skinPreview);

    central->setLayout(layout);
    setCentralWidget(central);

    m_statusLabel = new QLabel(tr("Kész"), this);
    m_statusLabel->setObjectName(QStringLiteral("statusLabel"));
    m_statusLabel->setWordWrap(true);
    statusBar()->addPermanentWidget(m_statusLabel);

    connect(m_saveBackendButton, &QPushButton::clicked,
            this, &MainWindow::onSaveBackendUrl);
    connect(m_backendUrlEdit, &QLineEdit::returnPressed,
            this, &MainWindow::onSaveBackendUrl);
    connect(m_loginButton, &QPushButton::clicked,
            this, &MainWindow::onLoginClicked);
    connect(m_logoutButton, &QPushButton::clicked,
            this, &MainWindow::onLogoutClicked);
    connect(m_skinPreview, &SkinPreviewWidget::skinUploadFinished,
            this, &MainWindow::onSkinUploadFinished);
    connect(m_ramSlider, &QSlider::valueChanged,
            this, &MainWindow::onRamSliderChanged);
    connect(m_newsButton, &QPushButton::clicked,
            this, &MainWindow::onNewsClicked);
    connect(m_crashButton, &QPushButton::clicked,
            this, &MainWindow::onCrashReportClicked);
    connect(m_importModpackButton, &QPushButton::clicked,
            this, &MainWindow::onImportModpackClicked);
}

void MainWindow::loadSettings()
{
    SettingsStore::Settings loaded;
    if (SettingsStore::load(m_settingsPath, loaded))
        m_settings = loaded;
    else
        m_settings = SettingsStore::defaults();

    m_backendUrlEdit->setText(m_settings.backendUrl);
    if (m_ramSlider != nullptr)
        m_ramSlider->setValue(m_settings.maxMemoryMb);
    if (m_ramLabel != nullptr)
        m_ramLabel->setText(QString::number(m_settings.maxMemoryMb));
    if (m_settings.backendUrl.isEmpty())
        return;

    m_authClient->setBackendUrl(m_settings.backendUrl);
    if (m_authClient->backendUrl().isEmpty()) {
        showStatus(tr("A mentett backend URL érvénytelen. Add meg újra."));
        return;
    }
    // Show the normalized URL (no trailing slash, no path).
    m_backendUrlEdit->setText(m_authClient->backendUrl());
}

bool MainWindow::saveSettings()
{
    return SettingsStore::save(m_settingsPath, m_settings);
}

void MainWindow::updateAuthUi()
{
    const AuthSessionPtr session =
        m_sessions != nullptr ? m_sessions->currentSession() : AuthSessionPtr();
    const bool loggedIn = session != nullptr;

    m_authStatusLabel->setText(loggedIn
        ? tr("Bejelentkezve: %1").arg(session->username)
        : tr("Nincs bejelentkezve"));

    const bool hasBackend = m_authClient != nullptr && !m_authClient->backendUrl().isEmpty();
    m_backendUrlEdit->setEnabled(!loggedIn && !m_busy);
    m_saveBackendButton->setEnabled(!loggedIn && !m_busy);
    m_loginButton->setEnabled(!loggedIn && !m_busy && hasBackend);
    m_logoutButton->setEnabled(loggedIn && !m_busy);
    m_skinPreview->setEnabled(loggedIn);
}

void MainWindow::setBusy(bool busy)
{
    m_busy = busy;
    updateAuthUi();
}

void MainWindow::attemptSessionRestore()
{
    if (m_settings.username.isEmpty())
        return;
    if (m_authClient->backendUrl().isEmpty()) {
        showStatus(tr("A mentett munkamenethez nincs érvényes backend URL."));
        return;
    }

    m_sessions->restore(m_settings.username, [this](bool ok, const QString &error) {
        if (!ok)
            showStatus(tr("A munkamenet visszaállítása sikertelen: %1").arg(error));
        else
            showStatus(tr("Munkamenet visszaállítva: %1").arg(m_settings.username));
        updateAuthUi();
    });
}

void MainWindow::onSaveBackendUrl()
{
    if (m_sessions->currentSession() != nullptr) {
        showStatus(tr("Bejelentkezve a backend URL nem módosítható. Jelentkezz ki előbb."));
        return;
    }

    const QString raw = m_backendUrlEdit->text().trimmed();
    if (raw.isEmpty()) {
        showStatus(tr("Add meg a backend URL-t (pl. https://api.example.com)."));
        return;
    }

    m_authClient->setBackendUrl(raw);
    if (m_authClient->backendUrl().isEmpty()) {
        showStatus(tr("Érvénytelen backend URL. HTTPS kötelező; http:// csak localhost/127.0.0.1 esetén engedett."));
        return;
    }

    m_settings.backendUrl = m_authClient->backendUrl();
    m_backendUrlEdit->setText(m_settings.backendUrl);
    if (!saveSettings()) {
        showStatus(tr("A backend URL érvényes, de a beállítások mentése sikertelen."));
        return;
    }
    showStatus(tr("Backend URL mentve: %1").arg(m_settings.backendUrl));
    updateAuthUi();
}

void MainWindow::onLoginClicked()
{
    if (m_authClient->backendUrl().isEmpty()) {
        showStatus(tr("Előbb add meg és mentsd a backend URL-t."));
        return;
    }

    LoginDialog dialog(this);
    if (!m_settings.username.isEmpty())
        dialog.setUsername(m_settings.username);
    if (dialog.exec() != QDialog::Accepted)
        return;

    const QString username = dialog.username().trimmed();
    const QString password = dialog.password();

    setBusy(true);
    showStatus(tr("Bejelentkezés: %1...").arg(username));
    m_sessions->login(username, password, deviceId(), deviceName(),
                      [this, username](bool ok, const QString &error) {
                          setBusy(false);
                          if (!ok) {
                              showStatus(tr("Bejelentkezés sikertelen: %1").arg(error));
                              return;
                          }
                          m_settings.username = username;
                          if (!saveSettings())
                              showStatus(tr("Bejelentkezve, de a beállítások mentése sikertelen."));
                          else
                              showStatus(tr("Bejelentkezve: %1").arg(username));
                          updateAuthUi();
                      });
}

void MainWindow::onLogoutClicked()
{
    setBusy(true);
    m_sessions->logout([this](bool ok, const QString &error) {
        setBusy(false);
        m_settings.username.clear();
        if (!saveSettings())
            showStatus(tr("Kijelentkezve, de a beállítások mentése sikertelen."));
        else if (ok)
            showStatus(tr("Kijelentkezve."));
        else
            showStatus(tr("A szerveroldali kijelentkezés nem sikerült; a helyi munkamenet törölve (%1).").arg(error));
        updateAuthUi();
    });
}

void MainWindow::onSkinUploadFinished(bool success, const QString &skinId, const QString &error)
{
    if (success) {
        showStatus(tr("Skin sikeresen frissítve (ID: %1)").arg(skinId));
        statusBar()->showMessage(tr("Skin feltöltés sikeres"), 5000);
        return;
    }

    showStatus(error.isEmpty()
                   ? tr("Skin feltöltés sikertelen.")
                   : tr("Skin feltöltés sikertelen: %1").arg(error));
    statusBar()->showMessage(tr("Skin feltöltés sikertelen"), 5000);
}

void MainWindow::onRamSliderChanged(int value)
{
    m_settings.maxMemoryMb = value;
    if (m_ramLabel != nullptr)
        m_ramLabel->setText(QString::number(value));
    saveSettings();
}

void MainWindow::onNewsClicked()
{
    if (m_authClient == nullptr || m_authClient->backendUrl().isEmpty()) {
        showStatus(tr("Előbb add meg a backend URL-t."));
        return;
    }
    auto *nam = new QNetworkAccessManager(this);
    QNetworkRequest req(QUrl(m_authClient->backendUrl() + QStringLiteral("/api/v1/news")));
    QNetworkReply *reply = nam->get(req);
    connect(reply, &QNetworkReply::finished, this, [this, reply, nam]() {
        if (reply->error() == QNetworkReply::NoError) {
            showStatus(tr("Hírek betöltve."));
        } else {
            showStatus(tr("Hírek betöltése sikertelen: %1").arg(reply->errorString()));
        }
        reply->deleteLater();
        nam->deleteLater();
    });
}

void MainWindow::onCrashReportClicked()
{
    CrashReportDialog dlg(m_authClient ? m_authClient->backendUrl() : QString(), this);
    // Best-effort: read a local crash log if present.
    QString logPath = appDataDirectory() + QStringLiteral("/crash.log");
    QFile f(logPath);
    if (f.open(QIODevice::ReadOnly | QIODevice::Text)) {
        dlg.setLogText(QString::fromUtf8(f.readAll()));
        f.close();
    }
    connect(&dlg, &CrashReportDialog::sendRequested, this, [this](const QString &msg, const QString &logs) {
        if (m_authClient == nullptr || m_authClient->backendUrl().isEmpty())
            return;
        auto *nam = new QNetworkAccessManager(this);
        QNetworkRequest req(QUrl(m_authClient->backendUrl() + QStringLiteral("/api/v1/crash-reports")));
        req.setHeader(QNetworkRequest::ContentTypeHeader, QStringLiteral("application/json"));
        QJsonObject obj;
        obj[QStringLiteral("message")] = msg;
        obj[QStringLiteral("logs")] = logs;
        QJsonDocument doc(obj);
        QNetworkReply *reply = nam->post(req, doc.toJson());
        connect(reply, &QNetworkReply::finished, this, [reply, nam]() {
            reply->deleteLater();
            nam->deleteLater();
        });
    });
    dlg.exec();
}

void MainWindow::onImportModpackClicked()
{
    QString file = QFileDialog::getOpenFileName(this, tr("Modpack importálása"), QString(), tr("Zip fájlok (*.zip)"));
    if (file.isEmpty())
        return;
    showStatus(tr("Modpack kiválasztva: %1 (import folyamatban...)").arg(file));
    // Extraction is intentionally left to the instance pipeline; here we only
    // record the selection for the user.
}

void MainWindow::showStatus(const QString &message)
{
    if (m_statusLabel != nullptr)
        m_statusLabel->setText(message);
}

bool MainWindow::isLoggedIn() const
{
    return m_sessions != nullptr && m_sessions->currentSession() != nullptr;
}

QString MainWindow::backendUrl() const
{
    return m_authClient != nullptr ? m_authClient->backendUrl() : QString();
}

QString MainWindow::settingsPath() const
{
    return m_settingsPath;
}

QString MainWindow::deviceId()
{
    const QByteArray unique = QSysInfo::machineUniqueId();
    if (!unique.isEmpty())
        return QString::fromUtf8(unique);
    return deviceName();
}

QString MainWindow::deviceName()
{
    const QString host = QSysInfo::machineHostName();
    return host.isEmpty() ? QStringLiteral("unknown-device") : host;
}
