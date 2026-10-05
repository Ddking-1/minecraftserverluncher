// SPDX-License-Identifier: GPL-3.0-only
#include <QCoreApplication>
#include <QTemporaryDir>

#include "auth/BackendAuthClient.h"
#include "core/SkinClient.h"

#include <cstdio>

static int failures = 0;

static void check(bool cond, const char *msg)
{
    if (!cond) {
        std::fprintf(stderr, "FAIL: %s\n", msg);
        ++failures;
    }
}

class ResultSpy : public QObject
{
    Q_OBJECT
public:
    explicit ResultSpy(SkinClient *client) : m_client(client)
    {
        connect(client, &SkinClient::uploadSucceeded, this, [this](const QString &) { m_success = true; });
        connect(client, &SkinClient::uploadFailed, this, [this](const QString &r) { m_failed = true; m_reason = r; });
    }
    bool m_success = false;
    bool m_failed = false;
    QString m_reason;

private:
    SkinClient *m_client;
};

int main(int argc, char **argv)
{
    QCoreApplication app(argc, argv);

    // Dimension validation is the client-side gate.
    check(SkinClient::isSupportedSize(64, 64), "64x64 supported");
    check(SkinClient::isSupportedSize(64, 32), "64x32 supported");
    check(!SkinClient::isSupportedSize(32, 32), "32x32 rejected");
    check(!SkinClient::isSupportedSize(128, 128), "128x128 rejected");

    BackendAuthClient backend;
    backend.setBackendUrl(QStringLiteral("http://127.0.0.1:1"));
    SkinClient client(&backend);
    ResultSpy spy(&client);

    // Empty token must fail fast without a network call.
    client.uploadSkin(QString(), QStringLiteral("nope.png"), QStringLiteral("default"));
    check(spy.m_failed, "empty token fails");
    check(spy.m_reason.contains(QStringLiteral("logged in")), "empty token reason");

    // Missing file must fail without a network call.
    spy.m_failed = false;
    QTemporaryDir dir;
    client.uploadSkin(QStringLiteral("token"), dir.path() + QStringLiteral("/missing.png"),
                      QStringLiteral("default"));
    check(spy.m_failed, "missing file fails");

    if (failures == 0)
        std::printf("skin_client: all checks passed\n");
    return failures == 0 ? 0 : 1;
}

#include "test_skin_client.moc"
