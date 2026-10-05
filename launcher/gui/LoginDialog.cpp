// SPDX-License-Identifier: GPL-3.0-only
#include "LoginDialog.h"

#include <QVBoxLayout>

LoginDialog::LoginDialog(QWidget *parent)
    : QDialog(parent)
{
    buildUi();
    setWindowTitle(QStringLiteral("Login"));
}

void LoginDialog::buildUi()
{
    QVBoxLayout *layout = new QVBoxLayout(this);

    m_username = new QLineEdit;
    m_username->setPlaceholderText(QStringLiteral("Username"));
    layout->addWidget(m_username);

    m_password = new QLineEdit;
    m_password->setEchoMode(QLineEdit::Password);
    m_password->setPlaceholderText(QStringLiteral("Password"));
    layout->addWidget(m_password);

    m_ok = new QPushButton(QStringLiteral("OK"));
    layout->addWidget(m_ok);
    connect(m_ok, &QPushButton::clicked, this, &QDialog::accept);
}

QString LoginDialog::username() const
{
    return m_username->text();
}

QString LoginDialog::password() const
{
    return m_password->text();
}
