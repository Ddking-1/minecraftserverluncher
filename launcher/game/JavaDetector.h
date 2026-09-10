// SPDX-License-Identifier: GPL-3.0-only
#pragma once

#include <QObject>
#include <QString>

class JavaDetector : public QObject
{
    Q_OBJECT
public:
    static bool isJavaAvailable();
    static QString getJavaPath();

private:
    static bool checkJavaVersion(const QString &javaPath);
};