// SPDX-License-Identifier: GPL-3.0-only
#include "SplashScreen.h"

#include <QPainter>

SplashScreen::SplashScreen(const QString &imagePath, QWidget *parent)
    : QWidget(parent), m_imagePath(imagePath)
{
    setWindowFlags(Qt::SplashScreen | Qt::FramelessWindowHint);
    setAttribute(Qt::WA_TranslucentBackground, false);
}

bool SplashScreen::load()
{
    if (!m_pixmap.load(m_imagePath))
        return false;
    resize(m_pixmap.size());
    return true;
}

const QPixmap &SplashScreen::pixmap() const
{
    return m_pixmap;
}

void SplashScreen::paintEvent(QPaintEvent *event)
{
    Q_UNUSED(event);
    QPainter painter(this);
    if (!m_pixmap.isNull())
        painter.drawPixmap(0, 0, m_pixmap);
    else
        painter.fillRect(rect(), Qt::black);
}
