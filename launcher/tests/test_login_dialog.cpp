// SPDX-License-Identifier: GPL-3.0-only
#include <QApplication>
#include <QTimer>

#include "gui/LoginDialog.h"

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
    QApplication app(argc, argv);
    LoginDialog dlg;
    dlg.setModal(false);
    dlg.show();

    check(dlg.username().isEmpty(), "username should be empty initially");
    check(dlg.password().isEmpty(), "password should be empty initially");
    check(!dlg.windowTitle().isEmpty(), "dialog should have a window title");

    QTimer::singleShot(0, &app, &QCoreApplication::quit);

    if (failures == 0)
        std::printf("login_dialog: all checks passed\n");
    return failures == 0 ? 0 : 1;
}
