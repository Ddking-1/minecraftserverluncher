// SPDX-License-Identifier: GPL-3.0-only
#include "MainWindow.h"

#include "auth/BackendAuthClient.h"
#include "auth/SessionManager.h"
#include "auth/TokenStorage.h"
#include "core/FriendlyError.h"
#include "core/SystemMemory.h"
#include "gui/CrashReportDialog.h"
#include "gui/LoginDialog.h"
#include "ui/SkinPreviewWidget.h"

#include <QCheckBox>
#include <QFileDialog>
#include <QFile>
#include <QFileInfo>
#include <QJsonArray>
#include <QJsonDocument>
#include <QJsonObject>
#include <QJsonParseError>
#include <QNetworkAccessManager>
#include <QNetworkReply>
#include <QMessageBox>
#include <QNetworkRequest>
#include <QProcess>
#include <QSlider>
#include <QSysInfo>

#include <QDir>
#include <QDirIterator>
#include <QGroupBox>
#include <QHBoxLayout>
#include <QLabel>
#include <QLineEdit>
#include <QLocale>
#include <QPushButton>
#include <QSignalBlocker>
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
    m_allowedMaxMemoryMb = SystemMemory::allowedMaxMb();
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
    auto *ramOuter = new QVBoxLayout(ramBox);
    auto *ramLayout = new QHBoxLayout();
    m_ramLabel = new QLabel(tr("4096"), ramBox);
    m_ramSlider = new QSlider(Qt::Horizontal, ramBox);
    m_ramSlider->setObjectName(QStringLiteral("ramSlider"));
    // The upper bound comes from the machine, not from a fixed 64 GB guess.
    m_ramSlider->setRange(SystemMemory::kMinMb, m_allowedMaxMemoryMb);
    m_ramSlider->setValue(m_settings.maxMemoryMb);
    m_ramSlider->setSingleStep(256);
    m_ramSlider->setPageStep(512);
    m_ramSlider->setToolTip(tr("A játéknak adott memória. A felső korlát a gép memóriája "
                              "mínusz %1, amit a rendszernek hagyunk.")
                               .arg(formatMb(SystemMemory::kReservedMb)));
    ramLayout->addWidget(m_ramSlider, 1);
    ramLayout->addWidget(m_ramLabel);
    ramOuter->addLayout(ramLayout);

    // Short, plain-language statistics: what the machine has, what we allow.
    m_ramInfoLabel = new QLabel(ramBox);
    m_ramInfoLabel->setObjectName(QStringLiteral("ramInfoLabel"));
    m_ramInfoLabel->setWordWrap(true);
    m_ramInfoLabel->setTextInteractionFlags(Qt::TextSelectableByMouse);
    ramOuter->addWidget(m_ramInfoLabel);

    layout->addWidget(ramBox);
    updateRamInfo();

    // Offline mode (F6)
    m_offlineModeCheckBox = new QCheckBox(tr("Offline mód"), central);
    m_offlineModeCheckBox->setObjectName(QStringLiteral("offlineModeCheckBox"));
    layout->addWidget(m_offlineModeCheckBox);
    connect(m_offlineModeCheckBox, &QCheckBox::toggled, this, [this](bool on) {
        if (m_busy)
            return; // never flip transport/auth state mid-request
        m_settings.offlineMode = on;
        (void)saveSettings();
        updateAuthUi();
        if (on) {
            const bool hadSession = m_sessions != nullptr && m_sessions->currentSession() != nullptr;
            showStatus(hadSession
                           ? tr("Offline mód bekapcsolva: a meglévő munkamenet inaktív, hálózati funkciók letiltva.")
                           : tr("Offline mód bekapcsolva: hálózati funkciók letiltva."));
        } else {
            showStatus(tr("Offline mód kikapcsolva."));
        }
    });

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

    // A hand-edited launcher.json (or a settings file copied from a bigger
    // machine) must not ask for more memory than this computer has.
    const int requestedMemoryMb = m_settings.maxMemoryMb;
    m_settings.maxMemoryMb = SystemMemory::clampToAllowed(requestedMemoryMb, m_allowedMaxMemoryMb);
    if (m_settings.maxMemoryMb != requestedMemoryMb) {
        (void)saveSettings();
        const int totalMb = SystemMemory::totalPhysicalMb();
        if (totalMb > 0) {
            showStatus(tr("A mentett %1 memória-beállítás túl nagy ehhez a géphez: összesen %2 memória "
                          "van, ebből %3 a rendszernek van fenntartva, így legfeljebb %4 engedélyezett. "
                          "A beállítást a biztonságos maximumra állítottuk.")
                           .arg(formatMb(requestedMemoryMb), formatMb(totalMb),
                                formatMb(SystemMemory::kReservedMb),
                                formatMb(m_settings.maxMemoryMb)));
        } else {
            showStatus(tr("A mentett %1 memória-beállítás érvénytelen volt, ezért %2-re módosítottuk.")
                           .arg(formatMb(requestedMemoryMb), formatMb(m_settings.maxMemoryMb)));
        }
    }

    m_backendUrlEdit->setText(m_settings.backendUrl);
    if (m_ramSlider != nullptr)
        m_ramSlider->setValue(m_settings.maxMemoryMb);
    if (m_ramLabel != nullptr)
        m_ramLabel->setText(QString::number(m_settings.maxMemoryMb));
    updateRamInfo();
    if (m_offlineModeCheckBox != nullptr)
        m_offlineModeCheckBox->setChecked(m_settings.offlineMode);
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
    const bool offline = m_settings.offlineMode;
    m_backendUrlEdit->setEnabled(!loggedIn && !m_busy && !offline);
    m_saveBackendButton->setEnabled(!loggedIn && !m_busy && !offline);
    m_loginButton->setEnabled(!loggedIn && !m_busy && hasBackend && !offline);
    m_logoutButton->setEnabled(loggedIn && !m_busy);
    // Offline mode must actually disable network-backed features, not just the
    // login button: skin upload requires a live backend session.
    m_skinPreview->setEnabled(loggedIn && !offline);
    if (m_offlineModeCheckBox != nullptr)
        m_offlineModeCheckBox->setEnabled(!m_busy);
}

