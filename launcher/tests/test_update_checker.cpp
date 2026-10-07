// SPDX-License-Identifier: GPL-3.0-only
// Qt Test-based verification for the update checker.
#include <QtTest>

#include "updates/UpdateChecker.h"

class TestUpdateChecker : public QTest
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
    UpdateManifest::ManifestComponent component(QStringLiteral("0.2.0"), QStringLiteral("0.1.0"));
    component.name = QStringLiteral("launcher");

    QCOMPARE(UpdateChecker::check(QStringLiteral("0.1.0"), component),
             UpdateChecker::Decision::UpdateAvailable);

    QCOMPARE(UpdateChecker::check(QStringLiteral("0.2.0"), component),
             UpdateChecker::Decision::UpToDate);

    UpdateManifest::ManifestComponent incompatible(QStringLiteral("0.3.0"), QStringLiteral("0.2.0"));
    incompatible.name = QStringLiteral("launcher");
    QCOMPARE(UpdateChecker::check(QStringLiteral("0.1.0"), incompatible),
             UpdateChecker::Decision::Incompatible);

    UpdateManifest::ManifestComponent emptyMin(QStringLiteral("0.1.0"), QString());
    emptyMin.name = QStringLiteral("launcher");
    QCOMPARE(UpdateChecker::check(QStringLiteral("0.1.0"), emptyMin),
             UpdateChecker::Decision::UpToDate);
}

QTEST_APPLESS_MAIN(TestUpdateChecker)
#include "test_update_checker.moc"
