// SPDX-License-Identifier: GPL-3.0-only
#include <iostream>

#include <QCryptographicHash>
#include <QDir>
#include <QFile>
#include <QJsonArray>
#include <QJsonDocument>
#include <QJsonObject>
#include <QTemporaryDir>
#include <QUrl>

#include "launch/LaunchTask.h"
#include "launch/steps/ResolveLibraries.h"
#include "minecraft/MinecraftInstance.h"

namespace {

bool check(bool cond, const char *what)
{
    if (!cond)
        std::cerr << "FAIL: " << what << std::endl;
    return cond;
}

QString writeFile(const QString &path, const QByteArray &data)
{
    QDir().mkpath(QFileInfo(path).absolutePath());
    QFile f(path);
    if (!f.open(QIODevice::WriteOnly))
        return QString();
    f.write(data);
    f.close();
    return path;
}

QString sha1(const QByteArray &data)
{
    return QString::fromLatin1(QCryptographicHash::hash(data, QCryptographicHash::Sha1).toHex());
}

bool runStep(MinecraftInstance &instance, QString *failure)
{
    LaunchTask task(&instance);
    ResolveLibraries step(&task, &instance);
    bool ok = false;
    QString err;
    QObject::connect(&step, &ResolveLibraries::readyForLaunch, [&ok]() { ok = true; });
    QObject::connect(&step, &ResolveLibraries::failed, [&err](const QString &reason) { err = reason; });
    step.executeTask();
    if (failure)
        *failure = err;
    return ok && err.isEmpty();
}

bool writeLibraryMetadata(const QString &instanceRoot, const QString &relPath,
                         const QString &url, const QString &hash)
{
    QJsonObject lib;
    lib.insert(QStringLiteral("path"), relPath);
    lib.insert(QStringLiteral("url"), url);
    lib.insert(QStringLiteral("sha1"), hash);

    QJsonArray libs;
    libs.append(lib);

    QJsonObject root;
    root.insert(QStringLiteral("libraries"), libs);

    QFile out(instanceRoot + QStringLiteral("/version.json"));
    if (!out.open(QIODevice::WriteOnly))
        return false;
    out.write(QJsonDocument(root).toJson(QJsonDocument::Compact));
    out.close();
    return true;
}

} // namespace

int main()
{
    QTemporaryDir dir;
    if (!check(dir.isValid(), "temporary dir"))
        return 1;

    const QString instanceRoot = dir.filePath(QStringLiteral("instance"));
    QDir().mkpath(instanceRoot);

    MinecraftInstance instance;
    instance.setInstanceRoot(instanceRoot);

    const QByteArray payload("library-binary");
    const QString srcFile = dir.filePath(QStringLiteral("sources/libA.jar"));
    if (!check(!writeFile(srcFile, payload).isEmpty(), "write source file"))
        return 1;

    const QString relPath = QStringLiteral("com/example/libA.jar");
    if (!check(writeLibraryMetadata(instanceRoot, relPath,
                                    QUrl::fromLocalFile(srcFile).toString(),
                                    sha1(payload)),
               "write metadata"))
        return 1;

    QString failure;
    if (!check(runStep(instance, &failure), "resolve libraries success")
        || !check(failure.isEmpty(), "no failure reason"))
        return 1;

    const QString outPath = QDir(instance.libDir()).filePath(relPath);
    QFile out(outPath);
    if (!check(out.exists(), "library created"))
        return 1;
    if (!check(out.open(QIODevice::ReadOnly), "open created library"))
        return 1;
    if (!check(out.readAll() == payload, "library payload matches"))
        return 1;

    // Path traversal is rejected and must not write outside libraries.
    if (!check(writeLibraryMetadata(instanceRoot, QStringLiteral("../escape.jar"),
                                    QUrl::fromLocalFile(srcFile).toString(),
                                    sha1(payload)),
               "write traversal metadata"))
        return 1;
    if (!check(!runStep(instance, &failure), "path traversal should fail")
        || !check(!failure.isEmpty(), "path traversal failure message"))
        return 1;
    if (!check(!QFile::exists(dir.filePath(QStringLiteral("instance/.minecraft/escape.jar"))),
               "no escaped library written"))
        return 1;

    // Hash mismatch is rejected.
    if (!check(writeLibraryMetadata(instanceRoot, QStringLiteral("com/example/bad.jar"),
                                    QUrl::fromLocalFile(srcFile).toString(),
                                    QStringLiteral("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")),
               "write bad hash metadata"))
        return 1;
    if (!check(!runStep(instance, &failure), "hash mismatch should fail"))
        return 1;

    std::cout << "OK" << std::endl;
    return 0;
}