void MainWindow::setBusy(bool busy)
{
    m_busy = busy;
    updateAuthUi();
}

void MainWindow::attemptSessionRestore()
{
    if (m_settings.offlineMode) {
        showStatus(tr("Offline mód: munkamenet-visszaállítás kihagyva."));
        return;
    }
    if (m_settings.username.isEmpty())
        return;
    if (m_authClient->backendUrl().isEmpty()) {
        showStatus(tr("A mentett munkamenethez nincs érvényes backend URL."));
        return;
    }

    m_sessions->restore(m_settings.username, [this](bool ok, const QString &error) {
        if (!ok)
            showStatus(tr("A munkamenet visszaállítása sikertelen: %1")
                           .arg(FriendlyError::explain(error)));
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
    if (m_settings.offlineMode) {
        showStatus(tr("Offline módban nem lehet bejelentkezni."));
        return;
    }
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
                              showStatus(tr("Bejelentkezés sikertelen: %1")
                                             .arg(FriendlyError::explain(error)));
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
            showStatus(tr("A szerveroldali kijelentkezés nem sikerült; a helyi munkamenet "
                          "törölve (%1).")
                           .arg(FriendlyError::explain(error)));
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
                   : tr("Skin feltöltés sikertelen: %1").arg(FriendlyError::explain(error)));
    statusBar()->showMessage(tr("Skin feltöltés sikertelen"), 5000);
}

QString MainWindow::formatMb(int mb)
{
    if (mb <= 0)
        return QStringLiteral("?");
    if (mb % 1024 == 0)
        return tr("%1 GB").arg(mb / 1024);
    return tr("%1 GB").arg(QLocale().toString(mb / 1024.0, 'f', 1));
}

void MainWindow::updateRamInfo()
{
    if (m_ramInfoLabel == nullptr)
        return;

    const int totalMb = SystemMemory::totalPhysicalMb();
    const int selectedMb = m_settings.maxMemoryMb;
    const int allowedMb = m_allowedMaxMemoryMb;

    if (totalMb > 0) {
        m_ramInfoLabel->setText(
            tr("Gép memóriája: %1 · ebből legfeljebb %2 engedélyezett (%3 a rendszernek fenntartva) "
               "· beállítva: %4")
                .arg(formatMb(totalMb), formatMb(allowedMb),
                     formatMb(SystemMemory::kReservedMb), formatMb(selectedMb)));
    } else {
        // Unknown machine: do not pretend to know the size, keep it honest.
        m_ramInfoLabel->setText(
            tr("A gép memóriája nem állapítható meg, ezért a biztonságos maximum %1 · beállítva: %2")
                .arg(formatMb(allowedMb), formatMb(selectedMb)));
    }
}

void MainWindow::onRamSliderChanged(int value)
{
    const int clamped = SystemMemory::clampToAllowed(value, m_allowedMaxMemoryMb);
    if (clamped != value) {
        QSignalBlocker blocker(m_ramSlider);
        m_ramSlider->setValue(clamped);
        const int totalMb = SystemMemory::totalPhysicalMb();
        if (totalMb > 0) {
            showStatus(tr("Legfeljebb %1 memória engedélyezett ezen a gépen: %2 rendszer memória, "
                          "ebből %3 a rendszernek van fenntartva. A beállítást a maximumra állítottuk.")
                           .arg(formatMb(clamped), formatMb(totalMb),
                                formatMb(SystemMemory::kReservedMb)));
        } else {
            showStatus(tr("Legfeljebb %1 memória engedélyezett. A beállítást a maximumra állítottuk.")
                           .arg(formatMb(clamped)));
        }
    }

    m_settings.maxMemoryMb = clamped;
    if (m_ramLabel != nullptr)
        m_ramLabel->setText(QString::number(clamped));
    updateRamInfo();
    (void)saveSettings();
}

void MainWindow::onNewsClicked()
{
    if (m_settings.offlineMode) {
        showStatus(tr("Offline módban a hírek nem tölthetők be."));
        return;
    }
    if (m_authClient == nullptr || m_authClient->backendUrl().isEmpty()) {
        showStatus(tr("Előbb add meg a backend URL-t."));
        return;
    }
    // Reentrancy guard.
    if (m_newsButton != nullptr && !m_newsButton->isEnabled())
        return;
    if (m_newsButton != nullptr)
        m_newsButton->setEnabled(false);

    auto *nam = new QNetworkAccessManager(this);
    QNetworkRequest req(QUrl(m_authClient->backendUrl() + QStringLiteral("/api/v1/news")));
    req.setTransferTimeout(15000);
    QNetworkReply *reply = nam->get(req);
    connect(reply, &QNetworkReply::finished, this, [this, reply, nam]() {
        if (reply->error() == QNetworkReply::NoError) {
            // Bound the response so a hostile/huge payload cannot hang the UI.
            constexpr qint64 kMaxNewsBytes = 512 * 1024;
            constexpr int kMaxItems = 20;
            constexpr int kMaxFieldChars = 500;
            const QByteArray payload = reply->read(kMaxNewsBytes);
            QJsonParseError parseError{};
            const QJsonDocument doc = QJsonDocument::fromJson(payload, &parseError);
            if (parseError.error != QJsonParseError::NoError || !doc.isObject()) {
                showStatus(tr("Hírek betöltése sikertelen: érvénytelen válasz."));
                if (m_newsButton != nullptr)
                    m_newsButton->setEnabled(true);
                reply->deleteLater();
                nam->deleteLater();
                return;
            }
            QString text;
            const QJsonArray items = doc.object().value(QStringLiteral("items")).toArray();
            for (const QJsonValue &v : items) {
                if (text.size() > kMaxNewsBytes || text.count(QChar('\n')) > kMaxItems * 3)
                    break;
                const QJsonObject o = v.toObject();
                QString title = o.value(QStringLiteral("title")).toString();
                QString body = o.value(QStringLiteral("body")).toString();
                title.truncate(kMaxFieldChars);
                body.truncate(kMaxFieldChars);
                text += QStringLiteral("• %1\n%2\n\n").arg(title, body);
            }
            if (text.isEmpty())
                text = tr("Nincs új hír.");
            // Render as plain text: a hostile title must not be interpreted as HTML.
            QMessageBox box(QMessageBox::Information, tr("Hírek"), text,
                            QMessageBox::Ok, this);
            box.setTextFormat(Qt::PlainText);
            (void)box.exec();
            showStatus(tr("Hírek betöltve."));
        } else {
            showStatus(tr("Hírek betöltése sikertelen: %1")
                           .arg(FriendlyError::explain(reply->errorString())));
        }
        if (m_newsButton != nullptr)
            m_newsButton->setEnabled(true);
        reply->deleteLater();
        nam->deleteLater();
    });
}

void MainWindow::onCrashReportClicked()
{
    // Reentrancy guard.
    if (m_crashButton != nullptr && !m_crashButton->isEnabled())
        return;
    if (m_crashButton != nullptr)
        m_crashButton->setEnabled(false);

    CrashReportDialog dlg(m_authClient ? m_authClient->backendUrl() : QString(), this);
    // Best-effort: read a local crash log if present (bounded read).
    QString logPath = appDataDirectory() + QStringLiteral("/crash.log");
    QFile f(logPath);
    constexpr qint64 kMaxCrashLogBytes = 8 * 1024 * 1024;
    if (f.open(QIODevice::ReadOnly | QIODevice::Text)) {
        dlg.setLogText(QString::fromUtf8(f.read(kMaxCrashLogBytes)));
        f.close();
    }
    connect(&dlg, &CrashReportDialog::sendRequested, this, [this](const QString &msg, const QString &logs) {
        if (m_settings.offlineMode) {
            showStatus(tr("Offline módban a crash jelentés nem küldhető el."));
            return;
        }
        if (m_authClient == nullptr || m_authClient->backendUrl().isEmpty()) {
            showStatus(tr("Crash jelentés nem küldhető el: nincs backend URL."));
            return;
        }
        if (m_sessions == nullptr || m_sessions->currentSession() == nullptr) {
            // The endpoint is authenticated by design; fail fast with a clear
            // message instead of a bare 401 from the backend.
            showStatus(tr("Crash jelentés küldéséhez jelentkezz be."));
            return;
        }
        auto *nam = new QNetworkAccessManager(this);
        QNetworkRequest req(QUrl(m_authClient->backendUrl() + QStringLiteral("/api/v1/crash-reports")));
        req.setHeader(QNetworkRequest::ContentTypeHeader, QStringLiteral("application/json"));
        req.setTransferTimeout(15000);
        req.setRawHeader("Authorization",
                         ("Bearer " + m_sessions->currentSession()->access_token).toUtf8());
        QJsonObject obj;
        // Truncate so a hostile/pasted blob cannot exceed the backend's 1 MiB
        // request limit (which would turn into an opaque 400).
        obj[QStringLiteral("message")] = msg.left(4000);
        obj[QStringLiteral("logs")] = logs.left(256 * 1024);
        obj[QStringLiteral("machine")] = QSysInfo::machineHostName();
        obj[QStringLiteral("os")] = QSysInfo::prettyProductName();
        QJsonDocument doc(obj);
        QNetworkReply *reply = nam->post(req, doc.toJson());
        connect(reply, &QNetworkReply::finished, this, [this, reply, nam]() {
            if (reply->error() == QNetworkReply::NoError)
                showStatus(tr("Crash jelentés elküldve."));
            else
                showStatus(tr("Crash jelentés küldése sikertelen: %1")
                               .arg(FriendlyError::explain(reply->errorString())));
            if (m_crashButton != nullptr)
                m_crashButton->setEnabled(true);
            reply->deleteLater();
            nam->deleteLater();
        });
    });
    (void)dlg.exec();
    // Re-enable after dialog closes (covers cancel/no-send path).
    if (m_crashButton != nullptr)
        m_crashButton->setEnabled(true);
}

void MainWindow::onImportModpackClicked()
{
    // Reentrancy guard: prevent double-click launching two extractions.
    if (m_importModpackButton != nullptr && !m_importModpackButton->isEnabled())
        return;
    if (m_importModpackButton != nullptr)
        m_importModpackButton->setEnabled(false);

    // Single exit path: report the outcome and always re-enable the button so
    // an error cannot permanently lock the feature out.
    const auto finish = [this](const QString &message) {
        showStatus(message);
        if (m_importModpackButton != nullptr)
            m_importModpackButton->setEnabled(true);
    };

    QString file = QFileDialog::getOpenFileName(this, tr("Modpack importálása"), QString(), tr("Zip fájlok (*.zip)"));
    if (file.isEmpty()) {
        if (m_importModpackButton != nullptr)
            m_importModpackButton->setEnabled(true);
        return;
    }

    // Validate input file (empty data / missing file).
    QFileInfo fi(file);
    if (!fi.exists() || !fi.isFile() || fi.size() <= 0) {
        finish(tr("Modpack import sikertelen: üres vagy nem létező fájl."));
        return;
    }

    // Guard against path injection: reject paths containing quotes/newlines.
    if (file.contains(QChar('\'')) || file.contains(QChar('"')) || file.contains(QChar('\n')) ||
        file.contains(QChar('\r'))) {
        finish(tr("Modpack import sikertelen: érvénytelen fájlnév."));
        return;
    }

    // Choose a destination directory under the app data location.
    QString dest = appDataDirectory() + QStringLiteral("/modpacks/imported");
    QDir().mkpath(dest);
    if (dest.contains(QChar('\'')) || dest.contains(QChar('"')) || dest.contains(QChar('\n')) ||
        dest.contains(QChar('\r'))) {
        finish(tr("Modpack import sikertelen: érvénytelen célkönyvtár."));
        return;
    }

    // Pre-check the archive before extracting: path traversal (zip slip),
    // absurd uncompressed size (disk-fill DoS), entry-count bomb, and
    // unreadable/corrupt zips get distinct diagnostics.
    {
        const QString script = QStringLiteral(
                                   "$ErrorActionPreference='Stop';"
                                   "try {"
                                   "Add-Type -AssemblyName System.IO.Compression.FileSystem;"
                                   "$z=[System.IO.Compression.ZipFile]::OpenRead('%1');"
                                   "$e=$z.Entries;"
                                   "$bad=@($e | Where-Object {"
                                   "$_.FullName -match '\\.\\.' -or $_.FullName -match '^[/\\\\]' });"
                                   "$size=($e | Measure-Object -Property Length -Sum).Sum;"
                                   "$count=$e.Count;"
                                   "$z.Dispose();"
                                   "if ($bad.Count -gt 0) { exit 1 };"
                                   "if ($size -gt 4294967296) { exit 2 };"
                                   "if ($count -gt 200000) { exit 3 };"
                                   "exit 0"
                                   "} catch { exit 4 }")
                                   .arg(file);
        QProcess chk;
        chk.start(QStringLiteral("powershell"),
                  {QStringLiteral("-NoProfile"), QStringLiteral("-Command"), script});
        if (!chk.waitForFinished(30000)) {
            chk.kill();
            chk.waitForFinished(5000);
            finish(tr("Modpack ellenőrzés időtúllépés."));
            return;
        }
        switch (chk.exitCode()) {
        case 0:
            break;
        case 1:
            finish(tr("Modpack import sikertelen: gyanús (path traversal) bejegyzés a zipben."));
            return;
        case 2:
            finish(tr("Modpack import sikertelen: a kicsomagolt méret meghaladja a 4 GB-os limitet."));
            return;
        case 3:
            finish(tr("Modpack import sikertelen: túl sok fájl a zipben."));
            return;
        default:
            finish(tr("Modpack import sikertelen: a zip nem olvasható vagy sérült."));
            return;
        }
    }

    // Extract using PowerShell Expand-Archive (Windows). Run synchronously with
    // an argument list (no shell string interpolation) to avoid injection.
    const auto countFiles = [](const QString &dirPath) {
        int count = 0;
        QDirIterator it(dirPath, QDir::Files | QDir::Hidden, QDirIterator::Subdirectories);
        while (it.hasNext()) {
            it.next();
            ++count;
        }
        return count;
    };
    const int filesBefore = countFiles(dest);

    QProcess proc;
    QStringList args;
    args << QStringLiteral("-NoProfile")
         << QStringLiteral("-Command")
         << QStringLiteral("Expand-Archive -LiteralPath '%1' -DestinationPath '%2' -Force")
                .arg(file, dest);
    proc.start(QStringLiteral("powershell"), args);
    if (!proc.waitForFinished(120000)) {
        proc.kill();
        proc.waitForFinished(5000);
        finish(tr("Modpack import időtúllépés."));
        return;
    }
    if (proc.exitCode() != 0) {
        finish(tr("Modpack import sikertelen: %1").arg(QString::fromLocal8Bit(proc.readAllStandardError())));
        return;
    }

    // An archive with no entries "succeeds" silently; report it instead of
    // claiming a completed import.
    if (countFiles(dest) <= filesBefore) {
        finish(tr("Modpack import befejezve, de nem lett fájl kicsomagolva: %1").arg(dest));
        return;
    }
    finish(tr("Modpack importálva: %1").arg(dest));
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
