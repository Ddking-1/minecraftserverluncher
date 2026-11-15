#include "MainWindow.h"

#include "auth/BackendAuthClient.h"

#include <QStatusBar>
#include <QVBoxLayout>

MainWindow::MainWindow(QWidget *parent)
    : QMainWindow(parent)
{
    m_skinPreview = new SkinPreviewWidget(this);

    QWidget *central = new QWidget(this);
    QVBoxLayout *layout = new QVBoxLayout(central);
    layout->addWidget(m_skinPreview);
    central->setLayout(layout);
    setCentralWidget(central);

    m_statusLabel = new QLabel(tr("Kész"), this);
    m_statusLabel->setWordWrap(true);
    statusBar()->addPermanentWidget(m_statusLabel);

    connect(m_skinPreview, &SkinPreviewWidget::skinUploadFinished,
            this, &MainWindow::onSkinUploadFinished);
}

void MainWindow::setAuthClient(BackendAuthClient *client)
{
    m_skinPreview->setAuthClient(client);
}

void MainWindow::onSkinUploadFinished(bool success, const QString &skinId, const QString &error)
{
    if (success) {
        m_statusLabel->setText(tr("Skin sikeresen frissítve (ID: %1)").arg(skinId));
        statusBar()->showMessage(tr("Skin feltöltés sikeres"), 5000);
        return;
    }

    m_statusLabel->setText(error.isEmpty()
                               ? tr("Skin feltöltés sikertelen.")
                               : tr("Skin feltöltés sikertelen: %1").arg(error));
    statusBar()->showMessage(tr("Skin feltöltés sikertelen"), 5000);
}
