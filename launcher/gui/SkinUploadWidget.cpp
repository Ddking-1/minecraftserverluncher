// SPDX-License-Identifier: GPL-3.0-only
#include "SkinUploadWidget.h"

#include <QFileDialog>
#include <QHBoxLayout>
#include <QPixmap>
#include <QVBoxLayout>

SkinUploadWidget::SkinUploadWidget(QWidget *parent)
    : QWidget(parent)
{
    buildUi();
    m_variant = QStringLiteral("default");
    m_selectedFile.clear();
}

void SkinUploadWidget::buildUi()
{
    QVBoxLayout *layout = new QVBoxLayout(this);

    m_preview = new QLabel(QStringLiteral("No skin selected"), this);
    m_preview->setMinimumSize(64, 64);
    layout->addWidget(m_preview);

    QHBoxLayout *hbox = new QHBoxLayout;
    m_choose = new QPushButton(QStringLiteral("Choose..."), this);
    m_upload = new QPushButton(QStringLiteral("Upload"), this);
    hbox->addWidget(m_choose);
    hbox->addWidget(m_upload);

    connect(m_choose, &QPushButton::clicked, this, &SkinUploadWidget::chooseFile);
    connect(m_upload, &QPushButton::clicked, this, [this]() {
        if (m_selectedFile.isEmpty())
            emit uploadRequested(QString(), m_variant);
        else
            emit uploadRequested(m_selectedFile, m_variant);
    });

    layout->addLayout(hbox);
    setLayout(layout);
}

void SkinUploadWidget::chooseFile()
{
    const QString filePath = QFileDialog::getOpenFileName(this, QStringLiteral("Select skin"),
                                                           QString(), QStringLiteral("PNG (*.png)"));
    if (filePath.isEmpty())
        return;
    QPixmap pixmap(filePath);
    if (!pixmap.isNull()) {
        m_preview->setPixmap(pixmap.scaled(64, 64, Qt::KeepAspectRatioByExpanding, Qt::SmoothTransformation));
    }
    m_selectedFile = filePath;
}

void SkinUploadWidget::setAccessToken(const QString &token)
{
    m_accessToken = token;
}

QString SkinUploadWidget::accessToken() const
{
    return m_accessToken;
}

void SkinUploadWidget::setVariant(const QString &variant)
{
    m_variant = variant;
}
