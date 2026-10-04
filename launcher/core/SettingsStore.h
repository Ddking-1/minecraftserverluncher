// SPDX-License-Identifier: GPL-3.0-only
#pragma once

#include <QString>
#include <QStringList>

namespace SettingsStore {

// Project-owned launcher settings persisted as launcher.json.
struct Settings
{
    QString javaPath;
    int minMemoryMb = 2048;
    int maxMemoryMb = 4096;
    QStringList jvmArgs;
};

// Documented defaults.
Settings defaults();

// Fill unset fields with defaults.
void applyDefaults(Settings &s);

// Enforce memory bounds: min >= 512, max >= min, max <= 65536.
void clamp(Settings &s);

// A JVM argument is safe when it is non-empty, at most 1024 chars, equals its
// trimmed form and contains no control characters.
bool isSafeJvmArg(const QString &arg);

// Validate memory bounds and JVM arguments. Returns false and sets *error
// (when non-null) on the first problem.
bool validate(const Settings &s, QString *error = nullptr);

// Read settings from a JSON file. Returns false if the file is missing or
// malformed. Loaded values are defaulted and clamped.
bool load(const QString &path, Settings &out);

// Atomically write settings as JSON. Returns false on failure.
bool save(const QString &path, const Settings &in);

} // namespace SettingsStore
