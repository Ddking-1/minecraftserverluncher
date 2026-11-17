// SPDX-License-Identifier: GPL-3.0-only
#include "FriendlyError.h"

#include <QCoreApplication>
#include <QRegularExpression>
#include <QString>
#include <QStringList>

#include <initializer_list>

namespace {

// True when the raw error already *is* a plain Hungarian sentence produced by
// this module (or by a launch step), so explaining it again changes nothing.
bool isAlreadyFriendly(const QString &raw)
{
    const QString lower = raw.toLower();
    // QStringLiteral, not QLatin1String: these markers contain Hungarian
    // accents, which are multi-byte in the UTF-8 source.
    static const QStringList markers = {
        QStringLiteral("memóriából"),   QStringLiteral("nincs elég"),
        QStringLiteral("nincs bejelentkezve"), QStringLiteral("nem sikerült"),
        QStringLiteral("váratlan hiba"), QStringLiteral("leállt"),
        QStringLiteral("nem található"), QStringLiteral("nem indult"),
        QStringLiteral("csökkentsd"),   QStringLiteral("ellenőrizd"),
        QStringLiteral("próbáld"),      QStringLiteral("telepíts"),
        QStringLiteral("válassz"),      QStringLiteral("frissítsd"),
        QStringLiteral("indítsd"),      QStringLiteral("jelentkezz"),
        QStringLiteral("zárd be"),      QStringLiteral("túl sok"),
        QStringLiteral("foglalt"),      QStringLiteral("nincs beállítva"),
        QStringLiteral("várd meg"),     QStringLiteral("legfeljebb"),
    };
    for (const QString &marker : markers) {
        if (lower.contains(marker))
            return true;
    }
    return false;
}

bool anyOf(const QString &lower, std::initializer_list<const char *> needles)
{
    for (const char *needle : needles) {
        if (lower.contains(QLatin1String(needle)))
            return true;
    }
    return false;
}

// --- plain-language messages -------------------------------------------------
// Each one says what happened and what the player can do about it. Technical
// detail stays in the log file.

const char *const kNotStarted =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "A játék nem indult el. Csökkentsd a RAM beállítást, vagy zárj be más "
                      "programokat, majd próbáld újra.");
const char *const kMemory =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "Nincs elég memória. Csökkentsd a RAM beállítást a Beállításokban, vagy "
                      "zárj be más programokat, majd próbáld újra.");
const char *const kJavaMissing =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "Nincs beállítva Java futtatókörnyezet. Telepíts Java 17-et vagy újabbat, "
                      "vagy válaszd ki a Beállításokban.");
const char *const kUnsafeArgs =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "A játék indítási adatai szabálytalan karaktert tartalmaztak, ezért a "
                      "launcher biztonsági okból nem indította el. Indítsd újra a launchert, és "
                      "ha újra előjön, szólj a szerver üzemeltetőjének.");
const char *const kIncompleteLaunch =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "A játék indítási adatai hiányosak, ezért nem indult el. Indítsd újra a "
                      "launchert, és ha újra előjön, telepítsd újra ezt a verziót.");
const char *const kAlreadyRunning =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "A játék már fut, ezért nem indult el még egyszer. Zárd be a futó "
                      "játékot, ha újra szeretnéd indítani.");
const char *const kJavaVersion =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "A Java verziója nem ismerhető fel. Válassz másik Java-t a "
                      "Beállításokban, vagy telepíts újat.");
const char *const kBadCredentials =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "Hibás a felhasználónév vagy a jelszó, esetleg lejárt a bejelentkezés. "
                      "Jelentkezz be újra.");
const char *const kNotLoggedIn =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "Nincs bejelentkezve, vagy lejárt a munkamenet. Jelentkezz be újra.");
const char *const kTooManyAttempts =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "Túl sok próbálkozás történt. Várj egy kicsit, majd próbáld újra.");
const char *const kUsernameTaken =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "Ez a felhasználónév már foglalt. Válassz másikat.");
const char *const kNoBackend =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "Nincs beállítva szervercím (backend URL). Add meg a Beállításokban, majd "
                      "próbáld újra.");
const char *const kNoInstance =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "A játék példánya nincs betöltve. Indítsd újra a launchert.");
const char *const kNoPermission =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "Nincs jogosultság a fájlhoz vagy mappához. Zárd be a játékot, vagy "
                      "próbáld újra rendszergazdaként.");
