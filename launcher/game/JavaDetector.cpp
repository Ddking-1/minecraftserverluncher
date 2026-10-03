// SPDX-License-Identifier: GPL-3.0-only
#include "JavaDetector.h"

#include <QDir>
#include <QFile>
#include <QProcess>
#include <QProcessEnvironment>
#include <QRegularExpression>
#include <QStringList>

const int JavaDetector::kMinJavaMajor = 21;

QStringList JavaDetector::candidateJavaHomes()
{
    QStringList result;

    if (!qEnvironmentVariable("LAUNCHER_JAVA_HOME").isEmpty()) {
        result << qEnvironmentVariable("LAUNCHER_JAVA_HOME");
    }
    if (!qEnvironmentVariable("JAVA_HOME").isEmpty()) {
        result << qEnvironmentVariable("JAVA_HOME");
    }
    const QStringList roots = {
        "C:/Program Files/Eclipse Adoptium",
        "C:/Program Files/Java",
        "C:/Program Files/Microsoft",
        "C:/Program Files/Zulu",
        "C:/Program Files/Amazon Corretto",
        "C:/Program Files (x86)/Java",
        QDir::homePath() + QStringLiteral("/.jdks")
    };
    for (const QString &root : roots) {
        QDir dir(root);
        if (dir.exists()) {
            const QStringList subdirs = dir.entryList(QDir::Dirs | QDir::NoDotAndDotDot);
            for (const QString &subdir : subdirs) {
                result << root + "/" + subdir;
            }
        }
    }
    return result;
}

int JavaDetector::javaMajorVersion(const QString &javaExe)
{
    if (javaExe.isEmpty())
        return -1;

    QProcess p;
    p.start(javaExe, QStringList() << "-version");
    if (!p.waitForStarted(3000))
        return -1;
    p.waitForFinished(5000);

    QByteArray out = p.readAllStandardError() + p.readAllStandardOutput();
    QString text = QString::fromLocal8Bit(out);

    QRegularExpression re("([0-9][0-9A-Za-z._-]*)");
    auto m = re.match(text);
    if (!m.hasMatch())
        return -1;

    QString ver = m.captured(1);
    QStringList parts = ver.split('.');

    if (parts.size() >= 2 && parts[0] == "1") {
        return parts[1].toInt();
    }
    return parts[0].toInt();
}

QString JavaDetector::detectJava24()
{
    for (const QString &home : candidateJavaHomes()) {
        QString exe = home + "/bin/java.exe";
        if (QFile::exists(exe) && javaMajorVersion(exe) >= kMinJavaMajor) {
            return exe;
        }
    }
    return "";
}

bool JavaDetector::isJavaAvailable()
{
    return !detectJava24().isEmpty();
}

QString JavaDetector::getJavaPath()
{
    return detectJava24();
}
