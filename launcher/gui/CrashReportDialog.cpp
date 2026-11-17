// SPDX-License-Identifier: GPL-3.0-only
#include "CrashReportDialog.h"

#include <QFileDialog>
#include <QFile>
#include <QFileInfo>
#include <QHBoxLayout>
#include <QLabel>
#include <QMessageBox>
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
    if (!m_logView)
        return;
    // Hard cap so a multi-gigabyte log cannot freeze the UI or blow up the
    // report payload. Only the tail is relevant for a crash.
    constexpr int kMaxLogChars = 256 * 1024;
    if (text.size() > kMaxLogChars) {
        m_logView->setPlainText(tr("[A napló csonkolva, teljes hossz: %1 karakter]\n%2")
                                    .arg(QString::number(static_cast<qlonglong>(text.size())),
                                         text.right(kMaxLogChars)));
    } else {
        m_logView->setPlainText(text);
    }
}

void CrashReportDialog::onSend()
{
    if (m_sendButton != nullptr)
        m_sendButton->setEnabled(false);
    emit sendRequested(userMessage(), logText());
    accept();
}

void CrashReportDialog::onViewLogs()
{
    QString path = QFileDialog::getOpenFileName(this, tr("Naplófájl megnyitása"), QString(),
                                                tr("Naplók (*.log *.txt);;Minden fájl (*)"));
    if (path.isEmpty())
        return;
    QFileInfo info(path);
    if (!info.exists() || !info.isFile()) {
        QMessageBox::warning(this, tr("Hiba"), tr("A fájl nem található."));
        return;
    }
    if (info.size() <= 0) {
        QMessageBox::warning(this, tr("Hiba"), tr("A fájl üres."));
        return;
    }
    // Read at most 8 MB; the viewer itself trims further.
    constexpr qint64 kMaxReadBytes = 8 * 1024 * 1024;
    QFile f(path);
    if (!f.open(QIODevice::ReadOnly | QIODevice::Text)) {
        QMessageBox::warning(this, tr("Hiba"), tr("A fájl nem nyitható meg."));
        return;
    }
    const QByteArray data = f.read(kMaxReadBytes);
    f.close();
    if (data.isEmpty()) {
        QMessageBox::warning(this, tr("Hiba"), tr("A fájl nem olvasható."));
        return;
    }
    setLogText(QString::fromUtf8(data));
}
