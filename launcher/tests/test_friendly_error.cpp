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

namespace {

struct Case
{
    const char *raw;
    const char *expected; // a fragment the plain-language message must contain
    const char *label;
};

} // namespace

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
    // The raw text belongs in the log, not on the screen; the player needs an action.
    check(!fb.contains(QStringLiteral("totally unknown")), "fallback hides the raw text");
    check(fb.contains(QStringLiteral("Indítsd újra")), "fallback tells the player what to do");

    // Empty.
    check(FriendlyError::explain(QString()).contains(QStringLiteral("Váratlan hiba")),
          "empty error gets a generic fallback");

    // Idempotency.
    const QString once = FriendlyError::explain(QStringLiteral("session is not authenticated"));
    check(FriendlyError::explain(once) == once, "explaining twice is stable");

    // --- every raw error the launcher can produce ---------------------------
    // Each entry is a string the launcher actually emits or receives. The
    // translation must be Hungarian (not the raw text) and must carry the
    // expected hint, so a player always gets a sentence that says what to do.
    static const Case cases[] = {
        // launch / process
        {"Invalid parameters", "nem sikerült", "bad launch parameters"},
        {"Game directory does not exist: C:/games/inst", "nem található",
         "missing game directory"},
        {"Failed to start process", "nem sikerült", "process start failure"},
        {"Process crashed", "összeomlott", "process crash"},
        {"Process timed out", "túl sokáig", "process start timeout"},
        {"Unknown error occurred", "nem sikerült", "unknown process error"},
        {"Process error: 3", "nem sikerült", "numbered process error"},
        {"the game did not start with 4096 MB of memory (x)", "nem indult el",
         "game did not start"},

        // memory
        {"java.lang.OutOfMemoryError: Java heap space", "memória", "heap OOM"},
        {"There is insufficient memory for the Java Runtime Environment", "memória",
         "JRE memory exhaustion"},
        {"Cannot allocate memory", "memória", "allocation failure"},

        // java
        {"no Java runtime configured", "Java futtatókörnyezet", "missing Java"},
        {"invalid or implausible java executable", "Java futtatókörnyezet", "bad Java path"},
        {"could not query Java version (exit 1)", "Java verziója", "unreadable Java version"},

        // account / session
        {"Invalid or expired credentials", "Jelentkezz be", "expired credentials"},
        {"That username is already taken", "foglalt", "taken username"},
        {"Too many attempts, please wait and try again", "Túl sok próbálkozás",
         "rate limited login"},
        {"not logged in", "Jelentkezz be", "not logged in"},
        {"no stored session", "Jelentkezz be", "missing stored session"},

        // backend / network
        {"no backend configured", "szervercím", "missing backend URL"},
        {"Invalid request", "érvénytelen", "rejected request"},
        {"Not found", "nem található", "backend not found"},
        {"Server error, please try again later", "szerver hibát", "backend 5xx"},
        {"Request failed (HTTP 418)", "elutasította", "unexpected HTTP status"},
        {"invalid response from server", "Érvénytelen válasz", "malformed backend reply"},
        {"Host not found", "Hálózati hiba", "DNS failure"},
        {"SSL handshake failed", "biztonságos kapcsolat", "TLS failure"},
        {"request timed out (5s)", "Hálózati hiba", "request timeout"},
        // Real QNetworkReply::errorString() values seen from Qt itself.
        {"Host api.example.com not found", "Hálózati hiba", "Qt DNS errorString"},
        {"Connection refused", "Hálózati hiba", "Qt refused errorString"},
        {"Error transferring https://api.example.com/x - server replied: Service Unavailable",
         "szerver", "Qt transfer errorString"},
        {"Error downloading https://api.example.com/x - server replied: Not Found",
         "nem található", "Qt download errorString"},
        {"Error transferring https://api.example.com/x - server replied: Unauthorized",
         "Jelentkezz be", "Qt unauthorized errorString"},
        {"TLS initialization failed", "biztonságos kapcsolat", "Qt TLS errorString"},
        // core/GameLauncher launch-request validation.
        {"implausible memory settings", "memória", "GameLauncher bad memory"},
        {"min memory exceeds max memory", "memória", "GameLauncher min > max"},
        {"unsafe main class", "szabálytalan", "GameLauncher main class injection"},
        {"unsafe classpath", "szabálytalan", "GameLauncher classpath injection"},
        {"unsafe jvm argument", "szabálytalan", "GameLauncher jvm injection"},
        {"unsafe game argument", "szabálytalan", "GameLauncher game arg injection"},
        {"invalid or implausible java executable", "Java", "GameLauncher bad java"},
        {"empty classpath", "hiányosak", "GameLauncher empty classpath"},
        {"empty main class", "hiányosak", "GameLauncher empty main class"},
        {"a launch is already running", "már fut", "GameLauncher double launch"},
        {"failed to start process", "nem sikerült", "GameLauncher start failure"},
        {"execvp: No such file or directory", "nem található", "Qt start failure text"},

        // filesystem
        {"Permission denied", "jogosultság", "denied write"},
        {"No space left on device", "Elfogyott a hely", "disk full"},
        {"No such file or directory", "nem található", "missing file"},
        {"could not create natives directory: C:/x", "mappát", "directory creation"},

        // downloads
        {"could not open destination file", "lemezre írása", "destination open failure"},
        {"write failed", "lemezre írása", "write failure"},
        {"SHA-256 mismatch", "sérült", "checksum mismatch"},

        // updates
        {"invalid manifest JSON", "frissítés", "corrupt update manifest"},
        {"missing update object", "frissítés", "incomplete update manifest"},
        {"manifest signature verification failed", "aláírás", "untrusted update"},

        // skins
        {"invalid skin PNG", "skin", "bad skin file"},
        {"No skin selected", "skin", "no skin chosen"},
        {"upload failed (HTTP 413)", "elutasította", "rejected skin upload"},

        // settings
        {"maximum memory must not be below minimum memory", "memória-beállítás",
         "inverted memory range"},
        {"maximum memory must not exceed 65536 MB", "memória-beállítás",
         "oversized memory request"},

        // HTTP statuses
        {"HTTP 401", "Jelentkezz be", "HTTP 401"},
        {"HTTP 403", "Jelentkezz be", "HTTP 403"},
        {"HTTP 409", "foglalt", "HTTP 409"},
        {"HTTP 429", "Túl sok próbálkozás", "HTTP 429"},

        // news panel (MainWindow)
        {"invalid response", "Próbáld újra", "news body is not JSON"},

        // skin upload (SkinPreviewWidget / MainWindow)
        {"skin upload failed", "Ellenőrizd", "skin upload without detail"},
        {"avatar download failed", "előnézet", "avatar preview download"},
        {"Avatar betöltése sikertelen.", "előnézet", "legacy avatar message"},

        // crash report log viewer (CrashReportDialog)
        {"could not read log file: file not found", "Válassz másik naplófájlt",
         "missing crash log"},
        {"log file is empty", "Válassz másik fájlt", "empty crash log"},
        {"could not open log file", "Zárd be a játékot", "locked crash log"},
        {"could not read log file", "Válassz másik naplófájlt", "unreadable crash log"},

        // modpack import (MainWindow): PowerShell stderr must never be shown raw
        {"Expand-Archive : The archive entry was compressed using an unsupported "
         "compression method.",
         "Indítsd újra", "modpack extract failure"},
    };

    for (const Case &c : cases) {
        const QString raw = QString::fromUtf8(c.raw);
        const QString out = FriendlyError::explain(raw);
        if (out == raw || !out.contains(QString::fromUtf8(c.expected))
            || FriendlyError::explain(out) != out) {
            std::fprintf(stderr, "  raw=[%s] got=[%s]\n", c.raw, out.toUtf8().constData());
        }
        check(out != raw, c.label);
        check(out.contains(QString::fromUtf8(c.expected)), c.label);
        check(FriendlyError::explain(out) == out, c.label);
    }

    // No technical English wording may survive: a player must never see the raw
    // Qt/backend text, only the Hungarian explanation.
    static const char *const kEnglish[] = {
        "refused",  "Connection", "Error transferring", "Error downloading",
        "timed out", "Unable to",  "failed to",          "not found",
        "denied",   "HTTP ",      "manifest",           "signature",
    };
    for (const Case &c : cases) {
        const QString out = FriendlyError::explain(QString::fromUtf8(c.raw));
        bool clean = true;
        for (const char *word : kEnglish) {
            if (out.contains(QString::fromLatin1(word))) {
                clean = false;
                std::fprintf(stderr, "  raw=[%s] kept English [%s]: [%s]\n", c.raw, word,
                             out.toUtf8().constData());
            }
        }
        check(clean, c.label);
    }

    // The point of the whole module: a basic user must be told what to DO, not
    // only that something failed. Every message for an error case has to carry
    // an action, and every message shown by the GUI surfaces (news, skin, login,
    // session restore, crash report) goes through this table.
    static const char *const kActions[] = {
        "próbáld",  "ellenőrizd", "indítsd",  "válassz", "jelentkezz",
        "csökkentsd", "frissítsd", "zárd",    "telepíts", "add meg",
        "várd",     "szólj",      "töltsd",  "nyisd",   "adj meg",
        "nézd meg", "várj",       "indíts újra",
    };
    for (const Case &c : cases) {
        const QString raw = QString::fromUtf8(c.raw);
        // A clean exit is not an error and needs no instruction.
        if (raw.contains(QStringLiteral("code 0")))
            continue;
        const QString out = FriendlyError::explain(raw);
        bool hasAction = false;
        const QString outLower = out.toLower();
        for (const char *action : kActions) {
            if (outLower.contains(QString::fromUtf8(action))) {
                hasAction = true;
                break;
            }
        }
        if (!hasAction)
            std::fprintf(stderr, "  raw=[%s] has no action: [%s]\n", c.raw,
                         out.toUtf8().constData());
        check(hasAction, c.label);
    }

    if (failures == 0)
        std::printf("friendly_error: all checks passed\n");
    return failures == 0 ? 0 : 1;
}
