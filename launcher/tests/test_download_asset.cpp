// SPDX-License-Identifier: GPL-3.0-only
#include <QCoreApplication>
#include <QEventLoop>
#include <QFile>
#include <QFileInfo>
#include <QHostAddress>
#include <QTcpServer>
#include <QTcpSocket>
#include <QTemporaryDir>
#include <QTimer>
#include <QUrl>

#include "core/DownloadAsset.h"
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

    // Spin up a tiny local HTTP server on a random port to serve a known file,
    // then download it and verify integrity end to end.
    QTemporaryDir dir;
    check(dir.isValid(), "temp dir");

    const QString payload = QStringLiteral("hello-download-12345");
    const QString payloadPath = dir.path() + QStringLiteral("/src.txt");
    {
        QFile f(payloadPath);
        check(f.open(QIODevice::WriteOnly), "open src");
        f.write(payload.toUtf8());
    }
    const QString digest = FileIntegrity::sha256File(payloadPath);
    check(!digest.isEmpty(), "src digest computed");

    QTcpServer server;
    check(server.listen(QHostAddress::LocalHost, 0), "listen");
    QObject::connect(&server, &QTcpServer::newConnection, &server, [&server, &payload]() {
        QTcpSocket *sock = server.nextPendingConnection();
        QObject::connect(sock, &QTcpSocket::readyRead, sock, [sock, &payload]() {
            sock->readAll(); // consume request
            const QByteArray body = payload.toUtf8();
            sock->write("HTTP/1.1 200 OK\r\nContent-Length: "
                        + QByteArray::number(body.size()) + "\r\nConnection: close\r\n\r\n" + body);
            sock->disconnectFromHost();
        });
    });
    const QUrl url(QStringLiteral("http://127.0.0.1:%1/asset.bin").arg(server.serverPort()));

    const QString destPath = dir.path() + QStringLiteral("/out.bin");
    DownloadAsset downloader;
    int state = 0; // 0 pending, 1 done, 2 failed
    QEventLoop loop;
    QObject::connect(&downloader, &DownloadAsset::finished, &loop, [&](const QString &p) {
        check(p == QFileInfo(destPath).absoluteFilePath(), "finished path");
        state = 1;
        loop.quit();
    });
    QObject::connect(&downloader, &DownloadAsset::failed, &loop, [&](const QString &) {
        state = 2;
        loop.quit();
    });

    downloader.start(url, destPath, digest);
    QTimer::singleShot(15000, &loop, &QEventLoop::quit);
    loop.exec();

    check(state == 1, "download finished successfully");
    check(FileIntegrity::verify(destPath, digest), "downloaded file verified");

    // Wrong digest must fail and must NOT leave a file behind.
    const QString badDest = dir.path() + QStringLiteral("/bad.bin");
    DownloadAsset bad;
    int badState = 0;
    QObject::connect(&bad, &DownloadAsset::failed, &bad, [&](const QString &) { badState = 2; });
    QEventLoop badLoop;
    QObject::connect(&bad, &DownloadAsset::finished, &badLoop, [&](const QString &) { badState = 1; badLoop.quit(); });
    QObject::connect(&bad, &DownloadAsset::failed, &badLoop, [&](const QString &) { badLoop.quit(); });
    bad.start(url, badDest, QStringLiteral("0000000000000000000000000000000000000000000000000000000000000000"));
    QTimer::singleShot(15000, &badLoop, &QEventLoop::quit);
    badLoop.exec();
    check(badState != 1, "bad digest not accepted");
    check(!QFile::exists(badDest), "bad digest leaves no file");

    if (failures == 0)
        std::printf("download_asset: all checks passed\n");
    return failures == 0 ? 0 : 1;
}
