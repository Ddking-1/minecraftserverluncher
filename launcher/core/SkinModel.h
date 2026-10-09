// SPDX-License-Identifier: GPL-3.0-only
#pragma once

#include <QJsonObject>
#include <QString>

namespace SkinModel {

// Minecraft skin model variant.
enum class SkinVariant {
    Classic,
    Slim,
};

// A reference to a validated skin asset. The skin belongs to the account; the
// launcher only carries an id, the asset hash and a URL (docs/SKIN_SYSTEM.md).
struct SkinRef {
    QString id;
    QString sha256;
    QString url;
    SkinVariant variant = SkinVariant::Classic;
};

[[nodiscard]]
QString variantToString(SkinVariant variant);

[[nodiscard]]
SkinVariant variantFromString(QStringView value);

[[nodiscard]]
bool isValidSkinId(QStringView id);

[[nodiscard]]
bool isValidSha256(QStringView hash);

[[nodiscard]]
QJsonObject toJson(const SkinRef &skin);

[[nodiscard]]
SkinRef fromJson(const QJsonObject &obj);

} // namespace SkinModel
