// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher UI. Simple login dialog collecting credentials.
#pragma once

#include <QDialog>
#include <QLineEdit>
#include <QPushButton>
#include <QString>

class LoginDialog : public QDialog
{
    Q_OBJECT
public:
    explicit LoginDialog(QWidget *parent = nullptr);

    QString username() const;
    QString password() const;

private:
    QLineEdit *m_username = nullptr;
    QLineEdit *m_password = nullptr;
    QPushButton *m_ok = nullptr;

    void buildUi();
};
