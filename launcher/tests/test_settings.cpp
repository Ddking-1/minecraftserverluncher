// SPDX-License-Identifier: GPL-3.0-only
// Qt Test-based verification for InstancePersistence round-tripping.
#include <QtTest>

#include <QJsonObject>
#include <QString>
#include <QTemporaryDir>

#include "core/InstancePersistence.h"

class TestSettings : public QObject
{
    Q_OBJECT

private slots:
    void settingsRoundTrip();
    void metadataRoundTrip();
    void fileRoundTrip();
};

void TestSettings::settingsRoundTrip()
{
    // Settings round-trip.
    LauncherSettings settings;
    settings.javaPath = QStringLiteral("/usr/bin/java");
    settings.maxMemoryMb = 8192;
    settings.minMemoryMb = 1024;
    settings.jvmArgs = {QStringLiteral("-Xmx8G"), QStringLiteral("-XX:+UseG1GC")};

    const LauncherSettings loaded = InstancePersistence::settingsFromJson(
        InstancePersistence::settingsToJson(settings));
    QCOMPARE(loaded.javaPath, settings.javaPath);
    QCOMPARE(loaded.maxMemoryMb, settings.maxMemoryMb);
    QCOMPARE(loaded.minMemoryMb, settings.minMemoryMb);
    QCOMPARE(loaded.jvmArgs, settings.jvmArgs);
}

void TestSettings::metadataRoundTrip()
{
    // Metadata round-trip.
    InstanceMetadata meta;
    meta.id = QStringLiteral("inst-1");
    meta.uuid = QStringLiteral("uuid-1");
    meta.name = QStringLiteral("My Instance");
    meta.type = QStringLiteral("MinecraftInstance");
    meta.traits = {QStringLiteral("a"), QStringLiteral("b")};
    meta.lastLaunchMs = 123;
    meta.totalTimePlayedMs = 456;

    const InstanceMetadata loadedMeta = InstancePersistence::metadataFromJson(
        InstancePersistence::metadataToJson(meta));
    QCOMPARE(loadedMeta.id, meta.id);
    QCOMPARE(loadedMeta.uuid, meta.uuid);
    QCOMPARE(loadedMeta.name, meta.name);
    QCOMPARE(loadedMeta.type, meta.type);
    QCOMPARE(loadedMeta.traits, meta.traits);
    QCOMPARE(loadedMeta.lastLaunchMs, meta.lastLaunchMs);
    QCOMPARE(loadedMeta.totalTimePlayedMs, meta.totalTimePlayedMs);
}

void TestSettings::fileRoundTrip()
{
    QTemporaryDir dir;
    QVERIFY(dir.isValid());
    const QString path = dir.filePath(QStringLiteral("settings.json"));
    QVERIFY(InstancePersistence::writeJsonFile(
               path, InstancePersistence::settingsToJson(LauncherSettings())));

    QJsonObject out;
    QVERIFY(InstancePersistence::readJsonFile(path, out));
    QVERIFY(out.contains(QStringLiteral("max_memory_mb")));
}

QTEST_APPLESS_MAIN(TestSettings)
#include "test_settings.moc"
