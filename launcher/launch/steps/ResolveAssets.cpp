// SPDX-License-Identifier: GPL-3.0-only
#include "ResolveAssets.h"

#include <QCryptographicHash>
#include <QDir>
#include <QEventLoop>
#include <QFile>
#include <QFileInfo>
#include <QJsonDocument>
#include <QJsonObject>
#include <QNetworkAccessManager>
#include <QNetworkReply>
#include <QNetworkRequest>
#include <QRegularExpression>
#include <QSaveFile>
#include <QUrl>

#include "minecraft/MinecraftInstance.h"

namespace {

constexpr int kDownloadRetries = 3;

bool isSha1(const QString &hash)
{
    static const QRegularExpression re(QStringLiteral("^[A-Fa-f0-9]{40}$"));
    return re.match(hash).hasMatch();
}

QString sha1File(const QString &path)
{
    QFile file(path);
    if (!file.open(QIODevice::ReadOnly))
        return QString();
    QCryptographicHash hash(QCryptographicHash::Sha1);
    if (!hash.addData(&file))
        return QString();
    return QString::fromLatin1(hash.result().toHex());
}

bool fetchUrl(const QUrl &url, QByteArray &out, QString &error)
{
    if (!url.isValid()) {
        error = QStringLiteral("invalid url");
        return false;
    }
    if (url.isLocalFile()) {
        QFile f(url.toLocalFile());
        if (!f.open(QIODevice::ReadOnly)) {
            error = QStringLiteral("cannot open local file: %1").arg(url.toString());
            return false;
        }
        out = f.readAll();
        return true;
    }

    QNetworkAccessManager manager;
    manager.setTransferTimeout(15000);
    QNetworkRequest req(url);
    QNetworkReply *reply = manager.get(req);
    QEventLoop loop;
    QObject::connect(reply, &QNetworkReply::finished, &loop, &QEventLoop::quit);
    loop.exec();

    if (reply->error() != QNetworkReply::NoError) {
        error = reply->errorString();
        reply->deleteLater();
        return false;
    }
    out = reply->readAll();
    reply->deleteLater();
    return true;
}

bool fetchWithRetry(const QUrl &url, QByteArray &out, QString &error)
{
    QString lastError;
    for (int attempt = 0; attempt < kDownloadRetries; ++attempt) {
        if (fetchUrl(url, out, lastError))
            return true;
    }
    error = lastError;
    return false;
}

QUrl objectUrl(const QUrl &base, const QString &hash)
{
    const QString rel = QStringLiteral("%1/%2").arg(hash.left(2), hash);
    QUrl u(base);
    QString path = u.path();
    if (!path.endsWith(QLatin1Char('/')))
        path += QLatin1Char('/');
    path += rel;
    u.setPath(path);
    return u;
}

} // namespace

ResolveAssets::ResolveAssets(LaunchTask *parent, MinecraftInstance *instance)
    : LaunchStep(parent), m_instance(instance)
{
}

