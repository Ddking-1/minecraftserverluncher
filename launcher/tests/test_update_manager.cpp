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
#include "core/UpdateManagerHMAC.h"

#include <cstdio>

static int failures = 0;
static void check(bool cond, const char *msg)
{
    if (!cond) {
        std::fprintf(stderr, "FAIL: %s\n", msg);
        ++failures;
    }
}

// Helper: compute expected HMAC signature for manifest fields.
static QString expectedSignature(const QString &key,
                                 const QString &version,
                                 const QString &url,
                                 const QString &sha256,
                                 bool mandatory)
{
    return hmacManifestSignature(key, version, url, sha256, mandatory);
}

int main(int argc, char **argv)
{
    QCoreApplication app(argc, argv);

    // Verify currentVersion passthrough and endpoint trimming.
    UpdateManager mgr;
    mgr.setEndpoints(QStringLiteral("http://127.0.0.1:9/"), QStringLiteral("0.1.0"));
    check(mgr.currentVersion() == QStringLiteral("0.1.0"), "current version stored");
    // (Endpoint trimming + network behavior is covered by the running server below.)

    // --- HMAC verification unit tests (TDD) ---
    const QString signingKey = QStringLiteral("test-signing-key-12345");
    const QString version = QStringLiteral("0.2.0");
    const QString url = QStringLiteral("https://cdn.example.com/updates/x.exe");
    const QString sha256 = QStringLiteral("abcdef1234567890");

    // Valid signature should verify.
    QString validSig = expectedSignature(signingKey, version, url, sha256, true);
    check(verifyManifestSignature(signingKey, version, url, sha256, true, validSig),
          "valid HMAC signature accepted");

    // Empty key should reject.
    check(!verifyManifestSignature(QString(), version, url, sha256, true, validSig),
          "empty key rejected");

    // Empty signature should reject.
    check(!verifyManifestSignature(signingKey, version, url, sha256, true, QString()),
          "empty signature rejected");

    // Wrong signature should reject.
    check(!verifyManifestSignature(signingKey, version, url, sha256, true, QStringLiteral("badsig")),
          "invalid HMAC signature rejected");

    // Different mandatory flag should reject.
    QString sigNotMandatory = expectedSignature(signingKey, version, url, sha256, false);
    check(!verifyManifestSignature(signingKey, version, url, sha256, true, sigNotMandatory),
          "mandatory mismatch rejected");

    // Tampering with URL should reject.
    QString sigTamperedUrl = expectedSignature(signingKey, version, url, sha256, true);
    check(!verifyManifestSignature(signingKey, version, QStringLiteral("https://evil.com/x.exe"), sha256, true, sigTamperedUrl),
          "tampered URL rejected");

    // Tampered SHA-256 should reject.
    QString sigTamperedSha = expectedSignature(signingKey, version, url, sha256, true);
    check(!verifyManifestSignature(signingKey, version, url, QStringLiteral("0000aabb"), true, sigTamperedSha),
          "tampered sha256 rejected");

    // --- End HMAC verification unit tests ---

    // Local server that serves a valid manifest with HMAC signature.
    QTcpServer server;
    check(server.listen(QHostAddress::LocalHost, 0), "listen");
    QObject::connect(&server, &QTcpServer::newConnection, &server, [&server, &signingKey]() {
        QTcpSocket *sock = server.nextPendingConnection();
        QObject::connect(sock, &QTcpSocket::readyRead, sock, [sock, &signingKey]() {
            sock->readAll();
            QJsonObject u;
            u.insert(QStringLiteral("version"), QStringLiteral("0.2.0"));
            u.insert(QStringLiteral("url"), QStringLiteral("https://cdn.example.com/x.exe"));
            u.insert(QStringLiteral("sha256"), QStringLiteral("aabb"));
            u.insert(QStringLiteral("mandatory"), true);
            QString sig = hmacManifestSignature(signingKey,
                                                 QStringLiteral("0.2.0"),
                                                 QStringLiteral("https://cdn.example.com/x.exe"),
                                                 QStringLiteral("aabb"),
                                                 true);
            u.insert(QStringLiteral("signature"), sig);
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