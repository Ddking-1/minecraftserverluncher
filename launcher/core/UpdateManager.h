// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher core. Fetches the update manifest from the backend
// and reports whether an update is indicated. The actual download is handled
// by DownloadAsset (SHA-256 verified). Covers Batch 11 auto-update engine.
#pragma once

#include <QNetworkAccessManager>
#include <QObject>
#include <QString>

class UpdateManager : public QObject
{
    Q_OBJECT
public:
    explicit UpdateManager(QObject *parent = nullptr);

    // backendBase like "https://api.example.com", current like "0.1.0".
    void setEndpoints(const QString &backendBase, const QString &currentVersion);

    QString currentVersion() const;

    // Fetch the manifest from /api/v1/updates/manifest and emit updateInfo.
    void checkForUpdate();

signals:
    // version, url, sha256, mandatory
    void updateInfo(const QString &version, const QString &url, const QString &sha256, bool mandatory);
    void checkFailed(const QString &reason);

private:
    QString m_backend;
    QString m_current;
    QNetworkAccessManager *m_nam = nullptr;
};