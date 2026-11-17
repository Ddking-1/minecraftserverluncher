// SPDX-License-Identifier: GPL-3.0-only
#include "ResolveLibraries.h"

#include <QCryptographicHash>
#include <QDir>
#include <QEventLoop>
#include <QFile>
#include <QFileInfo>
#include <QJsonArray>
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

bool isSha1(const QString &hash)
{
    static const QRegularExpression re(QStringLiteral("^[A-Fa-f0-9]{40}$"));
    return re.match(hash).hasMatch();
}

bool isSafeRelativePath(const QString &path)
{
    if (path.isEmpty())
        return false;
    if (QDir::isAbsolutePath(path))
        return false;
    if (path.contains(QLatin1Char(':')))
        return false;
    const QStringList parts = path.split(QRegularExpression(QStringLiteral("[/\\\\]")),
                                         Qt::SkipEmptyParts);
    for (const QString &part : parts) {
        if (part == QLatin1String(".."))
            return false;
    }
    return !parts.isEmpty();
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

bool pathInside(const QString &baseDir, const QString &targetPath)
{
    const QDir base(baseDir);
    const QString rel = base.relativeFilePath(QDir::cleanPath(targetPath));
    if (rel.isEmpty() || rel == QLatin1String("."))
        return true;
    return !rel.startsWith(QStringLiteral("../")) && rel != QStringLiteral("..");
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

} // namespace

ResolveLibraries::ResolveLibraries(LaunchTask *parent, MinecraftInstance *instance)
    : LaunchStep(parent), m_instance(instance)
{
}

void ResolveLibraries::executeTask()
{
    if (!m_instance) {
        emit failed(QStringLiteral("no instance for library resolution"));
        return;
    }

    const QString libDir = m_instance->libDir();
    if (!QDir().mkpath(libDir)) {
        emit failed(QStringLiteral("could not create library directory: %1").arg(libDir));
        return;
    }

    const QString metadataPath = m_instance->instanceRoot() + QStringLiteral("/version.json");
    QFile metadataFile(metadataPath);
    if (!metadataFile.open(QIODevice::ReadOnly)) {
        emit failed(QStringLiteral("missing library metadata: %1").arg(metadataPath));
        return;
    }

    QJsonParseError parseErr{};
    const QJsonDocument metadataDoc = QJsonDocument::fromJson(metadataFile.readAll(), &parseErr);
    if (parseErr.error != QJsonParseError::NoError || !metadataDoc.isObject()) {
        emit failed(QStringLiteral("invalid library metadata JSON: %1").arg(metadataPath));
        return;
    }
    const QJsonArray libraries = metadataDoc.object().value(QStringLiteral("libraries")).toArray();
    if (libraries.isEmpty()) {
        emit logLine(QStringLiteral("No libraries declared in metadata."));
        emit readyForLaunch();
        return;
    }

    int downloaded = 0;
    for (const QJsonValue &value : libraries) {
        if (!value.isObject()) {
            emit failed(QStringLiteral("library metadata entry must be an object"));
            return;
        }
        const QJsonObject lib = value.toObject();
        const QString relPath = lib.value(QStringLiteral("path")).toString();
        const QString sha1 = lib.value(QStringLiteral("sha1")).toString();
        const QUrl url(lib.value(QStringLiteral("url")).toString());

        if (!isSafeRelativePath(relPath)) {
            emit failed(QStringLiteral("unsafe library path: %1").arg(relPath));
            return;
        }
        if (!isSha1(sha1)) {
            emit failed(QStringLiteral("invalid library sha1 for %1").arg(relPath));
            return;
        }
        if (!url.isValid() || url.isEmpty()) {
            emit failed(QStringLiteral("invalid library url for %1").arg(relPath));
            return;
        }

        const QString dstPath = QDir(libDir).filePath(relPath);
        if (!pathInside(libDir, dstPath)) {
            emit failed(QStringLiteral("library target escapes library dir: %1").arg(relPath));
            return;
        }
        if (QFileInfo::exists(dstPath) && sha1File(dstPath).compare(sha1, Qt::CaseInsensitive) == 0)
            continue;

        if (!QDir().mkpath(QFileInfo(dstPath).absolutePath())) {
            emit failed(QStringLiteral("could not create library parent path: %1").arg(dstPath));
            return;
        }

        QByteArray payload;
        QString fetchError;
        if (!fetchUrl(url, payload, fetchError)) {
            emit failed(QStringLiteral("library download failed for %1: %2").arg(relPath, fetchError));
            return;
        }
        const QString payloadHash = QString::fromLatin1(
            QCryptographicHash::hash(payload, QCryptographicHash::Sha1).toHex());
        if (payloadHash.compare(sha1, Qt::CaseInsensitive) != 0) {
            emit failed(QStringLiteral("library hash mismatch for %1").arg(relPath));
            return;
        }

        QSaveFile out(dstPath);
        if (!out.open(QIODevice::WriteOnly)
            || out.write(payload) != payload.size()
            || !out.commit()) {
            emit failed(QStringLiteral("failed to write library: %1").arg(dstPath));
            return;
        }
        downloaded++;
    }

    emit logLine(QStringLiteral("Resolved libraries in %1 (%2 downloaded)")
                     .arg(libDir)
                     .arg(downloaded));
    emit readyForLaunch();
}
