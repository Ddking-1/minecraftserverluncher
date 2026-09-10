#ifndef MINECRAFTLAUNCHER_H
#define MINECRAFTLAUNCHER_H

#include <QObject>
#include <QProcess>

class MinecraftLauncher : public QObject
{
    Q_OBJECT

public:
    explicit MinecraftLauncher(QObject *parent = nullptr);
    
    void launchGame(const QString &javaPath, const QString &username, const QString &sessionToken, const QString &gameDir);

signals:
    void gameStarted();
    void gameFinished(int exitCode);
    void launchError(const QString &error);

private slots:
    void onProcessStarted();
    void onProcessFinished(int exitCode, QProcess::ExitStatus exitStatus);
    void onProcessError(QProcess::ProcessError error);

private:
    QProcess *m_process;
};

#endif // MINECRAFTLAUNCHER_H