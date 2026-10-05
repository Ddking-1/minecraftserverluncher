// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher core. Local-only UI editor configuration.
#pragma once

#include <QString>
#include <QJsonObject>

// EditorConfig is the output of the launcher's local editor mode. It is a
// purely local UI layout/theme description. It never talks to the server; the
// resulting file is handed to the main admin manually and pushed as an update
// through the secure backend endpoint.
struct EditorConfig
{
    QString title;
    QString backgroundImage;
    QString accentColor;
    bool showNewsTab = true;
    bool showSettingsTab = true;

    // Sensible defaults so the editor always produces a valid file.
    static EditorConfig defaults();

    QJsonObject toJson() const;
    static EditorConfig fromJson(const QJsonObject &obj);

    // Read/write the config to a local file (JSON). Returns false on error.
    bool save(const QString &path) const;
    static bool load(const QString &path, EditorConfig &out);
};
