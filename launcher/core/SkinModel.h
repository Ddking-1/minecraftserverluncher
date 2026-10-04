// SPDX-License-Identifier: GPL-3.0-only
#pragma once

#include <QJsonObject>
#include <QString>

namespace SkinModel {

// Minecraft skin model variant.
enum class SkinVariant
{
    Classic,
    Slim,
};

// A reference to a validated skin asset. The skin belongs to the account; the
// launcher only carries an id, the asset hash and a URL (docs/SKIN_SYSTEM.md).
struct SkinRef
{
    QString id;
    QString sha256;
    QString url;
    SkinVariant variant = SkinVariant::Classic;
};

QString variantToString(SkinVariant variant);
SkinVariant variantFromString(const QString &value);

// A skin id is a lowercase alphanumeric/underscore/dash token, 1-64 chars.
bool isValidSkinId(const QString &id);
// A SHA-256 digest is exactly 64 hexadecimal characters.
bool isValidSha256(const QString &hash);

QJsonObject toJson(const SkinRef &skin);
SkinRef fromJson(const QJsonObject &obj);

} // namespace SkinModel
