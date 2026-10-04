// SPDX-License-Identifier: GPL-3.0-only
#pragma once

#include <QByteArray>
#include <QString>

#include "core/SkinModel.h"

namespace SkinCache {

// Content-addressed file name derived from the skin id and hash.
QString cacheFileName(const SkinModel::SkinRef &skin);

// Absolute path of the cached skin file inside cacheDir.
QString cachedPath(const QString &cacheDir, const SkinModel::SkinRef &skin);

// Store skin data if its SHA-256 matches skin.sha256. Creates cacheDir if
// needed. On success writes *outPath (when non-null) and returns true.
bool store(const QString &cacheDir, const SkinModel::SkinRef &skin,
           const QByteArray &data, QString *outPath = nullptr);

// True when a cached file exists AND its SHA-256 equals skin.sha256.
bool contains(const QString &cacheDir, const SkinModel::SkinRef &skin);

} // namespace SkinCache
