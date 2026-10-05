// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher core. Asynchronous asset downloader with SHA-256
// integrity verification.
#pragma once

#include <QObject>
#include <QString>
#include <QUrl>
#include <QByteArray>
#include <QStringList>

class QNetworkAccessManager;
class QNetworkReply;

// DownloadAsset downloads a remote file to a local path and verifies its
// SHA-256 digest before it is accepted. It is used for modpack, asset and
// library downloads, where a tampered or corrupted file must never be used.
//
// Security: the digest is checked before the file is considered valid; a
// mismatched digest never produces a "success". The target path is resolved
// to a canonical absolute path so callers cannot be tricked into writing
// outside an intended directory.
class DownloadAsset : public QObject
{
    Q_OBJECT
public:
    explicit DownloadAsset(QObject *parent = nullptr);

    // Download url to localPath (absolute) and verify it matches
    // expectedSha256 (lowercase hex). Only one download runs at a time.
    void start(const QUrl &url, const QString &localPath, const QString &expectedSha256);

    // True if a download is currently in progress.
    bool isRunning() const;

signals:
    // Emitted with the resolved absolute destination path on success.
    void finished(const QString &localPath);
    void failed(const QString &reason);

    // 0..100.
    void progress(qreal percent);

private:
    void handleReply();
    void finish(const QString &error);

    QNetworkAccessManager *m_nam = nullptr;
    QNetworkReply *m_reply = nullptr;
    QString m_localPath;
    QString m_expectedSha256;
};
