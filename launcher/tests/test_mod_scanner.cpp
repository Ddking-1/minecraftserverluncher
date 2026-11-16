// SPDX-License-Identifier: GPL-3.0-only
// Qt Test-based verification for the mod directory scanner.
#include <QtTest>

#include <QDir>
#include <QFile>
#include <QTemporaryDir>

#include "minecraft/ModScanner.h"

class TestModScanner : public QObject
{
    Q_OBJECT

private slots:
    void missingDir();
    void scanMods();
};

void TestModScanner::missingDir()
{
    QTemporaryDir dir;
    QVERIFY(dir.isValid());
    QVERIFY(ModScanner::scan(dir.filePath(QStringLiteral("nope"))).isEmpty());
}

void TestModScanner::scanMods()
{
    QTemporaryDir dir;
    QVERIFY(dir.isValid());

    const QString mods = dir.filePath(QStringLiteral("mods"));
    QDir().mkpath(mods);
    {
        QFile f(mods + QStringLiteral("/zeta.jar"));
        QVERIFY(f.open(QIODevice::WriteOnly));
        f.write("x");
    }
    {
        QFile f(mods + QStringLiteral("/Alpha.jar.disabled"));
        QVERIFY(f.open(QIODevice::WriteOnly));
        f.write("x");
    }
    {
        QFile f(mods + QStringLiteral("/beta.jar"));
        QVERIFY(f.open(QIODevice::WriteOnly));
        f.write("x");
    }
    {
        QFile f(mods + QStringLiteral("/notes.txt"));
        QVERIFY(f.open(QIODevice::WriteOnly));
        f.write("x");
    }

    const QList<ModScanner::ModEntry> entries = ModScanner::scan(mods);
    QCOMPARE(entries.size(), 3);
    QCOMPARE(entries.at(0).displayName, QStringLiteral("Alpha"));
    QVERIFY(!entries.at(0).enabled);
    QCOMPARE(entries.at(1).displayName, QStringLiteral("beta"));
    QVERIFY(entries.at(1).enabled);
    QCOMPARE(entries.at(2).displayName, QStringLiteral("zeta"));
    QVERIFY(entries.at(2).enabled);
    QCOMPARE(entries.at(0).fileName, QStringLiteral("Alpha.jar.disabled"));
}

QTEST_APPLESS_MAIN(TestModScanner)
#include "test_mod_scanner.moc"
