// SPDX-License-Identifier: GPL-3.0-only
#pragma once

#include <QDialog>
#include <QString>

class QTextEdit;
class QLineEdit;
class QPushButton;

// Crash report dialog: Close / Send / View logs. Machine info is collected
// in the background; an optional user message can be attached.
class CrashReportDialog : public QDialog
{
    Q_OBJECT
public:
    explicit CrashReportDialog(const QString &backendUrl, QWidget *parent = nullptr);

    QString userMessage() const;
    QString logText() const;
    void setLogText(const QString &text);

signals:
    void sendRequested(const QString &message, const QString &logs);

private slots:
    void onSend();
    void onViewLogs();

private:
    void buildUi();
    QString m_backendUrl;
    QTextEdit *m_logView = nullptr;
    QLineEdit *m_messageEdit = nullptr;
    QPushButton *m_sendButton = nullptr;
    QPushButton *m_closeButton = nullptr;
    QPushButton *m_viewLogsButton = nullptr;
};
