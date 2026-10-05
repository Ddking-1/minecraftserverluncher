// SPDX-License-Identifier: GPL-3.0-only
#include <QCoreApplication>
#include <QCryptographicHash>
#include <QEventLoop>
#include <QFile>
#include <QHash>
#include <QHostAddress>
#include <QList>
#include <QTcpServer>
#include <QTcpSocket>
#include <QTemporaryDir>
#include <QTimer>

#include "core/DownloadQueue.h"
#include "core/FileIntegrity.h"

#include <cstdio>

static int failures = 0;
static void check(bool cond, const char *msg)
{
    if (!cond) {
        std::fprintf(stderr, "FAIL: %s\n", msg);
        ++failures;
    }
}

int main(int argc, char **argv)
{
    QCoreApplication app(argc, argv);
    QTemporaryDir dir;
    check(dir.isValid(), "temp dir");

    // Local HTTP server serving three distinct files.
    const QList<QString> names = {QStringLiteral("a.bin"), QStringLiteral("b.bin"), QStringLiteral("c.bin")};
    QHash<QString, QString> digestOf;
    for (const QString &n : names) {
        // Compute the digest from the in-memory body so we never depend on a
        // test file write succeeding.
        digestOf[n] = QString::fromLatin1(QCryptographicHash::hash(n.toUtf8(), QCryptographicHash::Sha256).toHex());
    }
    QTcpServer server;
    check(server.listen(QHostAddress::LocalHost, 0), "listen");
    QObject::connect(&server, &QTcpServer::newConnection, &server, [&server, &names]() {
        QTcpSocket *sock = server.nextPendingConnection();
        QObject::connect(sock, &QTcpSocket::readyRead, sock, [sock, &names]() {
            const QByteArray req = sock->readAll();
            const QByteArray path = req.split(' ').value(1);
            // Path is "/a.bin" etc. Extract the filename without leading slash.
            QString name = QString::fromUtf8(path).remove(0, 1);
            QString body;
            for (const QString &n : names) {
                if (name == n) {
                    body = n.toUtf8();
                    break;
                }
            }
            if (body.isEmpty())
                body = "NOTFOUND";
            sock->write(QStringLiteral("HTTP/1.1 200 OK\r\nContent-Length: %1\r\nConnection: close\r\n\r\n").arg(body.size()).toUtf8());
            sock->write(body.toUtf8());
            sock->disconnectFromHost();
        });
    });

    const QString base = QStringLiteral("http://127.0.0.1:%1/").arg(server.serverPort());
    QList<DownloadQueue::Item> items;
    int completed = 0;
    int failed = 0;
    for (const QString &n : names) {
        DownloadQueue::Item it;
        it.url = QUrl(base + n);
        it.dest = dir.path() + QLatin1Char('/') + QStringLiteral("dl-") + n;
        it.sha256 = digestOf[n];
        items.append(it);
    }

    DownloadQueue queue(2);
    QEventLoop loop;
    QObject::connect(&queue, &DownloadQueue::itemFinished, &loop, [&](const QString &) { ++completed; });
    QObject::connect(&queue, &DownloadQueue::itemFailed, &loop, [&](const QString &, const QString &) { ++failed; });
    QObject::connect(&queue, &DownloadQueue::finished, &loop, &QEventLoop::quit);
    queue.enqueue(items);
    QTimer::singleShot(15000, &loop, &QEventLoop::quit);
    loop.exec();

    check(completed == names.size(), "all items finished");
    check(failed == 0, "no item failed");
    for (const QString &n : names) {
        check(FileIntegrity::verify(dir.path() + QLatin1Char('/') + QStringLiteral("dl-") + n, digestOf[n]),
              ("verified " + n).toUtf8().constData());
    }

    if (failures == 0)
        std::printf("download_queue: all checks passed\n");
    return failures == 0 ? 0 : 1;
}