void ResolveAssets::executeTask()
{
    if (!m_instance) {
        emit failed(QStringLiteral("no instance for asset resolution"));
        return;
    }
    const QString assetsDir = m_instance->assetsDir();
    if (!QDir().mkpath(assetsDir)) {
        emit failed(QStringLiteral("could not create asset directory: %1").arg(assetsDir));
        return;
    }
    const QString metadataPath = m_instance->instanceRoot() + QStringLiteral("/version.json");
    QFile metadataFile(metadataPath);
    if (!metadataFile.open(QIODevice::ReadOnly)) {
        emit failed(QStringLiteral("missing asset metadata: %1").arg(metadataPath));
        return;
    }

    QJsonParseError parseErr{};
    const QJsonDocument metadataDoc = QJsonDocument::fromJson(metadataFile.readAll(), &parseErr);
    if (parseErr.error != QJsonParseError::NoError || !metadataDoc.isObject()) {
        emit failed(QStringLiteral("invalid asset metadata JSON: %1").arg(metadataPath));
        return;
    }

    const QJsonObject assetIndexMeta = metadataDoc.object().value(QStringLiteral("asset_index")).toObject();
    const QString indexId = assetIndexMeta.value(QStringLiteral("id")).toString();
    const QString indexSha1 = assetIndexMeta.value(QStringLiteral("sha1")).toString();
    const QUrl indexUrl(assetIndexMeta.value(QStringLiteral("url")).toString());
    const QString baseUrlString =
        assetIndexMeta.value(QStringLiteral("objects_base_url"))
            .toString(QStringLiteral("https://resources.download.minecraft.net"));
    const QUrl objectsBase(baseUrlString);

    if (indexId.isEmpty() || !isSha1(indexSha1) || !indexUrl.isValid() || indexUrl.isEmpty()
        || !objectsBase.isValid() || objectsBase.isEmpty()) {
        emit failed(QStringLiteral("asset metadata missing required fields"));
        return;
    }

    const QString indexesDir = QDir(assetsDir).filePath(QStringLiteral("indexes"));
    const QString objectsDir = QDir(assetsDir).filePath(QStringLiteral("objects"));
    if (!QDir().mkpath(indexesDir) || !QDir().mkpath(objectsDir)) {
        emit failed(QStringLiteral("could not create asset subdirectories"));
        return;
    }

    QByteArray indexPayload;
    QString fetchError;
    if (!fetchWithRetry(indexUrl, indexPayload, fetchError)) {
        emit failed(QStringLiteral("asset index download failed: %1").arg(fetchError));
        return;
    }
    const QString downloadedIndexHash = QString::fromLatin1(
        QCryptographicHash::hash(indexPayload, QCryptographicHash::Sha1).toHex());
    if (downloadedIndexHash.compare(indexSha1, Qt::CaseInsensitive) != 0) {
        emit failed(QStringLiteral("asset index hash mismatch"));
        return;
    }

    const QString indexDst = QDir(indexesDir).filePath(indexId + QStringLiteral(".json"));
    QSaveFile outIndex(indexDst);
    if (!outIndex.open(QIODevice::WriteOnly)
        || outIndex.write(indexPayload) != indexPayload.size()
        || !outIndex.commit()) {
        emit failed(QStringLiteral("failed to write asset index: %1").arg(indexDst));
        return;
    }

    QJsonParseError indexParseErr{};
    const QJsonDocument indexDoc = QJsonDocument::fromJson(indexPayload, &indexParseErr);
    if (indexParseErr.error != QJsonParseError::NoError || !indexDoc.isObject()) {
        emit failed(QStringLiteral("asset index JSON invalid"));
        return;
    }
    const QJsonObject objects =
        indexDoc.object().value(QStringLiteral("objects")).toObject();

    int downloaded = 0;
    for (auto it = objects.begin(); it != objects.end(); ++it) {
        const QJsonObject obj = it.value().toObject();
        const QString hash = obj.value(QStringLiteral("hash")).toString();
        if (!isSha1(hash)) {
            emit failed(QStringLiteral("invalid asset hash for %1").arg(it.key()));
            return;
        }

        const QString rel = QStringLiteral("%1/%2").arg(hash.left(2), hash);
        const QString dst = QDir(objectsDir).filePath(rel);
        if (QFileInfo::exists(dst) && sha1File(dst).compare(hash, Qt::CaseInsensitive) == 0)
            continue;

        if (!QDir().mkpath(QFileInfo(dst).absolutePath())) {
            emit failed(QStringLiteral("could not create asset parent path"));
            return;
        }

        QByteArray payload;
        QString assetFetchError;
        if (!fetchWithRetry(objectUrl(objectsBase, hash), payload, assetFetchError)) {
            emit failed(QStringLiteral("asset download failed for %1: %2").arg(it.key(), assetFetchError));
            return;
        }
        const QString payloadHash = QString::fromLatin1(
            QCryptographicHash::hash(payload, QCryptographicHash::Sha1).toHex());
        if (payloadHash.compare(hash, Qt::CaseInsensitive) != 0) {
            emit failed(QStringLiteral("asset hash mismatch for %1").arg(it.key()));
            return;
        }

        QSaveFile out(dst);
        if (!out.open(QIODevice::WriteOnly)
            || out.write(payload) != payload.size()
            || !out.commit()) {
            emit failed(QStringLiteral("failed to write asset %1").arg(it.key()));
            return;
        }
        downloaded++;
    }

    emit logLine(QStringLiteral("Resolved assets in %1 (%2 downloaded)")
                     .arg(assetsDir)
                     .arg(downloaded));
    emit readyForLaunch();
}
