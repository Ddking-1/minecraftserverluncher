// SPDX-License-Identifier: GPL-3.0-only
#include "JavaDetector.h"

#include <QDir>
#include <QProcessEnvironment>
#include <QStringList>

bool JavaDetector::isJavaAvailable()
{
    // Check JAVA_HOME environment variable first
    QProcessEnvironment env = QProcessEnvironment::systemEnvironment();
    QString javaHome = env.value("JAVA_HOME");
    
    if (!javaHome.isEmpty()) {
        QDir dir(javaHome);
        if (dir.exists() && checkJavaVersion(javaHome)) {
            return true;
        }
    }

    // Check standard Java installation paths
    QStringList possiblePaths = {
        "C:/Program Files/Java/",
        "C:/Program Files (x86)/Java/"
    };

    for (const QString &path : possiblePaths) {
        QDir dir(path);
        if (dir.exists()) {
            // Look for JDK or JRE directories with version numbers
            QStringList subdirs = dir.entryList(QDir::Dirs | QDir::NoDotAndDotDot);
            
            for (const QString &subdir : subdirs) {
                if (subdir.startsWith("jdk") || subdir.startsWith("jre")) {
                    // Check the bin directory in this path
                    QString javaPath = QString("%1/%2/bin/java.exe").arg(path, subdir);
                    
                    QDir testDir(javaPath);
                    if (testDir.exists()) {
                        return true;
                    }
                }
            }
        }
    }

    return false;
}

QString JavaDetector::getJavaPath()
{
    QProcessEnvironment env = QProcessEnvironment::systemEnvironment();
    QString javaHome = env.value("JAVA_HOME");
    
    if (!javaHome.isEmpty()) {
        QDir dir(javaHome);
        if (dir.exists() && checkJavaVersion(javaHome)) {
            return javaHome + "/bin/java.exe";
        }
    }

    // Check standard Java installation paths
    QStringList possiblePaths = {
        "C:/Program Files/Java/",
        "C:/Program Files (x86)/Java/"
    };

    for (const QString &path : possiblePaths) {
        QDir dir(path);
        if (dir.exists()) {
            QStringList subdirs = dir.entryList(QDir::Dirs | QDir::NoDotAndDotDot);
            
            for (const QString &subdir : subdirs) {
                if (subdir.startsWith("jdk") || subdir.startsWith("jre")) {
                    // Check the bin directory in this path
                    QString javaPath = QString("%1/%2/bin/java.exe").arg(path, subdir);
                    
                    QDir testDir(javaPath);
                    if (testDir.exists()) {
                        return javaPath;
                    }
                }
            }
        }
    }

    return {};
}

bool JavaDetector::checkJavaVersion(const QString &javaPath)
{
    // Simple check for 64-bit JRE/JDK
    QProcess process;
    QStringList args = {"-version"};
    
    if (process.startDetached(javaPath + "/bin/java.exe", args)) {
        // Wait a bit to let the command finish
        process.waitForFinished(1000);
        
        QString output = process.readAllStandardError();
        
        return output.contains("64-Bit") || 
               output.contains("64-bit") ||
               (output.contains("Java(TM)") && !output.contains("32-Bit"));
    }
    
    return false;
}