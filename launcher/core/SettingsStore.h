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

// Read settings from a JSON file. Returns false if the file is missing or
// malformed. Loaded values are defaulted and clamped.
bool load(const QString &path, Settings &out);

// Atomically write settings as JSON. Returns false on failure.
bool save(const QString &path, const Settings &in);

} // namespace SettingsStore
