// SPDX-License-Identifier: GPL-3.0-only
#include "CrashReportDialog.h"

#include <QHBoxLayout>
#include <QLabel>
#include <QLineEdit>
#include <QPushButton>
#include <QTextEdit>
#include <QVBoxLayout>

CrashReportDialog::CrashReportDialog(const QString &backendUrl, QWidget *parent)
    : QDialog(parent), m_backendUrl(backendUrl)
{
    setWindowTitle(tr("Crash jelentés"));
    buildUi();
}

void CrashReportDialog::buildUi()
{
    auto *layout = new QVBoxLayout(this);

    auto *info = new QLabel(tr("A launcher összeomlott. Küldj jelentést a fejlesztőknek?"), this);
    info->setWordWrap(true);
    layout->addWidget(info);

    m_messageEdit = new QLineEdit(this);
    m_messageEdit->setPlaceholderText(tr("Opcionális üzenet (mit csináltál?)"));
    layout->addWidget(m_messageEdit);

    m_logView = new QTextEdit(this);
    m_logView->setReadOnly(true);
    m_logView->setMaximumHeight(160);
    layout->addWidget(m_logView);

    auto *btnRow = new QHBoxLayout();
    m_viewLogsButton = new QPushButton(tr("Napló megtekintése"), this);
    m_closeButton = new QPushButton(tr("Bezárás"), this);
    m_sendButton = new QPushButton(tr("Küldés"), this);
    btnRow->addWidget(m_viewLogsButton);
    btnRow->addStretch();
    btnRow->addWidget(m_closeButton);
    btnRow->addWidget(m_sendButton);
    layout->addLayout(btnRow);

    connect(m_viewLogsButton, &QPushButton::clicked, this, &CrashReportDialog::onViewLogs);
    connect(m_closeButton, &QPushButton::clicked, this, &QDialog::reject);
    connect(m_sendButton, &QPushButton::clicked, this, &CrashReportDialog::onSend);
}

QString CrashReportDialog::userMessage() const
{
    return m_messageEdit ? m_messageEdit->text() : QString();
}

QString CrashReportDialog::logText() const
{
    return m_logView ? m_logView->toPlainText() : QString();
}

void CrashReportDialog::setLogText(const QString &text)
{
    if (m_logView)
        m_logView->setPlainText(text);
}

void CrashReportDialog::onSend()
{
    emit sendRequested(userMessage(), logText());
    accept();
}

void CrashReportDialog::onViewLogs()
{
    // Placeholder: actual log loading wired by caller via setLogText.
}
