// SPDX-License-Identifier: GPL-3.0-only
#include "LoginDialog.h"

#include <QVBoxLayout>

namespace {

// Mirrors the backend rule: 3-32 ASCII letters, digits, '-' or '_'.
bool isValidUsername(const QString &username)
{
    if (username.size() < 3 || username.size() > 32)
        return false;
    for (const QChar &c : username) {
        const ushort code = c.unicode();
        const bool allowed = (code >= 'a' && code <= 'z')
            || (code >= 'A' && code <= 'Z')
            || (code >= '0' && code <= '9')
            || code == '-' || code == '_';
        if (!allowed)
            return false;
    }
    return true;
}

} // namespace

LoginDialog::LoginDialog(QWidget *parent)
    : QDialog(parent)
{
    buildUi();
    setWindowTitle(tr("Bejelentkezés"));
}

void LoginDialog::buildUi()
{
    QVBoxLayout *layout = new QVBoxLayout(this);

    m_username = new QLineEdit(this);
    m_username->setObjectName(QStringLiteral("usernameEdit"));
    m_username->setPlaceholderText(tr("Felhasználónév"));
    layout->addWidget(m_username);

    m_password = new QLineEdit(this);
    m_password->setObjectName(QStringLiteral("passwordEdit"));
    m_password->setEchoMode(QLineEdit::Password);
    m_password->setPlaceholderText(tr("Jelszó"));
    layout->addWidget(m_password);

    m_error = new QLabel(this);
    m_error->setObjectName(QStringLiteral("loginErrorLabel"));
    m_error->setWordWrap(true);
    m_error->setStyleSheet(QStringLiteral("color: #d33;"));
    m_error->hide();
    layout->addWidget(m_error);

    m_ok = new QPushButton(tr("Bejelentkezés"), this);
    m_ok->setObjectName(QStringLiteral("loginOkButton"));
    m_ok->setDefault(true);
    layout->addWidget(m_ok);

    connect(m_ok, &QPushButton::clicked, this, &LoginDialog::tryAccept);
    connect(m_password, &QLineEdit::returnPressed, this, &LoginDialog::tryAccept);
    connect(m_username, &QLineEdit::returnPressed, this, [this]() {
        m_password->setFocus();
    });
}

void LoginDialog::tryAccept()
{
    const QString username = m_username->text().trimmed();

    const auto fail = [this](const QString &message) {
        m_error->setText(message);
        m_error->show();
    };

    if (!isValidUsername(username)) {
        fail(tr("A felhasználónév 3-32 karakter lehet (betű, szám, '-' vagy '_')."));
        return;
    }
    if (m_password->text().size() < 8) {
        fail(tr("A jelszó legalább 8 karakter."));
        return;
    }

    m_error->hide();
    accept();
}

QString LoginDialog::username() const
{
    return m_username->text();
}

QString LoginDialog::password() const
{
    return m_password->text();
}

void LoginDialog::setUsername(const QString &username)
{
    m_username->setText(username);
    m_username->setCursorPosition(username.size());
}
