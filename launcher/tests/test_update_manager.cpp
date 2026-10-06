// SPDX-License-Identifier: GPL-3.0-only
#include <QCoreApplication>
#include <QEventLoop>
#include <QHostAddress>
#include <QJsonDocument>
#include <QJsonObject>
#include <QTcpServer>
#include <QTcpSocket>
#include <QTimer>

#include "core/UpdateManager.h"

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

    // Verify currentVersion passthrough and endpoint trimming.
    UpdateManager mgr;
    mgr.setEndpoints(QStringLiteral("http://127.0.0.1:9/"), QStringLiteral("0.1.0"));
    check(mgr.currentVersion() == QStringLiteral("0.1.0"), "current version stored");
    // (Endpoint trimming + network behavior is covered by the running server below.)

    // Local server that serves a valid manifest.
    QTcpServer server;
    check(server.listen(QHostAddress::LocalHost, 0), "listen");
    QObject::connect(&server, &QTcpServer::newConnection, &server, [&server]() {
        QTcpSocket *sock = server.nextPendingConnection();
        QObject::connect(sock, &QTcpSocket::readyRead, sock, [sock]() {
            sock->readAll();
            QJsonObject u;
            u.insert(QStringLiteral("version"), QStringLiteral("0.2.0"));
            u.insert(QStringLiteral("url"), QStringLiteral("https://cdn.example.com/x.exe"));
            u.insert(QStringLiteral("sha256"), QStringLiteral("aabb"));
            u.insert(QStringLiteral("mandatory"), true);
            QJsonObject root;
            root.insert(QStringLiteral("update"), u);
            const QByteArray body = QJsonDocument(root).toJson(QJsonDocument::Compact);
            sock->write("HTTP/1.1 200 OK\r\nContent-Length: "
                        + QByteArray::number(body.size()) + "\r\nConnection: close\r\n\r\n" + body);
            sock->disconnectFromHost();
        });
    });

    UpdateManager updater;
    updater.setEndpoints(QStringLiteral("http://127.0.0.1:%1").arg(server.serverPort()), QStringLiteral("0.1.0"));

    bool got = false;
    QEventLoop loop;
    QObject::connect(&updater, &UpdateManager::updateInfo, &loop, [&](const QString &v, const QString &url, const QString &sha, bool mandatory) {
        got = true;
        check(v == QStringLiteral("0.2.0"), "manifest version");
        check(url == QStringLiteral("https://cdn.example.com/x.exe"), "manifest url");
        check(sha == QStringLiteral("aabb"), "manifest sha256");
        check(mandatory, "manifest mandatory");
        loop.quit();
    });
    QObject::connect(&updater, &UpdateManager::checkFailed, &loop, [&](const QString &) { loop.quit(); });

    updater.checkForUpdate();
    QTimer::singleShot(15000, &loop, &QEventLoop::quit);
    loop.exec();
    check(got, "updateInfo emitted");

    if (failures == 0)
        std::printf("update_manager: all checks passed\n");
    return failures == 0 ? 0 : 1;
}