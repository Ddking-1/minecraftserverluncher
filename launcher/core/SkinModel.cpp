// SPDX-License-Identifier: GPL-3.0-only
#include "SkinModel.h"

#include <QRegularExpression>

namespace SkinModel {

QString variantToString(SkinVariant variant)
{
    return variant == SkinVariant::Slim ? QStringLiteral("slim") : QStringLiteral("classic");
}

SkinVariant variantFromString(QStringView value)
{
    return value == QLatin1String("slim") ? SkinVariant::Slim : SkinVariant::Classic;
}

bool isValidSkinId(QStringView id)
{
    if (id.isEmpty() || id.size() > 64)
        return false;
    static const QRegularExpression re(QStringLiteral("^[a-z0-9_-]+$"));
    return re.match(id.toString()).hasMatch();
}

bool isValidSha256(QStringView hash)
{
    if (hash.size() != 64)
        return false;
    static const QRegularExpression re(QStringLiteral("^[0-9a-fA-F]{64}$"));
    return re.match(hash.toString()).hasMatch();
}

QJsonObject toJson(const SkinRef &skin)
{
    QJsonObject obj;
    obj.insert(QStringLiteral("id"), skin.id);
    obj.insert(QStringLiteral("sha256"), skin.sha256);
    obj.insert(QStringLiteral("url"), skin.url);
    obj.insert(QStringLiteral("variant"), variantToString(skin.variant));
    return obj;
}

SkinRef fromJson(const QJsonObject &obj)
{
    SkinRef skin;
    skin.id = obj.value(QStringLiteral("id")).toString();
    skin.sha256 = obj.value(QStringLiteral("sha256")).toString();
    skin.url = obj.value(QStringLiteral("url")).toString();
    skin.variant = variantFromString(obj.value(QStringLiteral("variant")).toString());
    return skin;
}

} // namespace SkinModel
