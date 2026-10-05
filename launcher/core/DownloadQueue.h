// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher core. Bounded worker pool for asset downloads.
#pragma once

#include <QObject>
#include <QUrl>
#include <QString>
#include <QStringList>
#include <QList>
#include <QHash>

class DownloadAsset;

// DownloadQueue downloads a list of (url, destination, sha256) items using a
// bounded number of concurrent DownloadAsset workers. Items are processed
// FIFO; a failed item is reported via itemFailed and does not stop the rest.
//
// Security: every download is SHA-256 verified by DownloadAsset before it is
// accepted, and destinations are canonicalized to absolute paths.
class DownloadQueue : public QObject
{
    Q_OBJECT
public:
    struct Item
    {
        QUrl url;
        QString dest;
        QString sha256;
    };

    // maxConcurrent must be >= 1 (clamped internally).
    explicit DownloadQueue(int maxConcurrent = 4, QObject *parent = nullptr);

    // Append items and start processing if not already running.
    void enqueue(const QList<Item> &items);

    bool isRunning() const;
    int queuedCount() const;

signals:
    void itemFinished(const QString &dest);
    void itemFailed(const QString &dest, const QString &reason);
    void progress(qreal overallPercent); // across the whole queue
    void finished(); // emitted when the queue drains

private:
    void pump();
    void workerDone(QObject *worker);

    int m_maxConcurrent;
    QList<Item> m_pending;
    QList<DownloadAsset *> m_idle;
    QList<DownloadAsset *> m_active;
    QHash<DownloadAsset *, QString> m_activeDest;
    int m_total = 0;
    int m_done = 0;
};
