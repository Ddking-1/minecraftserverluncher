// SPDX-License-Identifier: GPL-3.0-only
#pragma once

#include <QObject>
#include <QString>
#include <QStringList>

class JavaDetector : public QObject
{
    Q_OBJECT
public:
    static bool isJavaAvailable();
    static QString getJavaPath();
    static int javaMajorVersion(const QString &javaExe);
    static QString detectJava24();

private:
    static QStringList candidateJavaHomes();
    static const int kMinJavaMajor;
};
