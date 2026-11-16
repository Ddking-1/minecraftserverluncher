// SPDX-License-Identifier: GPL-3.0-only
// Qt Test-based verification for the update checker.
#include <QtTest>

#include "updates/UpdateChecker.h"

class TestUpdateChecker : public QObject
{
    Q_OBJECT

private slots:
    void compareVersions();
    void checkUpdate();
};

void TestUpdateChecker::compareVersions()
{
    QCOMPARE(UpdateChecker::compareVersions(QStringLiteral("1.2.10"), QStringLiteral("1.2.9")), 1);
    QCOMPARE(UpdateChecker::compareVersions(QStringLiteral("1.2"), QStringLiteral("1.2.0")), 0);
    QCOMPARE(UpdateChecker::compareVersions(QStringLiteral("0.9"), QStringLiteral("1.0")), -1);
    QCOMPARE(UpdateChecker::compareVersions(QStringLiteral("2"), QStringLiteral("1.9.9")), 1);
}

void TestUpdateChecker::checkUpdate()
{
    UpdateManifest::ManifestComponent component;
    component.name = QStringLiteral("launcher");
    component.version = QStringLiteral("0.2.0");
    component.minLauncher = QStringLiteral("0.1.0");

    QCOMPARE(UpdateChecker::check(QStringLiteral("0.1.0"), component),
             UpdateChecker::Decision::UpdateAvailable);

    QCOMPARE(UpdateChecker::check(QStringLiteral("0.2.0"), component),
             UpdateChecker::Decision::UpToDate);

    UpdateManifest::ManifestComponent incompatible;
    incompatible.name = QStringLiteral("launcher");
    incompatible.version = QStringLiteral("0.3.0");
    incompatible.minLauncher = QStringLiteral("0.2.0");
    QCOMPARE(UpdateChecker::check(QStringLiteral("0.1.0"), incompatible),
             UpdateChecker::Decision::Incompatible);

    UpdateManifest::ManifestComponent emptyMin;
    emptyMin.name = QStringLiteral("launcher");
    emptyMin.version = QStringLiteral("0.1.0");
    QCOMPARE(UpdateChecker::check(QStringLiteral("0.1.0"), emptyMin),
             UpdateChecker::Decision::UpToDate);
}

QTEST_APPLESS_MAIN(TestUpdateChecker)
#include "test_update_checker.moc"
