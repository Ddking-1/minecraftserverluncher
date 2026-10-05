// SPDX-License-Identifier: GPL-3.0-only
#include "DownloadQueue.h"

#include "DownloadAsset.h"

#include <algorithm>

DownloadQueue::DownloadQueue(int maxConcurrent, QObject *parent)
    : QObject(parent)
{
    m_maxConcurrent = std::max(1, maxConcurrent);
    for (int i = 0; i < m_maxConcurrent; ++i) {
        auto *w = new DownloadAsset(this);
        m_idle.append(w);
    }
}

bool DownloadQueue::isRunning() const
{
    return !m_active.isEmpty() || !m_pending.isEmpty();
}

int DownloadQueue::queuedCount() const
{
    return m_pending.size() + m_active.size();
}

void DownloadQueue::enqueue(const QList<Item> &items)
{
    m_total += items.size();
    for (const Item &it : items)
        m_pending.append(it);
    pump();
}

void DownloadQueue::pump()
{
    // Start workers while there is both work and an idle worker.
    while (!m_pending.isEmpty() && !m_idle.isEmpty()) {
        DownloadAsset *w = m_idle.takeFirst();
        const Item it = m_pending.takeFirst();

        connect(w, &DownloadAsset::finished, this, [this, w](const QString &dest) {
            emit itemFinished(dest);
            workerDone(w);
        });
        connect(w, &DownloadAsset::failed, this, [this, w](const QString &reason) {
            // Identify the destination for the failed item: DownloadAsset does
            // not expose it, so track it via the active map.
            emit itemFailed(m_activeDest.value(w), reason);
            workerDone(w);
        });
        m_active.append(w);
        m_activeDest.insert(w, it.dest);
        w->start(it.url, it.dest, it.sha256);
    }

    if (m_active.isEmpty() && m_pending.isEmpty()) {
        emit finished();
    }
}

void DownloadQueue::workerDone(QObject *worker)
{
    DownloadAsset *w = static_cast<DownloadAsset *>(worker);
    // Disconnect the per-item lambdas so they do not fire again.
    disconnect(w, nullptr, this, nullptr);
    m_active.removeOne(w);
    m_activeDest.remove(w);
    m_idle.append(w);
    ++m_done;

    // Convert total progress (0..100).
    if (m_total > 0)
        emit progress(qreal(m_done) / qreal(m_total) * 100.0);

    pump();
}
