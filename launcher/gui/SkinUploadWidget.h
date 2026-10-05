// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher UI. Widget for previewing and uploading a skin file.
#pragma once

#include <QLabel>
#include <QPushButton>
#include <QString>
#include <QWidget>

class SkinUploadWidget : public QWidget
{
    Q_OBJECT
public:
    explicit SkinUploadWidget(QWidget *parent = nullptr);

    void setAccessToken(const QString &token);
    QString accessToken() const;
    void setVariant(const QString &variant);

signals:
    // Emitted when the user clicks Upload. filePath is empty if no file was
    // chosen.
    void uploadRequested(const QString &filePath, const QString &variant);

private:
    QString m_accessToken;
    QString m_variant;
    QString m_selectedFile;
    QLabel *m_preview = nullptr;
    QPushButton *m_choose = nullptr;
    QPushButton *m_upload = nullptr;

    void buildUi();
    void chooseFile();
};
