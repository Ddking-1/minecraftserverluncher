#ifndef SKIN_PREVIEW_WIDGET_H
#define SKIN_PREVIEW_WIDGET_H

#include <QLabel>
#include <QPushButton>
#include <QWidget>

class BackendAuthClient;

class SkinPreviewWidget : public QWidget {
    Q_OBJECT
public:
    explicit SkinPreviewWidget(QWidget *parent = nullptr);

    // The client used for upload/download requests. Without it the widget
    // reports that the user is not logged in instead of crashing.
    void setAuthClient(BackendAuthClient *client);

private slots:
    void onUploadClicked();

signals:
    void skinUploadFinished(bool success, const QString &skinId, const QString &error);

private:
    QLabel   *m_avatarLabel;
    QPushButton *m_uploadButton;
    QLabel   *m_statusLabel;
    BackendAuthClient *m_authClient = nullptr;
};

#endif // SKIN_PREVIEW_WIDGET_H
