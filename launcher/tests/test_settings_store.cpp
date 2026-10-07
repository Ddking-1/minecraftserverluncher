// SPDX-License-Identifier: GPL-3.0-only
// Qt Test-based verification for the launcher settings store.
#include <QtTest>

#include <QFile>
#include <QTemporaryDir>

#include "core/SettingsStore.h"

class TestSettingsStore : public QTest
{
    Q_OBJECT

private slots:
    void defaults();
    void clamp();
    void saveAndLoad();
    void missingFile();
    void malformedFile();
    void jvmArgSafety();
    void validationBounds();
};

void TestSettingsStore::defaults()
{
    const SettingsStore::Settings def = SettingsStore::defaults();
    QCOMPARE(def.minMemoryMb, 2048);
    QCOMPARE(def.maxMemoryMb, 4096);
}

void TestSettingsStore::clamp()
{
    SettingsStore::Settings s;
    s.minMemoryMb = 100;
    s.maxMemoryMb = 50;
    SettingsStore::clamp(s);
    QCOMPARE(s.minMemoryMb, 512);
    QCOMPARE(s.maxMemoryMb, 512);
}

void TestSettingsStore::saveAndLoad()
{
    QTemporaryDir dir;
    QVERIFY(dir.isValid());
    const QString path = dir.filePath(QStringLiteral("launcher.json"));

    SettingsStore::Settings out;
    out.javaPath = QStringLiteral("C:/java/bin/java.exe");
    out.minMemoryMb = 1024;
    out.maxMemoryMb = 8192;
    out.jvmArgs = QStringList{QStringLiteral("-XX:+UseG1GC"), QStringLiteral("-Dfoo=bar")};

    QVERIFY(SettingsStore::save(path, out));

    SettingsStore::Settings in;
    QVERIFY(SettingsStore::load(path, in));
    QCOMPARE(in.javaPath, out.javaPath);
    QCOMPARE(in.minMemoryMb, out.minMemoryMb);
    QCOMPARE(in.maxMemoryMb, out.maxMemoryMb);
    QCOMPARE(in.jvmArgs, out.jvmArgs);
}

void TestSettingsStore::missingFile()
{
    SettingsStore::Settings in;
    QVERIFY(!SettingsStore::load(QString(), in));
}

void TestSettingsStore::malformedFile()
{
    QTemporaryDir dir;
    QVERIFY(dir.isValid());
    const QString path = dir.filePath(QStringLiteral("bad.json"));
    {
        QFile bad(path);
        QVERIFY(bad.open(QIODevice::WriteOnly));
        bad.write("{not valid json");
    }
    SettingsStore::Settings in;
    QVERIFY(!SettingsStore::load(path, in));
}

void TestSettingsStore::jvmArgSafety()
{
    QVERIFY(SettingsStore::isSafeJvmArg(QStringLiteral("-XX:+UseG1GC")));
    QVERIFY(!SettingsStore::isSafeJvmArg(QString()));
    QVERIFY(!SettingsStore::isSafeJvmArg(QStringLiteral(" -Xmx1G")));
    QVERIFY(!SettingsStore::isSafeJvmArg(QStringLiteral("bad\narg")));
    QVERIFY(!SettingsStore::isSafeJvmArg(QString(2000, QLatin1Char('a'))));
}

void TestSettingsStore::validationBounds()
{
    QString error;
    SettingsStore::Settings good;
    good.minMemoryMb = 2048;
    good.maxMemoryMb = 4096;
    QVERIFY(SettingsStore::validate(good, &error));

    SettingsStore::Settings lowMin;
    lowMin.minMemoryMb = 100;
    lowMin.maxMemoryMb = 4096;
    QVERIFY(!SettingsStore::validate(lowMin, &error));
    QVERIFY(!error.isEmpty());

    SettingsStore::Settings inverted;
    inverted.minMemoryMb = 8192;
    inverted.maxMemoryMb = 4096;
    QVERIFY(!SettingsStore::validate(inverted, &error));

    SettingsStore::Settings unsafe;
    unsafe.minMemoryMb = 2048;
    unsafe.maxMemoryMb = 4096;
    unsafe.jvmArgs = QStringList{QStringLiteral("bad\narg")};
    QVERIFY(!SettingsStore::validate(unsafe, &error));
}

QTEST_APPLESS_MAIN(TestSettingsStore)
#include "test_settings_store.moc"
