// SPDX-License-Identifier: GPL-3.0-only
// Qt Test-based contract test: verifies BackendAuthClient sends the expected
// request path and JSON body for a login against a local fake HTTP server.
#include <QtTest>

#include <QByteArray>
#include <QCoreApplication>
#include <QEventLoop>
#include <QHostAddress>
#include <QJsonDocument>
#include <QJsonObject>
#include <QString>
#include <QStringList>
#include <QTcpServer>
#include <QTcpSocket>
#include <QTimer>

#include "auth/BackendAuthClient.h"

class FakeServer : public QObject
{
    Q_OBJECT
public:
    explicit FakeServer(QObject *parent = nullptr) : QObject(parent)
    {
        if (!m_server.listen(QHostAddress::LocalHost, 0))
            qFatal("cannot listen");
        connect(&m_server, &QTcpServer::newConnection, this, [this]() {
            QTcpSocket *socket = m_server.nextPendingConnection();
            connect(socket, &QTcpSocket::readyRead, this, [this, socket]() {
                const QByteArray request = socket->readAll();
                const QStringList lines = QString::fromUtf8(request).split(QStringLiteral("\r\n"));
                if (!lines.isEmpty()) {
                    const QStringList first = lines.at(0).split(QLatin1Char(' '));
                    if (first.size() >= 2)
                        lastPath = first.at(1);
                }
                const int bodyStart = request.indexOf("\r\n\r\n");
                if (bodyStart >= 0)
                    lastBody = request.mid(bodyStart + 4);

                QByteArray response = "HTTP/1.1 200 OK\r\n";
                response += "Content-Type: application/json\r\n";
                response += "Content-Length: " + QByteArray::number(responseBody.size()) + "\r\n";
                response += "Connection: close\r\n\r\n";
                response += responseBody;
                socket->write(response);
                socket->flush();
                socket->disconnectFromHost();
            });
        });
    }

    quint16 port() const { return m_server.serverPort(); }

    QTcpServer m_server;
    QString lastPath;
    QByteArray lastBody;
    QByteArray responseBody;
};

class TestBackendContract : public QTest
{
    Q_OBJECT

private slots:
    void loginRequest();
};

void TestBackendContract::loginRequest()
{
    FakeServer server;
    BackendAuthClient client;
    client.setBackendUrl(QStringLiteral("http://127.0.0.1:%1").arg(server.port()));
    server.responseBody =
        "{\"access_token\":\"a\",\"refresh_token\":\"r\",\"token_type\":\"Bearer\",\"expires_in\":900}";

    QEventLoop loop;
    bool done = false;
    AuthResult result;
    client.login(QStringLiteral("bob"), QStringLiteral("secret"), QStringLiteral("dev-1"),
                 QStringLiteral("Desktop"), [&](const AuthResult &r) {
                     result = r;
                     done = true;
                     loop.quit();
                 });
    QTimer::singleShot(5000, &loop, &QEventLoop::quit);
    loop.exec();

    QVERIFY(done);
    QVERIFY(result.success);
    QCOMPARE(result.accessToken, QStringLiteral("a"));
    QCOMPARE(result.refreshToken, QStringLiteral("r"));
    QCOMPARE(result.expiresIn, 900);
    QCOMPARE(server.lastPath, QStringLiteral("/api/v1/auth/login"));
    const QJsonObject body = QJsonDocument::fromJson(server.lastBody).object();
    QCOMPARE(body.value(QStringLiteral("username")).toString(), QStringLiteral("bob"));
    QCOMPARE(body.value(QStringLiteral("password")).toString(), QStringLiteral("secret"));
    QCOMPARE(body.value(QStringLiteral("device_id")).toString(), QStringLiteral("dev-1"));
}

QTEST_APPLESS_MAIN(TestBackendContract)
#include "test_backend_contract.moc"
