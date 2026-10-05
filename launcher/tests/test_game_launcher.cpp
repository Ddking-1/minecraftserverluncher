// SPDX-License-Identifier: GPL-3.0-only
#include <QCoreApplication>
#include <QFile>
#include <QTemporaryDir>

#include "core/GameLauncher.h"
#include "game/JavaDetector.h"

#include <cstdio>

static int failures = 0;
static void check(bool cond, const char *msg)
{
    if (!cond) {
        std::fprintf(stderr, "FAIL: %s\n", msg);
        ++failures;
    }
}

int main(int argc, char **argv)
{
    QCoreApplication app(argc, argv);

    GameLauncher launcher;
    LaunchRequest req;
    req.javaPath = QStringLiteral("C:/Windows/System32/cmd.exe"); // implausible
    req.classpath = QStringLiteral("jars");
    req.mainClass = QStringLiteral("net.minecraft.client.main.Main");
    req.sessionToken = QStringLiteral("supersecret-token");
    req.minMemoryMb = 1024;
    req.maxMemoryMb = 4096;
    req.extraJvmArgs = QStringList() << QStringLiteral("-XX:+UseG1GC");
    req.gameArgs = QStringList() << QStringLiteral("--username") << QStringLiteral("alice");

    // Implausible java must be rejected.
    bool started = launcher.start(req);
    check(!started, "implausible java rejected");

    // Arguments never contain the token.
    const QStringList args = launcher.preparedArguments(req);
    bool tokenInArgs = false;
    for (const QString &a : args)
        if (a.contains(QStringLiteral("supersecret-token")))
            tokenInArgs = true;
    check(!tokenInArgs, "token not on command line");

    // Token only in environment.
    const QProcessEnvironment env = launcher.preparedEnvironment(req);
    check(env.value(QStringLiteral("MCSERVER_SESSION_TOKEN")) == QStringLiteral("supersecret-token"),
          "token in environment");
    check(!launcher.preparedEnvironment(req).value(QStringLiteral("MCSERVER_SESSION_TOKEN")).isEmpty()
              || req.sessionToken.isEmpty(),
          "env token set");

    // Valid launch request with real JDK (if present) reaches started.
    const QString java = JavaDetector::getJavaPath();
    if (!java.isEmpty()) {
        QTemporaryDir dir;
        GameLauncher real;
        LaunchRequest good;
        good.javaPath = java;
        good.mainClass = QStringLiteral("java.lang.Object");
        good.classpath = QStringLiteral("none-needed");
        // java -cp none-needed java.lang.Object returns non-zero, but the
        // process should START (gameStarted) even if it exits quickly.
        bool ok = real.start(good);
        check(ok, "real java starts process");
    }

    if (failures == 0)
        std::printf("game_launcher: all checks passed\n");
    return failures == 0 ? 0 : 1;
}
