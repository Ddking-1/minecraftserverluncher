// SPDX-License-Identifier: GPL-3.0-only
// Standalone verification for the update manifest parser.
#include <iostream>

#include <QFile>
#include <QJsonArray>
#include <QJsonDocument>
#include <QJsonObject>
#include <QTemporaryDir>

#include "updates/UpdateManifest.h"

namespace {

const char *const kValidHash =
    "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855";

bool check(bool cond, const char *what)
{
    if (!cond)
        std::cerr << "FAIL: " << what << std::endl;
    return cond;
}

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

int main()
{
    // Valid manifest parses fully.
    const UpdateManifest::ParseResult ok = UpdateManifest::parse(validManifest());
    if (!check(ok.ok, "valid manifest should parse")
        || !check(ok.component.name == QStringLiteral("launcher"), "component name")
        || !check(ok.component.version == QStringLiteral("1.0.0"), "component version")
        || !check(ok.component.files.size() == 1, "file count")
        || !check(ok.component.dependencies == QStringList{QStringLiteral("java")}, "dependencies")
        || !check(ok.component.minLauncher == QStringLiteral("0.1.0"), "minLauncher"))
        return 1;

    // Missing name is rejected.
    {
        QJsonObject obj = validManifest();
        obj.remove(QStringLiteral("name"));
        if (!check(!UpdateManifest::parse(obj).ok, "missing name must fail"))
            return 1;
    }

    // Missing files array is rejected.
    {
        QJsonObject obj = validManifest();
        obj.remove(QStringLiteral("files"));
        if (!check(!UpdateManifest::parse(obj).ok, "missing files must fail"))
            return 1;
    }

    // Path traversal is rejected.
    {
        QJsonObject obj = validManifest();
        QJsonArray files = obj.value(QStringLiteral("files")).toArray();
        QJsonObject file = files.at(0).toObject();
        file.insert(QStringLiteral("path"), QStringLiteral("../../etc/passwd"));
        files.replace(0, file);
        obj.insert(QStringLiteral("files"), files);
        if (!check(!UpdateManifest::parse(obj).ok, "path traversal must fail"))
            return 1;
    }

    // Bad hash length is rejected.
    {
        QJsonObject obj = validManifest();
        QJsonArray files = obj.value(QStringLiteral("files")).toArray();
        QJsonObject file = files.at(0).toObject();
        file.insert(QStringLiteral("sha256"), QStringLiteral("deadbeef"));
        files.replace(0, file);
        obj.insert(QStringLiteral("files"), files);
        if (!check(!UpdateManifest::parse(obj).ok, "short sha256 must fail"))
            return 1;
    }

    // parseFile round-trip through disk.
    {
        QTemporaryDir dir;
        if (!check(dir.isValid(), "temporary dir"))
            return 1;
        const QString path = dir.filePath(QStringLiteral("manifest.json"));
        QFile out(path);
        if (!check(out.open(QIODevice::WriteOnly), "open temp manifest"))
            return 1;
        out.write(QJsonDocument(validManifest()).toJson());
        out.close();

        const UpdateManifest::ParseResult fromFile = UpdateManifest::parseFile(path);
        if (!check(fromFile.ok, "parseFile valid")
            || !check(fromFile.component.files.size() == 1, "parseFile file count"))
            return 1;

        const UpdateManifest::ParseResult missing =
            UpdateManifest::parseFile(dir.filePath(QStringLiteral("nope.json")));
        if (!check(!missing.ok, "parseFile missing must fail"))
            return 1;
    }

    std::cout << "OK" << std::endl;
    return 0;
}
