// SPDX-License-Identifier: GPL-3.0-only
// Qt Test-based verification for the file integrity verifier.
#include <QtTest>

#include <QFile>
#include <QTemporaryDir>

#include "core/FileIntegrity.h"

class TestFileIntegrity : public QTest
{
    Q_OBJECT

private slots:
    void emptyFileHash();
    void nonEmptyHash();
    void verifyMatch();
    void verifyCaseInsensitive();
    void verifyMismatch();
    void missingFile();
};

namespace {
const char *const kEmptyHash =
    "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855";
} // namespace

void TestFileIntegrity::emptyFileHash()
{
    QTemporaryDir dir;
    QVERIFY(dir.isValid());
    const QString empty = dir.filePath(QStringLiteral("empty.bin"));
    {
        QFile f(empty);
        QVERIFY(f.open(QIODevice::WriteOnly));
    }
    QCOMPARE(FileIntegrity::sha256File(empty), QString::fromLatin1(kEmptyHash));
}

void TestFileIntegrity::nonEmptyHash()
{
    QTemporaryDir dir;
    QVERIFY(dir.isValid());
    const QString data = dir.filePath(QStringLiteral("data.bin"));
    {
        QFile f(data);
        QVERIFY(f.open(QIODevice::WriteOnly));
        f.write("hello");
    }
    const QString hash = FileIntegrity::sha256File(data);
    QVERIFY(!hash.isEmpty());
}

void TestFileIntegrity::verifyMatch()
{
    QTemporaryDir dir;
    QVERIFY(dir.isValid());
    const QString data = dir.filePath(QStringLiteral("data.bin"));
    {
        QFile f(data);
        QVERIFY(f.open(QIODevice::WriteOnly));
        f.write("hello");
    }
    const QString hash = FileIntegrity::sha256File(data);
    QVERIFY(FileIntegrity::verify(data, hash));
}

void TestFileIntegrity::verifyCaseInsensitive()
{
    QTemporaryDir dir;
    QVERIFY(dir.isValid());
    const QString data = dir.filePath(QStringLiteral("data.bin"));
    {
        QFile f(data);
        QVERIFY(f.open(QIODevice::WriteOnly));
        f.write("hello");
    }
    const QString hash = FileIntegrity::sha256File(data);
    QVERIFY(FileIntegrity::verify(data, hash.toUpper()));
}

void TestFileIntegrity::verifyMismatch()
{
    QTemporaryDir dir;
    QVERIFY(dir.isValid());
    const QString data = dir.filePath(QStringLiteral("data.bin"));
    {
        QFile f(data);
        QVERIFY(f.open(QIODevice::WriteOnly));
        f.write("hello");
    }
    QVERIFY(!FileIntegrity::verify(data, QString::fromLatin1(kEmptyHash)));
}

void TestFileIntegrity::missingFile()
{
    QVERIFY(FileIntegrity::sha256File(QString()).isEmpty());
    QVERIFY(!FileIntegrity::verify(QString(), QString()));
}

QTEST_APPLESS_MAIN(TestFileIntegrity)
#include "test_file_integrity.moc"
