#ifndef SKIN_PREVIEW_WIDGET_H
#define SKIN_PREVIEW_WIDGET_H

#include <QLabel>
#include <QPushButton>
#include <QWidget>

class SkinPreviewWidget : public QWidget {
    Q_OBJECT
public:
    explicit SkinPreviewWidget(QWidget *parent = nullptr);

private slots:
    void onUploadClicked();

signals:
    void skinUploadFinished(bool success, const QString &skinId, const QString &error);

private:
    QLabel   *m_avatarLabel;
    QPushButton *m_uploadButton;
    QLabel   *m_statusLabel;
};

#endif // SKIN_PREVIEW_WIDGET_H
