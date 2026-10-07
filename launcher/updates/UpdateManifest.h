// SPDX-License-Identifier: GPL-3.0-only
#pragma once

#include <QJsonObject>
#include <QList>
#include <QString>
#include <QStringList>

namespace UpdateManifest {

// A single file entry in a component manifest.
struct ManifestFile
{
    QString path;
    qint64 size = 0;
    QString sha256;
    QString url;
};

// A versioned, independently updatable component (launcher, Fabric, modpack...).
struct ManifestComponent
{
    QString name;
    QString version;
    QList<ManifestFile> files;
    QStringList dependencies;
    QString minLauncher;
    QString signature;
};

// Result of parsing a manifest. On failure `error` explains why and
// `component` is empty.
struct ParseResult
{
    bool ok = false;
    QString error;
    ManifestComponent component;
};

// Parses a manifest object. Rejects missing required fields, malformed file
// entries and unsafe (path-traversing / absolute) file paths.
[[nodiscard]]
ParseResult parse(const QJsonObject &obj);

// Reads and parses a manifest from a JSON file.
[[nodiscard]]
ParseResult parseFile(const QString &path);

} // namespace UpdateManifest
