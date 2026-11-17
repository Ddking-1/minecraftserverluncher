// SPDX-License-Identifier: GPL-3.0-only
#include <QCoreApplication>
#include <QString>

#include "core/FriendlyError.h"

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

    // Pass-through of already-friendly Hungarian messages.
    const QString mem = QStringLiteral("A játék elfogyott a memóriából és leállt. Kért memória: 6144 MB.");
    check(FriendlyError::explain(mem) == mem, "friendly memory message passes through");

    // Memory (technical).
    check(FriendlyError::explain(QStringLiteral("java.lang.OutOfMemoryError: Java heap space"))
              .contains(QStringLiteral("memória")),
          "OOM becomes a memory message");
    check(FriendlyError::explain(QStringLiteral("not enough RAM")).contains(QStringLiteral("memória")),
          "RAM wording becomes a memory message");

    // Java runtime.
    check(FriendlyError::explain(QStringLiteral("no Java runtime configured for this instance"))
              .contains(QStringLiteral("Java futtatókörnyezet")),
          "missing Java becomes a clear message");
    check(FriendlyError::explain(QStringLiteral("configured Java runtime not found: C:/java"))
              .contains(QStringLiteral("Java futtatókörnyezet")),
          "missing Java binary becomes a clear message");
    check(FriendlyError::explain(QStringLiteral("could not parse Java version output"))
              .contains(QStringLiteral("Java verziója")),
          "unparseable Java version becomes a clear message");

    // Session.
    check(FriendlyError::explain(QStringLiteral("session is not authenticated"))
              .contains(QStringLiteral("bejelentkezve")),
          "auth failure becomes a login message");
    check(FriendlyError::explain(QStringLiteral("no authenticated session available"))
              .contains(QStringLiteral("bejelentkezve")),
          "missing session becomes a login message");

    // Instance.
    check(FriendlyError::explain(QStringLiteral("no instance to launch"))
              .contains(QStringLiteral("példánya nincs betöltve")),
          "missing instance becomes a clear message");

    // Directories.
    check(FriendlyError::explain(QStringLiteral("could not create natives directory: C:/x"))
              .contains(QStringLiteral("mappát")),
          "directory failure becomes a clear message");

    // Downloads.
    check(FriendlyError::explain(QStringLiteral("a download is already in progress"))
              .contains(QStringLiteral("Már fut egy letöltés")),
          "concurrent download becomes a clear message");
    check(FriendlyError::explain(QStringLiteral("invalid download URL"))
              .contains(QStringLiteral("Érvénytelen letöltési link")),
          "bad URL becomes a clear message");
    check(FriendlyError::explain(QStringLiteral("unsupported URL scheme"))
              .contains(QStringLiteral("Érvénytelen letöltési link")),
          "bad scheme becomes a clear message");
    check(FriendlyError::explain(QStringLiteral("HTTP 404")).contains(QStringLiteral("nem található")),
          "HTTP 404 becomes a clear message");
    check(FriendlyError::explain(QStringLiteral("HTTP 500")).contains(QStringLiteral("szerver hibát")),
          "HTTP 5xx becomes a clear message");
    check(FriendlyError::explain(QStringLiteral("could not open destination file"))
              .contains(QStringLiteral("fájl lemezre írása")),
          "file open failure becomes a clear message");
    check(FriendlyError::explain(QStringLiteral("SHA-256 mismatch"))
              .contains(QStringLiteral("sérült")),
          "checksum mismatch becomes a clear message");
    check(FriendlyError::explain(QStringLiteral("Connection timed out"))
              .contains(QStringLiteral("Hálózati hiba")),
          "network timeout becomes a clear message");

    // Game process exit.
    check(FriendlyError::explain(QStringLiteral("game process exited (code 137)"))
              .contains(QStringLiteral("leállt (kód 137)")),
          "crash exit code is reported plainly");
    check(FriendlyError::explain(QStringLiteral("game process exited (code 0)"))
              .contains(QStringLiteral("normálisan leállt")),
          "clean exit is reported plainly");
    check(FriendlyError::explain(QStringLiteral("the game did not start with 4096 MB of memory (x)"))
              .contains(QStringLiteral("nem indult el")),
          "failed start becomes a clear message");

    // Fallback.
    const QString fb = FriendlyError::explain(QStringLiteral("totally unknown xyz 123"));
    check(fb.contains(QStringLiteral("Váratlan hiba")), "unknown error gets a generic fallback");
    check(fb.contains(QStringLiteral("totally unknown")), "fallback keeps a snippet");

    // Empty.
    check(FriendlyError::explain(QString()).contains(QStringLiteral("Váratlan hiba")),
          "empty error gets a generic fallback");

    // Idempotency.
    const QString once = FriendlyError::explain(QStringLiteral("session is not authenticated"));
    check(FriendlyError::explain(once) == once, "explaining twice is stable");

    if (failures == 0)
        std::printf("friendly_error: all checks passed\n");
    return failures == 0 ? 0 : 1;
}
