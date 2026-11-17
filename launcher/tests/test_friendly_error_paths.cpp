// SPDX-License-Identifier: GPL-3.0-only
// The two error sources that used to emit raw English text (the game process
// launcher and the update checker) must hand the UI a plain Hungarian
// explanation, so no technical string can reach a player.
#include <QCoreApplication>
#include <QDir>
#include <QEventLoop>
#include <QHostAddress>
#include <QTcpServer>
#include <QTcpSocket>
#include <QTemporaryDir>
#include <QTimer>

#include "core/UpdateManager.h"
#include "game/MinecraftLauncher.h"

#include <cstdio>

static int failures = 0;
static void check(bool cond, const char *msg)
{
    if (!cond) {
        std::fprintf(stderr, "FAIL: %s\n", msg);
        ++failures;
    }
}

namespace {

// True when the message is a plain Hungarian sentence: it must contain at
// least one of the Hungarian action words used by FriendlyError and it must not
// be the raw technical text (which is English).
bool looksFriendly(const QString &message)
{
    static const char *const kHungarian[] = {
        "nem sikerült", "nem található", "nem indult", "próbáld", "ellenőrizd",
        "telepíts",     "válassz",      "indítsd",    "jelentkezz", "csökkentsd",
        "frissítsd",    "aláírás",      "frissítés",  "leállt",     "hiba",
    };
    for (const char *word : kHungarian) {
        if (message.contains(QString::fromUtf8(word)))
            return true;
    }
    return false;
}

// Minimal one-shot HTTP server used to feed UpdateManager a fixed body.
class OneShotHttpServer : public QTcpServer
{
public:
    explicit OneShotHttpServer(QByteArray body) : m_body(std::move(body))
    {
        connect(this, &QTcpServer::newConnection, this, [this]() {
            QTcpSocket *sock = nextPendingConnection();
            connect(sock, &QTcpSocket::readyRead, sock, [this, sock]() {
                sock->readAll();
                sock->write("HTTP/1.1 200 OK\r\nContent-Length: " +
                            QByteArray::number(m_body.size()) +
                            "\r\nConnection: close\r\n\r\n" + m_body);
                sock->disconnectFromHost();
            });
        });
    }

private:
    QByteArray m_body;
};

} // namespace

int main(int argc, char **argv)
{
    QCoreApplication app(argc, argv);

    // --- MinecraftLauncher --------------------------------------------------
    {
        MinecraftLauncher launcher;
        QString lastError;
        int errors = 0;
        QObject::connect(&launcher, &MinecraftLauncher::launchError, &launcher,
                         [&lastError, &errors](const QString &error) {
                             ++errors;
                             lastError = error;
                         });

        launcher.launchGame(QString(), QStringLiteral("player"), QString(), QString(),
                            QStringLiteral("C:/games/inst"));
        check(errors >= 1, "missing parameters report an error");
        check(looksFriendly(lastError), "missing parameters explained in Hungarian");
        check(!lastError.contains(QStringLiteral("Invalid parameters")),
              "raw parameter text is not shown");

        QTemporaryDir gameDir;
        check(gameDir.isValid(), "temporary game directory");

        launcher.launchGame(QStringLiteral("java"), QStringLiteral("player"), QString(),
                            QString(), QStringLiteral("C:/definitely/missing/game-dir"));
        check(errors >= 2, "missing game directory reports an error");
        check(looksFriendly(lastError), "missing game directory explained in Hungarian");
        check(!lastError.contains(QStringLiteral("does not exist")),
              "raw directory text is not shown");

        // A Java path that cannot start must still be explained, never dumped.
        launcher.launchGame(QStringLiteral("C:/definitely/missing/java.exe"),
                            QStringLiteral("player"), QString(), QString(),
                            gameDir.path());
        check(errors >= 3, "unstartable Java reports an error");
        check(looksFriendly(lastError), "unstartable Java explained in Hungarian");
        check(!lastError.contains(QStringLiteral("Failed to start")),
              "raw process text is not shown");
    }

    // --- UpdateManager ------------------------------------------------------
    {
        OneShotHttpServer server(QByteArrayLiteral("{\"not_an_update\":true}"));
        check(server.listen(QHostAddress::LocalHost), "update server listening");

        UpdateManager updater;
        updater.setEndpoints(QStringLiteral("http://127.0.0.1:%1").arg(server.serverPort()),
                             QStringLiteral("0.1.0"));
        updater.setSigningKey(QStringLiteral("test-key"));

        QString reason;
        QEventLoop loop;
        QObject::connect(&updater, &UpdateManager::checkFailed, &loop,
                         [&reason, &loop](const QString &text) {
                             reason = text;
                             loop.quit();
                         });
        QObject::connect(&updater, &UpdateManager::updateInfo, &loop,
                         [&loop](const QString &, const QString &, const QString &, bool) {
                             loop.quit();
                         });
        updater.checkForUpdate();
        QTimer::singleShot(10000, &loop, &QEventLoop::quit);
        loop.exec();

        check(!reason.isEmpty(), "broken manifest is rejected");
        check(looksFriendly(reason), "broken manifest explained in Hungarian");
        check(!reason.contains(QStringLiteral("update object")),
              "raw manifest text is not shown");
    }

    {
        // Signed-looking manifest with a wrong signature: the message must be
        // the safe "not installing" explanation, not the technical wording.
        OneShotHttpServer server(QByteArrayLiteral(
            "{\"update\":{\"version\":\"0.2.0\",\"url\":\"https://cdn.example.com/x.exe\","
            "\"sha256\":\"aabb\",\"mandatory\":true,\"signature\":\"deadbeef\"}}"));
        check(server.listen(QHostAddress::LocalHost), "signature server listening");

        UpdateManager updater;
        updater.setEndpoints(QStringLiteral("http://127.0.0.1:%1").arg(server.serverPort()),
                             QStringLiteral("0.1.0"));
        updater.setSigningKey(QStringLiteral("test-key"));

        QString reason;
        QEventLoop loop;
        QObject::connect(&updater, &UpdateManager::checkFailed, &loop,
                         [&reason, &loop](const QString &text) {
                             reason = text;
                             loop.quit();
                         });
        QObject::connect(&updater, &UpdateManager::updateInfo, &loop,
                         [&loop](const QString &, const QString &, const QString &, bool) {
                             loop.quit();
                         });
        updater.checkForUpdate();
        QTimer::singleShot(10000, &loop, &QEventLoop::quit);
        loop.exec();

        check(!reason.isEmpty(), "unverified manifest is rejected");
        check(looksFriendly(reason), "unverified manifest explained in Hungarian");
        check(reason.contains(QString::fromUtf8("aláírás")),
              "unverified manifest names the signature problem");
        check(!reason.contains(QStringLiteral("signature verification")),
              "raw signature text is not shown");
    }

    if (failures == 0)
        std::printf("friendly_error_paths: all checks passed\n");
    return failures == 0 ? 0 : 1;
}
