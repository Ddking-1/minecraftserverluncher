// SPDX-License-Identifier: GPL-3.0-only
#include "SkinPreviewWidget.h"

#include "SkinPreviewWidget.h"

#include "core/SkinClient.h"

#include <QFileDialog>
#include <QPixmap>
#include <QVBoxLayout>

#include <QDebug>

SkinPreviewWidget::SkinPreviewWidget(QWidget *parent)
    : QWidget(parent)
{
    auto *layout = new QVBoxLayout(this);

    m_avatarLabel = new QLabel(this);
    m_avatarLabel->setFixedSize(64, 64);
    m_avatarLabel->setAlignment(Qt::AlignCenter);
    layout->addWidget(m_avatarLabel);

    m_uploadButton = new QPushButton(QStringLiteral("Skin feltöltése"), this);
    layout->addWidget(m_uploadButton);

    m_statusLabel = new QLabel(this);
    m_statusLabel->setWordWrap(true);
    layout->addWidget(m_statusLabel);

    connect(m_uploadButton, &QPushButton::clicked, this, &SkinPreviewWidget::onUploadClicked);
    setLayout(layout);
}

void SkinPreviewWidget::onUploadClicked()
{
    QString fileName =
        QFileDialog::getOpenFileName(this, QStringLiteral("Skin kiválasztása"),
                                     QString(), QStringLiteral("PNG Fájlok (*.png)"));

    if (fileName.isEmpty())
        return;

    SkinClient *client = new SkinClient(nullptr, this);
    client->uploadSkin(fileName, [this, client](bool success, const QString &skinId, const QString &error) {
        if (!success) {
            m_statusLabel->setText(error.isEmpty()
                                       ? tr("Feltöltés sikertelen.")
                                       : tr("Feltöltés sikertelen: %1").arg(error));
            return;
        }

        client->downloadAvatar(skinId, [this, client](bool success, const QPixmap &avatar) {
            if (!success || avatar.isNull()) {
                m_statusLabel->setText(tr("Ábras betöltése sikertelen."));
                return;
            }

            m_avatarLabel->setPixmap(avatar);
            m_statusLabel->setText(tr("Skin sikeresen feltöltve!"));
        });
    });
}