const char *const kNoSpace =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "Elfogyott a hely a lemezen. Szabadíts fel helyet, majd próbáld újra.");
const char *const kFileMissing =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "A fájl nem található. Ellenőrizd, hogy nem törölted-e, majd indítsd "
                      "újra a launchert.");
const char *const kDirCreate =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "Nem sikerült létrehozni a szükséges mappát. Ellenőrizd a jogosultságokat "
                      "és a szabad helyet a lemezen.");
const char *const kDownloadBusy =
    QT_TRANSLATE_NOOP("FriendlyError", "Már fut egy letöltés. Várd meg, amíg befejeződik.");
const char *const kBadUrl =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "Érvénytelen letöltési link. Frissítsd az alkalmazást vagy a modpackot.");
const char *const kWriteFailed =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "Nem sikerült a fájl lemezre írása. Ellenőrizd a szabad helyet és a "
                      "jogosultságokat.");
const char *const kChecksum =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "A letöltött fájl sérült (ellenőrzőösszeg-eltérés). Próbáld újra a "
                      "letöltést.");
const char *const kSsl =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "Nem sikerült a biztonságos kapcsolat a szerverrel. Ellenőrizd a gép "
                      "dátumát és idejét, majd próbáld újra.");
const char *const kNetwork =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "Hálózati hiba. Ellenőrizd az internetkapcsolatot, majd próbáld újra.");
const char *const kUpdateSignature =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "A frissítés ellenőrzése nem sikerült (hibás aláírás). Biztonsági okból "
                      "nem telepítjük. Indítsd újra a launchert; ha újra előjön, jelezd a "
                      "fejlesztőnek.");
const char *const kUpdateBroken =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "A frissítés adatai sérültek vagy hiányosak. Próbáld újra később; ha nem "
                      "javul, töltsd le újra a telepítőt.");
const char *const kSkinInvalid =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "A kiválasztott fájl nem megfelelő skin. Válassz PNG képet (64×64 vagy "
                      "64×32), majd próbáld újra.");
const char *const kSkinUpload =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "A skin feltöltése nem sikerült. Ellenőrizd, hogy be vagy jelentkezve, majd "
                      "próbáld újra.");
const char *const kAvatarLoad =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "A skin feltöltődött, de az előnézet (avatar) nem tölthető be. Ellenőrizd "
                      "az internetkapcsolatot, majd nyisd meg újra a launchert.");
const char *const kFileEmpty =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "A kiválasztott fájl üres. Válassz másik fájlt, vagy indítsd újra a "
                      "játékot, hogy új napló készüljön.");
const char *const kLogOpen =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "A naplófájl nem nyitható meg. Zárd be a játékot vagy azt a programot, "
                      "amely használja, majd próbáld újra.");
const char *const kLogRead =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "A naplófájl tartalma nem olvasható (sérült vagy nem szöveges fájl). "
                      "Válassz másik naplófájlt.");
const char *const kSettingsMemory =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "Érvénytelen memória-beállítás. A maximum nem lehet kevesebb a minimumnál, "
                      "és legfeljebb 65536 MB lehet. Válassz érvényes értékeket a "
                      "Beállításokban.");
const char *const kProcessCrash =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "A játék összeomlott. Indítsd újra; ha rendszeresen előfordul, csökkentsd a "
                      "RAM-ot, vagy ellenőrizd a modokat.");
const char *const kProcessTimedOut =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "A játék indítása túl sokáig tartott. Ellenőrizd a RAM beállítást és a "
                      "Java-t, majd próbáld újra.");
const char *const kStartFailed =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "A játék elindítása nem sikerült. Ellenőrizd, hogy a megfelelő Java telepítve "
                      "van-e, majd próbáld újra.");
const char *const kServerError =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "A szerver hibát jelzett. Próbáld újra később.");
const char *const kNotFound =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "A kért adat nem található a szerveren. Frissítsd az alkalmazást, majd "
                      "próbáld újra.");
const char *const kInvalidRequest =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "A kérés érvénytelen volt. Próbáld újra; ha megismétlődik, frissítsd az "
                      "alkalmazást.");
const char *const kInvalidResponse =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "Érvénytelen válasz érkezett a szerverről. Próbáld újra később.");
const char *const kGeneric =
    QT_TRANSLATE_NOOP("FriendlyError",
                      "Váratlan hiba történt. Indítsd újra a launchert, és ha újra előjön, "
                      "nézd meg a naplót, majd jelezd a fejlesztőnek.");

