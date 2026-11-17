// SPDX-License-Identifier: GPL-3.0-only
#include <iostream>

#include <QCryptographicHash>
#include <QDir>
#include <QFile>
#include <QFileInfo>
#include <QJsonDocument>
#include <QJsonObject>
#include <QTemporaryDir>
#include <QUrl>

#include "launch/LaunchTask.h"
#include "launch/steps/ResolveAssets.h"
#include "minecraft/MinecraftInstance.h"

namespace {

bool check(bool cond, const char *what)
{
    if (!cond)
        std::cerr << "FAIL: " << what << std::endl;
    return cond;
}

QString sha1(const QByteArray &data)
{
    return QString::fromLatin1(QCryptographicHash::hash(data, QCryptographicHash::Sha1).toHex());
}

bool writeFile(const QString &path, const QByteArray &data)
{
    QDir().mkpath(QFileInfo(path).absolutePath());
    QFile f(path);
    if (!f.open(QIODevice::WriteOnly))
        return false;
    f.write(data);
    f.close();
    return true;
}

bool runStep(MinecraftInstance &instance, QString *failure)
{
    LaunchTask task(&instance);
    ResolveAssets step(&task, &instance);
    bool ok = false;
    QString err;
    QObject::connect(&step, &ResolveAssets::readyForLaunch, [&ok]() { ok = true; });
    QObject::connect(&step, &ResolveAssets::failed, [&err](const QString &reason) { err = reason; });
    step.executeTask();
    if (failure)
        *failure = err;
    return ok && err.isEmpty();
}

bool writeVersionMetadata(const QString &instanceRoot, const QString &indexId,
                          const QString &indexURL, const QString &indexSha1,
                          const QString &objectsBase)
{
    QJsonObject assetIndex;
    assetIndex.insert(QStringLiteral("id"), indexId);
    assetIndex.insert(QStringLiteral("url"), indexURL);
    assetIndex.insert(QStringLiteral("sha1"), indexSha1);
    assetIndex.insert(QStringLiteral("objects_base_url"), objectsBase);

    QJsonObject root;
    root.insert(QStringLiteral("asset_index"), assetIndex);

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

    const QByteArray assetPayload("asset-content");
    const QString assetHash = sha1(assetPayload);

    const QString objectSource = dir.filePath(
        QStringLiteral("source-objects/%1/%2").arg(assetHash.left(2), assetHash));
    if (!check(writeFile(objectSource, assetPayload), "write object source"))
        return 1;

    QJsonObject object;
    object.insert(QStringLiteral("hash"), assetHash);
    object.insert(QStringLiteral("size"), assetPayload.size());

    QJsonObject objects;
    objects.insert(QStringLiteral("minecraft/sounds/test.ogg"), object);

    QJsonObject indexRoot;
    indexRoot.insert(QStringLiteral("objects"), objects);

    const QByteArray indexBytes = QJsonDocument(indexRoot).toJson(QJsonDocument::Compact);
    const QString indexPath = dir.filePath(QStringLiteral("source-index/index.json"));
    if (!check(writeFile(indexPath, indexBytes), "write asset index source"))
        return 1;

    if (!check(writeVersionMetadata(instanceRoot,
                                    QStringLiteral("test-index"),
                                    QUrl::fromLocalFile(indexPath).toString(),
                                    sha1(indexBytes),
                                    QUrl::fromLocalFile(dir.filePath(QStringLiteral("source-objects"))).toString()),
               "write version metadata"))
        return 1;

    QString failure;
    if (!check(runStep(instance, &failure), "resolve assets success")
        || !check(failure.isEmpty(), "no failure reason"))
        return 1;

    const QString indexOut = QDir(instance.assetsDir()).filePath(QStringLiteral("indexes/test-index.json"));
    const QString objectOut = QDir(instance.assetsDir())
                                  .filePath(QStringLiteral("objects/%1/%2").arg(assetHash.left(2), assetHash));

    if (!check(QFile::exists(indexOut), "asset index written")
        || !check(QFile::exists(objectOut), "asset object written"))
        return 1;

    QFile obj(objectOut);
    if (!check(obj.open(QIODevice::ReadOnly), "open object output")
        || !check(obj.readAll() == assetPayload, "asset payload matches"))
        return 1;

    // Hash mismatch in index metadata must fail.
    if (!check(writeVersionMetadata(instanceRoot,
                                    QStringLiteral("bad-index"),
                                    QUrl::fromLocalFile(indexPath).toString(),
                                    QStringLiteral("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
                                    QUrl::fromLocalFile(dir.filePath(QStringLiteral("source-objects"))).toString()),
               "write bad index metadata"))
        return 1;
    if (!check(!runStep(instance, &failure), "asset index hash mismatch should fail"))
        return 1;

    std::cout << "OK" << std::endl;
    return 0;
}
