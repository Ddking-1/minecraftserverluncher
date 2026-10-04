// SPDX-License-Identifier: GPL-3.0-only
// Standalone verification for the skin model.
#include <iostream>

#include "core/SkinModel.h"

namespace {

const char *const kHash =
    "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855";

bool check(bool cond, const char *what)
{
    if (!cond)
        std::cerr << "FAIL: " << what << std::endl;
    return cond;
}

} // namespace

int main()
{
    using namespace SkinModel;

    if (!check(variantToString(SkinVariant::Classic) == QStringLiteral("classic"), "classic string")
        || !check(variantToString(SkinVariant::Slim) == QStringLiteral("slim"), "slim string")
        || !check(variantFromString(QStringLiteral("slim")) == SkinVariant::Slim, "slim parse")
        || !check(variantFromString(QStringLiteral("bogus")) == SkinVariant::Classic, "unknown parse"))
        return 1;

    if (!check(isValidSkinId(QStringLiteral("steve-01")), "valid id")
        || !check(!isValidSkinId(QStringLiteral("Bad ID")), "space rejected")
        || !check(!isValidSkinId(QStringLiteral("")), "empty rejected")
        || !check(!isValidSkinId(QStringLiteral("A")) , "uppercase rejected"))
        return 1;

    if (!check(isValidSha256(QString::fromLatin1(kHash)), "valid hash")
        || !check(!isValidSha256(QStringLiteral("deadbeef")), "short hash rejected")
        || !check(!isValidSha256(QString(64, QLatin1Char('z'))), "non-hex rejected"))
        return 1;

    SkinRef skin;
    skin.id = QStringLiteral("steve-01");
    skin.sha256 = QString::fromLatin1(kHash);
    skin.url = QStringLiteral("https://example.com/skin.png");
    skin.variant = SkinVariant::Slim;

    const SkinRef roundTrip = fromJson(toJson(skin));
    if (!check(roundTrip.id == skin.id, "round-trip id")
        || !check(roundTrip.sha256 == skin.sha256, "round-trip sha256")
        || !check(roundTrip.url == skin.url, "round-trip url")
        || !check(roundTrip.variant == skin.variant, "round-trip variant"))
        return 1;

    std::cout << "OK" << std::endl;
    return 0;
}
