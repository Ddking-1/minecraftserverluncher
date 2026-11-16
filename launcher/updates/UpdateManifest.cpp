// SPDX-License-Identifier: GPL-3.0-only
#include "UpdateManifest.h"

#include <QFile>
#include <QJsonArray>
#include <QJsonDocument>
#include <QJsonValue>
#include <QRegularExpression>

namespace UpdateManifest {

namespace {

// A manifest file path must be relative and must not traverse upwards. This
// blocks path-traversal attacks described in docs/UPDATE_SPEC.md.
[[nodiscard]]
bool isSafePath(QStringView path)
{
    if (path.isEmpty())
        return false;
    if (path.startsWith(QLatin1Char('/')) || path.startsWith(QLatin1Char('\\')))
        return false;
    if (path.contains(QLatin1Char(':'))) // drive letter or URI scheme
        return false;
    const auto parts = path.split(QRegularExpression(QStringLiteral("[/\\\\]")), Qt::SkipEmptyParts);
    for (const auto &part : parts) {
        if (part.compare(QLatin1String(".."), Qt::CaseInsensitive) == 0)
            return false;
    }
    return true;
}

[[nodiscard]]
ParseResult fail(QStringView message)
{
    ParseResult result;
    result.ok = false;
    result.error = message.toString();
    return result;
}

} // namespace

[[nodiscard]]
ParseResult parse(const QJsonObject &obj)
{
    const QString name = obj.value(QStringLiteral("name")).toString();
    const QString version = obj.value(QStringLiteral("version")).toString();
    if (name.isEmpty())
        return fail(u"manifest: missing name");
    if (version.isEmpty())
        return fail(u"manifest: missing version");

    const QJsonValue filesValue = obj.value(QStringLiteral("files"));
    if (!filesValue.isArray())
        return fail(u"manifest: files must be an array");

    ManifestComponent component;
    component.name = name;
    component.version = version;
    component.minLauncher = obj.value(QStringLiteral("min_launcher")).toString();
    component.signature = obj.value(QStringLiteral("signature")).toString();

    const QJsonArray deps = obj.value(QStringLiteral("dependencies")).toArray();
    for (const QJsonValue &d : deps)
        component.dependencies.append(d.toString());

    const QJsonArray files = filesValue.toArray();
    for (const QJsonValue &value : files) {
        if (!value.isObject())
            return fail(u"manifest: file entry must be an object");
        const QJsonObject fileObj = value.toObject();

        ManifestFile file;
        file.path = fileObj.value(QStringLiteral("path")).toString();
        const QJsonValue sizeValue = fileObj.value(QStringLiteral("size"));
        if (sizeValue.type() == QJsonValue::Double) {
            file.size = static_cast<qint64>(sizeValue.toDouble());
        } else if (sizeValue.type() == QJsonValue::String) {
            bool ok = false;
            const qint64 size = sizeValue.toString().toLongLong(&ok);
            if (ok)
                file.size = size;
        }
        file.sha256 = fileObj.value(QStringLiteral("sha256")).toString();
        file.url = fileObj.value(QStringLiteral("url")).toString();

        if (!isSafePath(file.path))
            return fail(QStringLiteral("manifest: unsafe file path '%1'").arg(file.path));
        if (file.sha256.isEmpty())
            return fail(QStringLiteral("manifest: missing sha256 for '%1'").arg(file.path));
        if (file.sha256.size() != 64)
            return fail(QStringLiteral("manifest: sha256 must be 64 hex chars for '%1'").arg(file.path));
        if (file.url.isEmpty())
            return fail(QStringLiteral("manifest: missing url for '%1'").arg(file.path));

        component.files.append(file);
    }

    ParseResult result;
    result.ok = true;
    result.component = component;
    return result;
}

ParseResult parseFile(QStringView path)
{
    QFile file{path.toString()};
    if (!file.open(QIODevice::ReadOnly))
        return fail(QStringLiteral("manifest: cannot open '%1'").arg(path));

    const QByteArray data = file.readAll();
    file.close();

    QJsonParseError err{};
    const QJsonDocument doc = QJsonDocument::fromJson(data, &err);
    if (err.error != QJsonParseError::NoError || !doc.isObject())
        return fail(QStringLiteral("manifest: invalid JSON in '%1'").arg(path));

    return parse(doc.object());
}

} // namespace UpdateManifest
