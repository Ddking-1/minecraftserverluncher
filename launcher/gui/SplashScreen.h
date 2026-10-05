// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher UI. Instant "fake render" splash screen.
#pragma once

#include <QWidget>
#include <QPixmap>
#include <QString>

// SplashScreen shows a pre-rendered image immediately on startup so weak
// machines do not sit on a blank/black window while heavier subsystems load.
// It performs no rendering work itself ("fake render"): it simply paints a
// bundled image and reports when it has been shown.
class SplashScreen : public QWidget
{
    Q_OBJECT
public:
    explicit SplashScreen(const QString &imagePath, QWidget *parent = nullptr);

    // Load the image at imagePath. Returns false if it could not be loaded.
    bool load();

    // Backing image accessor (mainly for tests).
    const QPixmap &pixmap() const;

protected:
    void paintEvent(QPaintEvent *event) override;

private:
    QString m_imagePath;
    QPixmap m_pixmap;
};
