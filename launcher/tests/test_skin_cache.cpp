// SPDX-License-Identifier: GPL-3.0-only
// Qt Test-based verification for the content-addressed skin cache.
#include <QtTest>

#include <QCryptographicHash>
#include <QFile>
#include <QTemporaryDir>

#include "core/SkinCache.h"

class TestSkinCache : public QTest
{
    Q_OBJECT

private slots:
    void storeAndContains();
    void mismatchedHash();
    void storeRejectsTampered();
};

namespace {
SkinModel::SkinRef makeSkin(const QString &id, const QByteArray &data)
{
    SkinModel::SkinRef skin;
    skin.id = id;
    skin.sha256 = QString::fromLatin1(
        QCryptographicHash::hash(data, QCryptographicHash::Sha256).toHex());
    skin.url = QStringLiteral("https://example.com/skin.png");
    return skin;
}
} // namespace

void TestSkinCache::storeAndContains()
{
    QTemporaryDir dir;
    QVERIFY(dir.isValid());
    const QString cacheDir = dir.filePath(QStringLiteral("skincache"));

    const QByteArray data("fake-png-bytes");
    const SkinModel::SkinRef skin = makeSkin(QStringLiteral("steve-01"), data);

    QString outPath;
    QVERIFY(SkinCache::store(cacheDir, skin, data, &outPath));
    QVERIFY(QFile::exists(outPath));
    QVERIFY(SkinCache::contains(cacheDir, skin));
}

void TestSkinCache::mismatchedHash()
{
    QTemporaryDir dir;
    QVERIFY(dir.isValid());
    const QString cacheDir = dir.filePath(QStringLiteral("skincache"));

    const QByteArray data("fake-png-bytes");
    const SkinModel::SkinRef skin = makeSkin(QStringLiteral("steve-01"), data);

    QString outPath;
    QVERIFY(SkinCache::store(cacheDir, skin, data, &outPath));

    QVERIFY(!SkinCache::contains(cacheDir, makeSkin(QStringLiteral("steve-01"), QByteArray("other"))));
}

void TestSkinCache::storeRejectsTampered()
{
    QTemporaryDir dir;
    QVERIFY(dir.isValid());
    const QString cacheDir = dir.filePath(QStringLiteral("skincache"));

    const QByteArray data("fake-png-bytes");
    const SkinModel::SkinRef skin = makeSkin(QStringLiteral("steve-01"), data);

    QVERIFY(!SkinCache::store(cacheDir, skin, QByteArray("tampered")));
}

QTEST_APPLESS_MAIN(TestSkinCache)
#include "test_skin_cache.moc"
