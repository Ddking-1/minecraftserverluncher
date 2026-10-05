// SPDX-License-Identifier: GPL-3.0-only
#include "EditorSettingsWidget.h"

void EditorSettingsWidget::buildUi()
{
    QVBoxLayout *layout = new QVBoxLayout(this);

    m_title = new QLineEdit;
    m_title->setPlaceholderText(QStringLiteral("Title"));
    layout->addWidget(m_title);

    m_accent = new QLineEdit;
    m_accent->setPlaceholderText(QStringLiteral("Accent color"));
    layout->addWidget(m_accent);

    m_background = new QLineEdit;
    m_background->setPlaceholderText(QStringLiteral("Background image"));
    layout->addWidget(m_background);

    m_showNews = new QCheckBox(QStringLiteral("Show news tab"));
    layout->addWidget(m_showNews);

    m_showSettings = new QCheckBox(QStringLiteral("Show settings tab"));
    layout->addWidget(m_showSettings);

    setLayout(layout);
}

void EditorSettingsWidget::load(const EditorConfig &cfg)
{
    m_title->setText(cfg.title);
    m_accent->setText(cfg.accentColor);
    m_background->setText(cfg.backgroundImage);
    m_showNews->setChecked(cfg.showNewsTab);
    m_showSettings->setChecked(cfg.showSettingsTab);
}

EditorConfig EditorSettingsWidget::config() const
{
    EditorConfig cfg = EditorConfig::defaults();
    cfg.title = m_title->text();
    cfg.accentColor = m_accent->text();
    cfg.backgroundImage = m_background->text();
    cfg.showNewsTab = m_showNews->isChecked();
    cfg.showSettingsTab = m_showSettings->isChecked();
    return cfg;
}

QString EditorSettingsWidget::accentColorOf(const EditorConfig &cfg)
{
    return cfg.accentColor;
}

EditorSettingsWidget::EditorSettingsWidget(QWidget *parent)
    : QWidget(parent)
{
    buildUi();
}
