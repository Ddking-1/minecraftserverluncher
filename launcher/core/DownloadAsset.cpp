// SPDX-License-Identifier: GPL-3.0-only
#include "DownloadAsset.h"

#include "FileIntegrity.h"

#include <QFile>
#include <QFileInfo>
#include <QNetworkAccessManager>
#include <QNetworkReply>
#include <QNetworkRequest>
#include <QSaveFile>
#include <QTimer>

DownloadAsset::DownloadAsset(QObject *parent)
    : QObject(parent), m_nam(new QNetworkAccessManager(this))
{
    m_nam->setTransferTimeout(30000);
}
bool DownloadAsset::isRunning() const
{
    return m_reply != nullptr;
}

void DownloadAsset::setMaxRetries(int retries)
{
    m_maxRetries = qMax(0, retries);
}

void DownloadAsset::start(const QUrl &url, const QString &localPath, const QString &expectedSha256)
{
    if (m_reply) {
        emit failed(QStringLiteral("a download is already in progress"));
        return;
    }
    if (!url.isValid() || url.isRelative()) {
        emit failed(QStringLiteral("invalid download URL"));
        return;
    }
    if (url.scheme() != QLatin1String("http") && url.scheme() != QLatin1String("https")) {
        emit failed(QStringLiteral("unsupported URL scheme"));
        return;
    }

    // Canonicalize the destination to an absolute path so a caller cannot
    // write outside its intended directory via relative/.. paths.
    const QString canonical = QFileInfo(localPath).absoluteFilePath();

    m_url = url;
    m_localPath = canonical;
    m_expectedSha256 = expectedSha256.trimmed();
    m_attempts = 0;
    issueRequest();
}

void DownloadAsset::issueRequest()
{
    QNetworkRequest req(m_url);
    req.setAttribute(QNetworkRequest::RedirectPolicyAttribute, QNetworkRequest::NoLessSafeRedirectPolicy);
    m_reply = m_nam->get(req);
    connect(m_reply, &QNetworkReply::downloadProgress, this, [this](qint64 recv, qint64 total) {
        if (total > 0)
            emit progress(qreal(recv) / qreal(total) * 100.0);
    });
    connect(m_reply, &QNetworkReply::finished, this, &DownloadAsset::handleReply);
}

void DownloadAsset::handleReply()
{
    QNetworkReply *reply = m_reply;
    m_reply = nullptr;
    ++m_attempts;

    if (reply->error() != QNetworkReply::NoError) {
        const QString err = reply->errorString();
        reply->deleteLater();
        // Transient network error: retry if we have attempts left.
        if (m_attempts <= m_maxRetries) {
            scheduleRetry();
            return;
        }
        finish(err);
        return;
    }
    const int status = reply->attribute(QNetworkRequest::HttpStatusCodeAttribute).toInt();
    if (status < 200 || status >= 300) {
        reply->deleteLater();
        if (status >= 500 && m_attempts <= m_maxRetries) {
            scheduleRetry();
            return;
        }
        finish(QStringLiteral("HTTP %1").arg(status));
        return;
    }
    const QByteArray data = reply->readAll();
    reply->deleteLater();

    // Atomically write then verify the digest before accepting the file.
    QSaveFile file(m_localPath);
    if (!file.open(QIODevice::WriteOnly)) {
        finish(QStringLiteral("could not open destination file"));
        return;
    }
    if (file.write(data) != data.size()) {
        finish(QStringLiteral("write failed"));
        return;
    }
    if (!file.commit()) {
        finish(QStringLiteral("commit failed"));
        return;
    }

    if (!FileIntegrity::verify(m_localPath, m_expectedSha256)) {
        QFile::remove(m_localPath);
        // Never retry a checksum mismatch: the file is corrupt, not transient.
        finish(QStringLiteral("SHA-256 mismatch"));
        return;
    }
    emit progress(100.0);
    emit finished(m_localPath);
}

void DownloadAsset::scheduleRetry()
{
    // Retry after a short delay without blocking the event loop.
    QTimer::singleShot(200, this, [this]() {
        if (m_reply == nullptr)
            issueRequest();
    });
}

void DownloadAsset::finish(const QString &error)
{
    if (error.isEmpty())
        emit finished(m_localPath);
    else
        emit failed(error);
}
