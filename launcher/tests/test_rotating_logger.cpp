// SPDX-License-Identifier: GPL-3.0-only
// Qt Test-based verification for the rotating file logger.
#include <QtTest>

#include <QFile>
#include <QTemporaryDir>

#include "core/RotatingFileLogger.h"

class TestRotatingLogger : public QTest
{
    Q_OBJECT

private slots:
    void rotation();
    void retentionCap();
};

void TestRotatingLogger::rotation()
{
    QTemporaryDir dir;
    QVERIFY(dir.isValid());
    const QString logFile = dir.filePath(QStringLiteral("launcher.log"));

    // Small threshold forces rotation on the second write.
    QVERIFY(RotatingFileLogger::log(logFile, QStringLiteral("first line"), 8, 2));
    QVERIFY(RotatingFileLogger::log(logFile, QStringLiteral("second line"), 8, 2));

    QVERIFY(QFile::exists(logFile + QStringLiteral(".1")));
}

void TestRotatingLogger::retentionCap()
{
    QTemporaryDir dir;
    QVERIFY(dir.isValid());
    const QString logFile = dir.filePath(QStringLiteral("launcher.log"));

    for (int i = 0; i < 5; ++i)
        RotatingFileLogger::log(logFile, QStringLiteral("line %1").arg(i), 1, 2);

    QVERIFY(QFile::exists(logFile + QStringLiteral(".1")));
    QVERIFY(QFile::exists(logFile + QStringLiteral(".2")));
    QVERIFY(!QFile::exists(logFile + QStringLiteral(".3")));
}

QTEST_APPLESS_MAIN(TestRotatingLogger)
#include "test_rotating_logger.moc"
