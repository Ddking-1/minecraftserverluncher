#pragma once

#include <QLabel>
#include <QMainWindow>
#include <QVBoxLayout>
#include <QWidget>

#include "ui/SkinPreviewWidget.h"

class BackendAuthClient;

class MainWindow : public QMainWindow
{
    Q_OBJECT

public:
    explicit MainWindow(QWidget *parent = nullptr);

    // Supplies the authenticated backend client to the embedded skin panel.
    void setAuthClient(BackendAuthClient *client);

private slots:
    void onSkinUploadFinished(bool success, const QString &skinId, const QString &error);

private:
    SkinPreviewWidget *m_skinPreview = nullptr;
    QLabel *m_statusLabel = nullptr;
};
