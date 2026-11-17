// SPDX-License-Identifier: GPL-3.0-only
#pragma once

#include <QMainWindow>
#include <QString>

#include <memory>

#include <memory>

#include "core/SettingsStore.h"

class BackendAuthClient;
class SessionManager;
class SkinPreviewWidget;
class TokenStorage;

class QCheckBox;
class QLabel;
class QLineEdit;
class QPushButton;
class QSlider;

// MainWindow owns the launcher's backend session:
//  - the backend URL is edited, validated and persisted here,
//  - login/logout go through SessionManager (tokens are stored with DPAPI on
//    Windows via TokenStorage),
//  - the restored/authenticated BackendAuthClient is handed to the embedded
//    SkinPreviewWidget, so skin uploads work from the GUI.
class MainWindow : public QMainWindow
{
    Q_OBJECT

public:
    explicit MainWindow(QWidget *parent = nullptr);
    ~MainWindow() override;

    [[nodiscard]] bool isLoggedIn() const;
    [[nodiscard]] QString backendUrl() const;

    // Location of launcher.json (also used by tests to isolate state).
    [[nodiscard]] QString settingsPath() const;
    [[nodiscard]] static QString defaultSettingsPath();

private slots:
    void onSaveBackendUrl();
    void onLoginClicked();
    void onLogoutClicked();
    void onSkinUploadFinished(bool success, const QString &skinId, const QString &error);
    void onNewsClicked();
    void onCrashReportClicked();
    void onImportModpackClicked();
    void onRamSliderChanged(int value);

private:
    void buildUi();
    void loadSettings();
    [[nodiscard]] bool saveSettings();
    void updateAuthUi();
    void setBusy(bool busy);
    void attemptSessionRestore();
    void showStatus(const QString &message);
    [[nodiscard]] static QString deviceId();
    [[nodiscard]] static QString deviceName();

    QString m_settingsPath;
    SettingsStore::Settings m_settings;
    std::unique_ptr<TokenStorage> m_tokenStorage;
    BackendAuthClient *m_authClient = nullptr;
    SessionManager *m_sessions = nullptr;

    SkinPreviewWidget *m_skinPreview = nullptr;
    QSlider *m_ramSlider = nullptr;
    QLabel *m_ramLabel = nullptr;
    QCheckBox *m_offlineModeCheckBox = nullptr;
    QPushButton *m_newsButton = nullptr;
    QPushButton *m_crashButton = nullptr;
    QPushButton *m_importModpackButton = nullptr;
    QLineEdit *m_backendUrlEdit = nullptr;
    QPushButton *m_saveBackendButton = nullptr;
    QPushButton *m_loginButton = nullptr;
    QPushButton *m_logoutButton = nullptr;
    QLabel *m_authStatusLabel = nullptr;
    QLabel *m_statusLabel = nullptr;
    bool m_busy = false;
};
