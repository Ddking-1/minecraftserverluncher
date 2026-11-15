// SPDX-License-Identifier: GPL-3.0-only
#include "SkinPreviewWidget.h"

#include "auth/BackendAuthClient.h"
#include "core/SkinClient.h"

#include <QDebug>
#include <QFileDialog>
#include <QPixmap>
#include <QVBoxLayout>

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

void SkinPreviewWidget::setAuthClient(BackendAuthClient *client)
{
    m_authClient = client;
}

void SkinPreviewWidget::onUploadClicked()
{
    if (m_authClient == nullptr || m_authClient->accessToken().isEmpty()) {
        const QString error = tr("Bejelentkezés szükséges a skin feltöltéséhez.");
        m_statusLabel->setText(error);
        emit skinUploadFinished(false, QString(), error);
        return;
    }

    QString fileName =
        QFileDialog::getOpenFileName(this, QStringLiteral("Skin kiválasztása"),
                                     QString(), QStringLiteral("PNG Fájlok (*.png)"));

    if (fileName.isEmpty())
        return;

    auto *client = new SkinClient(m_authClient, this);
    client->uploadSkin(fileName, [this, client](bool success, const QString &skinId, const QString &error) {
        if (!success) {
            m_statusLabel->setText(error.isEmpty()
                                       ? tr("Feltöltés sikertelen.")
                                       : tr("Feltöltés sikertelen: %1").arg(error));
            emit skinUploadFinished(false, QString(), error);
            client->deleteLater();
            return;
        }

        client->downloadAvatar(skinId, [this, client](bool success, const QPixmap &avatar) {
            client->deleteLater();
            if (!success || avatar.isNull()) {
                m_statusLabel->setText(tr("Avatár betöltése sikertelen."));
                emit skinUploadFinished(false, skinId, tr("Avatár betöltése sikertelen."));
                return;
            }

            m_avatarLabel->setPixmap(avatar);
            m_statusLabel->setText(tr("Skin sikeresen feltöltve!"));
            emit skinUploadFinished(true, skinId, QString());
        });
    });
}
