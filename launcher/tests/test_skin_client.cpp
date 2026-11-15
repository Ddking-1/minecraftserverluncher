// SPDX-License-Identifier: GPL-3.0-only
#include <QCoreApplication>
#include <QFile>
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

int main(int argc, char **argv)
{
    QCoreApplication app(argc, argv);

    // Dimension validation is the client-side gate.
    check(SkinClient::isSupportedSize(64, 64), "64x64 supported");
    check(SkinClient::isSupportedSize(64, 32), "64x32 supported");
    check(!SkinClient::isSupportedSize(32, 32), "32x32 rejected");
    check(!SkinClient::isSupportedSize(128, 128), "128x128 rejected");

    BackendAuthClient backend;
    // The client refuses plaintext backends.
    backend.setBackendUrl(QStringLiteral("http://127.0.0.1:1"));
    check(backend.backendUrl().isEmpty(), "non-HTTPS backend rejected");

    SkinClient client(&backend);
    QTemporaryDir dir;
    check(dir.isValid(), "temp dir created");

    bool called = false;
    bool success = true;
    QString error;

    // No backend configured: fail fast without a network call.
    client.uploadSkin(dir.path() + QStringLiteral("/missing.png"),
                      [&called, &success, &error](bool ok, QString, QString err) {
                          called = true;
                          success = ok;
                          error = err;
                      });
    check(called, "callback invoked without backend");
    check(!success, "upload fails without backend");
    check(error.contains(QStringLiteral("backend")), "reason mentions backend");

    // With a backend and a valid token, a missing file fails locally.
    backend.setBackendUrl(QStringLiteral("https://127.0.0.1:1"));
    backend.setAccessToken(QStringLiteral("test-token"));
    called = false;
    success = true;
    error.clear();
    client.uploadSkin(dir.path() + QStringLiteral("/missing.png"),
                      [&called, &success, &error](bool ok, QString, QString err) {
                          called = true;
                          success = ok;
                          error = err;
                      });
    check(called && !success, "missing file fails");
    check(error.contains(QStringLiteral("open")), "missing file reason");

    // A file that is not a PNG is rejected before any network request.
    const QString badPath = dir.path() + QStringLiteral("/bad.png");
    QFile bad(badPath);
    check(bad.open(QIODevice::WriteOnly), "write bad fixture");
    bad.write("this is not a PNG");
    bad.close();

    called = false;
    success = true;
    error.clear();
    client.uploadSkin(badPath,
                      [&called, &success, &error](bool ok, QString, QString err) {
                          called = true;
                          success = ok;
                          error = err;
                      });
    check(called && !success, "invalid PNG fails");
    check(error.contains(QStringLiteral("PNG")), "invalid PNG reason");

    if (failures == 0)
        std::printf("skin_client: all checks passed\n");
    return failures == 0 ? 0 : 1;
}
