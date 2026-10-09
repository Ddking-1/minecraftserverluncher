// SPDX-License-Identifier: GPL-3.0-only
// Qt Test-based verification for the update manifest parser.
#include <QtTest>

#include <QFile>
#include <QJsonArray>
#include <QJsonDocument>
#include <QJsonObject>
#include <QTemporaryDir>

#include "updates/UpdateManifest.h"

class TestUpdateManifest : public QTest
{
    Q_OBJECT

private slots:
    void validManifest();
    void missingName();
    void missingFiles();
    void pathTraversal();
    void badHashLength();
    void parseFileRoundTrip();
};

namespace {
const char *const kValidHash =
    "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855";

QJsonObject validManifest()
{
    QJsonObject file;
    file.insert(QStringLiteral("path"), QStringLiteral("bin/app.exe"));
    file.insert(QStringLiteral("size"), 123);
    file.insert(QStringLiteral("sha256"), QString::fromLatin1(kValidHash));
    file.insert(QStringLiteral("url"), QStringLiteral("https://example.com/app.exe"));

    QJsonArray files;
    files.append(file);

    QJsonArray deps;
    deps.append(QStringLiteral("java"));

    QJsonObject obj;
    obj.insert(QStringLiteral("name"), QStringLiteral("launcher"));
    obj.insert(QStringLiteral("version"), QStringLiteral("1.0.0"));
    obj.insert(QStringLiteral("files"), files);
    obj.insert(QStringLiteral("dependencies"), deps);
    obj.insert(QStringLiteral("min_launcher"), QStringLiteral("0.1.0"));
    obj.insert(QStringLiteral("signature"), QStringLiteral("sig"));
    return obj;
}
} // namespace

void TestUpdateManifest::validManifest()
{
    const UpdateManifest::ParseResult ok = UpdateManifest::parse(validManifest());
    QVERIFY(ok.ok);
    QCOMPARE(ok.component.name, QStringLiteral("launcher"));
    QCOMPARE(ok.component.version, QStringLiteral("1.0.0"));
    QCOMPARE(ok.component.files.size(), 1);
    QCOMPARE(ok.component.dependencies, QStringList{QStringLiteral("java")});
    QCOMPARE(ok.component.minLauncher, QStringLiteral("0.1.0"));
}

void TestUpdateManifest::missingName()
{
    QJsonObject obj = validManifest();
    obj.remove(QStringLiteral("name"));
    const UpdateManifest::ParseResult result = UpdateManifest::parse(obj);
    QVERIFY(!result.ok);
}

void TestUpdateManifest::missingFiles()
{
    QJsonObject obj = validManifest();
    obj.remove(QStringLiteral("files"));
    const UpdateManifest::ParseResult result = UpdateManifest::parse(obj);
    QVERIFY(!result.ok);
}

void TestUpdateManifest::pathTraversal()
{
    QJsonObject obj = validManifest();
    QJsonArray files = obj.value(QStringLiteral("files")).toArray();
    QJsonObject file = files.at(0).toObject();
    file.insert(QStringLiteral("path"), QStringLiteral("../../etc/passwd"));
    files.replace(0, file);
    obj.insert(QStringLiteral("files"), files);
    const UpdateManifest::ParseResult result = UpdateManifest::parse(obj);
    QVERIFY(!result.ok);
}

void TestUpdateManifest::badHashLength()
{
    QJsonObject obj = validManifest();
    QJsonArray files = obj.value(QStringLiteral("files")).toArray();
    QJsonObject file = files.at(0).toObject();
    file.insert(QStringLiteral("sha256"), QStringLiteral("deadbeef"));
    files.replace(0, file);
    obj.insert(QStringLiteral("files"), files);
    const UpdateManifest::ParseResult result = UpdateManifest::parse(obj);
    QVERIFY(!result.ok);
}

void TestUpdateManifest::parseFileRoundTrip()
{
    QTemporaryDir dir;
    QVERIFY(dir.isValid());
    const QString path = dir.filePath(QStringLiteral("manifest.json"));
    {
        QFile out(path);
        QVERIFY(out.open(QIODevice::WriteOnly));
        out.write(QJsonDocument(validManifest()).toJson());
    }

    const UpdateManifest::ParseResult fromFile = UpdateManifest::parseFile(path);
    QVERIFY(fromFile.ok);
    QCOMPARE(fromFile.component.files.size(), 1);

    const UpdateManifest::ParseResult missing =
        UpdateManifest::parseFile(dir.filePath(QStringLiteral("nope.json")));
    QVERIFY(!missing.ok);
}

QTEST_APPLESS_MAIN(TestUpdateManifest)
#include "test_update_manifest.moc"