QString httpMessage(int code)
{
    if (code == 0)
        return QCoreApplication::translate("FriendlyError", kNetwork);
    if (code == 400)
        return QCoreApplication::translate("FriendlyError", kInvalidRequest);
    if (code == 401 || code == 403)
        return QCoreApplication::translate("FriendlyError", kNotLoggedIn);
    if (code == 404)
        return QCoreApplication::translate("FriendlyError", kNotFound);
    if (code == 409)
        return QCoreApplication::translate("FriendlyError", kUsernameTaken);
    if (code == 429)
        return QCoreApplication::translate("FriendlyError", kTooManyAttempts);
    if (code >= 500)
        return QCoreApplication::translate("FriendlyError", kServerError);
    if (code >= 400)
        return QCoreApplication::translate(
            "FriendlyError", "A szerver elutasította a kérést (%1). Próbáld újra később.")
            .arg(code);
    return QString();
}

} // namespace

namespace FriendlyError {

QString explain(const QString &raw)
{
    if (raw.trimmed().isEmpty())
        return QCoreApplication::translate("FriendlyError", kGeneric);

    if (isAlreadyFriendly(raw))
        return raw;

    const QString lower = raw.toLower();

    const auto say = [](const char *message) {
        return QCoreApplication::translate("FriendlyError", message);
    };

    // A launch that never started: checked before the memory rules, because the
    // message also mentions the requested amount of memory.
    if (anyOf(lower, {"did not start"}))
        return say(kNotStarted);

    // --- launch request validation (core/GameLauncher) ----------------------
    if (anyOf(lower, {"already running"}))
        return say(kAlreadyRunning);
    if (anyOf(lower, {"unsafe main class", "unsafe classpath", "unsafe jvm argument",
                      "unsafe game argument"}))
        return say(kUnsafeArgs);
    if (anyOf(lower, {"empty classpath", "empty main class"}))
        return say(kIncompleteLaunch);

    // --- settings -----------------------------------------------------------
    // Before the memory rules: these sentences also contain the word "memory".
    if (anyOf(lower, {"maximum memory", "minimum memory", "memory must",
                      "must not exceed", "below minimum", "exceeds max memory",
                      "min memory exceeds"}))
        return say(kSettingsMemory);

    // --- memory -------------------------------------------------------------
    if (anyOf(lower, {"outofmemoryerror", "out of memory", "heap space", "not enough ram",
                      "cannot allocate memory", "insufficient memory",
                      "reserve enough space", "memory", " ram"}))
        return say(kMemory);

    // --- Java runtime -------------------------------------------------------
    if (anyOf(lower, {"could not query java", "java version", "parse java"}))
        return say(kJavaVersion);
    if (anyOf(lower, {"no java runtime", "no java", "java runtime", "configured java",
                      "java executable", "implausible java", "java not found",
                      "java: not found", "java home"}))
        return say(kJavaMissing);

    // --- account / session --------------------------------------------------
    if (anyOf(lower, {"invalid or expired credentials", "invalid credentials",
                      "incorrect password", "wrong password", "unauthorized"}))
        return say(kBadCredentials);
    if (anyOf(lower, {"too many attempts", "too many requests", "rate limit"}))
        return say(kTooManyAttempts);
    if (anyOf(lower, {"already taken", "already exists"}))
        return say(kUsernameTaken);
    if (anyOf(lower, {"not logged in", "no stored session", "no token", "missing token"}))
        return say(kNotLoggedIn);
    if (anyOf(lower, {"backup codes", "totp", "2fa", "two-factor"}))
        return say(kBadCredentials);
    if (anyOf(lower, {"session", "authentic"}))
        return say(kNotLoggedIn);

    // --- backend configuration ---------------------------------------------
    if (anyOf(lower, {"no backend", "backend url", "no server url"}))
        return say(kNoBackend);

    // --- instance -----------------------------------------------------------
    if (anyOf(lower, {"no instance"}))
        return say(kNoInstance);

    // --- skin ---------------------------------------------------------------
    // Before the skin rule: an avatar preview failure is not a bad skin file.
    if (lower.contains(QLatin1String("avatar")) &&
        anyOf(lower, {"download", "load", "failed", "could not", "sikertelen", "betölt"}))
        return say(kAvatarLoad);
    if (anyOf(lower, {"no skin selected", "skin", "avatar"}))
        return raw.contains(QLatin1String("upload"), Qt::CaseInsensitive) ? say(kSkinUpload)
                                                                         : say(kSkinInvalid);

    // --- filesystem ---------------------------------------------------------
    if (anyOf(lower, {"access denied", "permission denied", "permissionerror",
                      "operation not permitted"}))
        return say(kNoPermission);
    if (anyOf(lower, {"no space", "disk full", "not enough space", "quota exceeded"}))
        return say(kNoSpace);
    // Log viewing has its own wording: telling the player to restart the
    // launcher does not help when they picked an unreadable log file.
    if (anyOf(lower, {"could not open log file"}))
        return say(kLogOpen);
    if (anyOf(lower, {"could not read log file"}))
        return say(kLogRead);
    if (anyOf(lower, {"file is empty", "empty file", "log file is empty"}))
        return say(kFileEmpty);
    if (anyOf(lower, {"no such file", "file not found", "does not exist", "missing file",
                      "cannot open", "failed to open"}))
        return say(kFileMissing);

    // --- downloads ----------------------------------------------------------
    if (anyOf(lower, {"already in progress"}))
        return say(kDownloadBusy);
    if (anyOf(lower, {"invalid download url", "unsupported url scheme", "invalid url"}))
        return say(kBadUrl);
    if (anyOf(lower, {"could not create", "directory", "mkdir"}))
        return say(kDirCreate);
    if (anyOf(lower, {"could not open destination", "write failed", "commit failed"}))
        return say(kWriteFailed);
    if (anyOf(lower, {"sha-256", "checksum", "mismatch"}))
        return say(kChecksum);

    // --- game process -------------------------------------------------------
    // Before the network rules: "Process timed out" is a start failure, not a
    // download problem.
    if (anyOf(lower, {"process timed out"}))
        return say(kProcessTimedOut);

    // --- network ------------------------------------------------------------
    if (anyOf(lower, {"ssl", "certificate", "tls", "handshake"}))
        return say(kSsl);
    // Qt words a DNS failure as "Host api.example.com not found", which must not
    // be mistaken for a backend 404 ("Not found").
    if (lower.contains(QLatin1String("host")) &&
        anyOf(lower, {"not found", "unreachable", "resolve", "lookup"}))
        return say(kNetwork);
    if (anyOf(lower, {"timeout", "timed out", "connection", "refused", "network",
                      "host not found", "unreachable", "reset by peer", "offline",
                      "name resolution", "socket"}))
        return say(kNetwork);

    // --- updates ------------------------------------------------------------
    if (anyOf(lower, {"signature"}))
        return say(kUpdateSignature);
    if (anyOf(lower, {"manifest", "update object", "update"}))
        return say(kUpdateBroken);

    // --- game process -------------------------------------------------------
    if (anyOf(lower, {"crashed", "crash"}))
        return say(kProcessCrash);
    if (anyOf(lower, {"failed to start", "could not start", "start process", "process error",
                      "invalid parameters", "unknown error"}))
        return say(kStartFailed);

    // --- HTTP status codes --------------------------------------------------
    {
        const QRegularExpression re(QStringLiteral("http\\D{0,4}(\\d{1,3})"),
                                    QRegularExpression::CaseInsensitiveOption);
        const auto m = re.match(raw);
        if (m.hasMatch()) {
            const QString hm = httpMessage(m.captured(1).toInt());
            if (!hm.isEmpty())
                return hm;
        }
    }

    // --- game process exit --------------------------------------------------
    {
        const QRegularExpression re(QStringLiteral("exited\\s*\\(code\\s*(\\d+)\\)"),
                                    QRegularExpression::CaseInsensitiveOption);
        const auto m = re.match(raw);
        if (m.hasMatch()) {
            const int code = m.captured(1).toInt();
            if (code == 0)
                return QCoreApplication::translate("FriendlyError", "A játék normálisan leállt.");
            return QCoreApplication::translate(
                       "FriendlyError",
                       "A játék váratlanul leállt (kód %1). Indítsd újra; ha újra előjön, "
                       "nézd meg a naplót.")
                .arg(code);
        }
    }

    // --- generic server wording --------------------------------------------
    if (anyOf(lower, {"server error", "internal server"}))
        return say(kServerError);
    if (anyOf(lower, {"not found"}))
        return say(kNotFound);
    if (anyOf(lower, {"invalid response"}))
        return say(kInvalidResponse);
    if (anyOf(lower, {"invalid request", "bad request", "invalid"}))
        return say(kInvalidRequest);
    if (anyOf(lower, {"error", "failed", "failure", "unable"}))
        return say(kServerError);

    // Fallback: never echo the raw technical text (callers log it with qWarning),
    // but still tell the player what they can do.
    return say(kGeneric);
}

} // namespace FriendlyError
