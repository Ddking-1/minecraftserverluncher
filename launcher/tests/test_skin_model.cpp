// SPDX-License-Identifier: GPL-3.0-only
// Qt Test-based verification for the skin model.
#include <QtTest>

#include "core/SkinModel.h"

class TestSkinModel : public QObject
{
    Q_OBJECT

private slots:
    void variantToString();
    void variantFromString();
    void isValidSkinId();
    void isValidSha256();
    void roundTrip();
};

namespace {
const char *const kHash =
    "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855";
} // namespace

void TestSkinModel::variantToString()
{
    QCOMPARE(SkinModel::variantToString(SkinModel::SkinVariant::Classic), QStringLiteral("classic"));
    QCOMPARE(SkinModel::variantToString(SkinModel::SkinVariant::Slim), QStringLiteral("slim"));
}

void TestSkinModel::variantFromString()
{
    QCOMPARE(SkinModel::variantFromString(QStringLiteral("slim")), SkinModel::SkinVariant::Slim);
    QCOMPARE(SkinModel::variantFromString(QStringLiteral("bogus")), SkinModel::SkinVariant::Classic);
}

void TestSkinModel::isValidSkinId()
{
    QVERIFY(SkinModel::isValidSkinId(QStringLiteral("steve-01")));
    QVERIFY(!SkinModel::isValidSkinId(QStringLiteral("Bad ID")));
    QVERIFY(!SkinModel::isValidSkinId(QStringLiteral("")));
    QVERIFY(!SkinModel::isValidSkinId(QStringLiteral("A")));
}

void TestSkinModel::isValidSha256()
{
    QVERIFY(SkinModel::isValidSha256(QString::fromLatin1(kHash)));
    QVERIFY(!SkinModel::isValidSha256(QStringLiteral("deadbeef")));
    QVERIFY(!SkinModel::isValidSha256(QString(64, QLatin1Char('z'))));
}

void TestSkinModel::roundTrip()
{
    SkinModel::SkinRef skin;
    skin.id = QStringLiteral("steve-01");
    skin.sha256 = QString::fromLatin1(kHash);
    skin.url = QStringLiteral("https://example.com/skin.png");
    skin.variant = SkinModel::SkinVariant::Slim;

    const SkinModel::SkinRef roundTrip = SkinModel::fromJson(SkinModel::toJson(skin));
    QCOMPARE(roundTrip.id, skin.id);
    QCOMPARE(roundTrip.sha256, skin.sha256);
    QCOMPARE(roundTrip.url, skin.url);
    QCOMPARE(roundTrip.variant, skin.variant);
}

QTEST_APPLESS_MAIN(TestSkinModel)
#include "test_skin_model.moc"
