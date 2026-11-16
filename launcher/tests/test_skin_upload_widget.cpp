// SPDX-License-Identifier: GPL-3.0-only
#include <QTest>
#include <QFileDialog>
#include <QStandardPaths>

#include "gui/SkinUploadWidget.h"

class TestSkinUploadWidget : public QObject
{
    Q_OBJECT
private slots:
    void initTestCase()
    {
        QStandardPaths::setTestModeEnabled(true);
    }

    void testAccessToken()
    {
        SkinUploadWidget w;
        w.setAccessToken(QStringLiteral("tok"));
        QCOMPARE(w.accessToken(), QStringLiteral("tok"));
    }

    void testUploadSignalEmission()
    {
        SkinUploadWidget w;
        w.setVariant(QStringLiteral("default"));

        bool emitted = false;
        QString lastPath;
        QString lastVariant;
        QObject::connect(&w, &SkinUploadWidget::uploadRequested, [&](const QString &path, const QString &variant) {
            emitted = true;
            lastPath = path;
            lastVariant = variant;
        });

        QMetaObject::invokeMethod(&w, "uploadRequested", Q_ARG(QString, QString()),
                                  Q_ARG(QString, QStringLiteral("default")));
        QVERIFY(emitted);
        QCOMPARE(lastVariant, QStringLiteral("default"));
        QVERIFY(lastPath.isEmpty());
    }
};

QTEST_MAIN(TestSkinUploadWidget)
#include "test_skin_upload_widget.moc"
