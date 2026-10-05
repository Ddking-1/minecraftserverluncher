// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher UI. Widget for editing EditorConfig fields.
#pragma once

#include "core/EditorConfig.h"

#include <QCheckBox>
#include <QLineEdit>
#include <QVBoxLayout>
#include <QWidget>

// EditorSettingsWidget binds the EditorConfig fields to Qt Widgets controls so
// the local editor can load, edit and save the UI layout configuration.
class EditorSettingsWidget : public QWidget
{
    Q_OBJECT

public:
    explicit EditorSettingsWidget(QWidget *parent = nullptr);

    void load(const EditorConfig &cfg);
    EditorConfig config() const;
    static QString accentColorOf(const EditorConfig &cfg);

private:
    QLineEdit *m_title = nullptr;
    QLineEdit *m_accent = nullptr;
    QLineEdit *m_background = nullptr;
    QCheckBox *m_showNews = nullptr;
    QCheckBox *m_showSettings = nullptr;

    void buildUi();
};
