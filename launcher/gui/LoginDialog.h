// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher UI. Login dialog collecting and validating credentials.
#pragma once

#include <QDialog>
#include <QLabel>
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

    // Pre-fills the username field (e.g. with the last account that logged in).
    void setUsername(const QString &username);

private slots:
    void tryAccept();

private:
    QLineEdit *m_username = nullptr;
    QLineEdit *m_password = nullptr;
    QPushButton *m_ok = nullptr;
    QLabel *m_error = nullptr;

    void buildUi();
};
