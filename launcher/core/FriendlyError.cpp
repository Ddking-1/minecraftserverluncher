// SPDX-License-Identifier: GPL-3.0-only
#include "FriendlyError.h"

#include <QRegularExpression>
#include <QString>

namespace {

bool isAlreadyFriendly(const QString &raw)
{
    const QString lower = raw.toLower();
    return lower.contains(QStringLiteral("memóriából")) ||
           lower.contains(QStringLiteral("csökkentsd")) ||
           lower.contains(QStringLiteral("nincs elég")) ||
           lower.contains(QStringLiteral("nincs bejelentkezve")) ||
           lower.contains(QStringLiteral("nem sikerült")) ||
           lower.contains(QStringLiteral("váratlan hiba")) ||
           lower.contains(QStringLiteral("leállt")) ||
           lower.contains(QStringLiteral("nem található")) ||
           lower.contains(QStringLiteral("nem indult"));
}

QString httpMessage(const QString &raw)
{
    const QRegularExpression re(QStringLiteral("HTTP\\s*(\\d{3})"), QRegularExpression::CaseInsensitiveOption);
    const auto m = re.match(raw);
    if (!m.hasMatch())
        return QString();
    const int code = m.captured(1).toInt();
    if (code == 404)
        return QStringLiteral("A kért fájl nem található a szerveren. Frissítsd az alkalmazást vagy a modpackot.");
    if (code == 403)
        return QStringLiteral("Hozzáférés megtagadva a szerveren. Próbáld újra később.");
    if (code >= 500)
        return QStringLiteral("A szerver hibát jelzett (%1). Próbáld újra később.").arg(code);
    if (code >= 400)
        return QStringLiteral("A szerver elutasította a kérést (%1).").arg(code);
    return QString();
}

} // namespace

namespace FriendlyError {

QString explain(const QString &raw)
{
    if (raw.trimmed().isEmpty())
        return QStringLiteral("Váratlan hiba történt. Részletek a naplóban.");

    if (isAlreadyFriendly(raw))
        return raw;

    const QString lower = raw.toLower();

    // Failed start (checked before the generic memory branch, because that
    // message also contains "memory").
    if (lower.contains(QStringLiteral("did not start"))) {
        return QStringLiteral("A játék nem indult el. Csökkentsd a RAM beállítást, vagy zárj be más programokat, majd próbáld újra.");
    }

    // Memory.
    if (lower.contains(QStringLiteral("memory")) || lower.contains(QStringLiteral("ram")) ||
        lower.contains(QStringLiteral("heap")) || lower.contains(QStringLiteral("memóri"))) {
        return QStringLiteral("Nincs elég memória. Csökkentsd a RAM beállítást, vagy zárj be más programokat, majd próbáld újra.");
    }

    // Java runtime.
    if (lower.contains(QStringLiteral("java")) && (lower.contains(QStringLiteral("not found")) ||
        lower.contains(QStringLiteral("configured")) || lower.contains(QStringLiteral("no java")) ||
        lower.contains(QStringLiteral("runtime")))) {
        return QStringLiteral("Nincs beállítva Java futtatókörnyezet. Telepíts Java-t (17+), vagy válaszd ki a Beállításokban.");
    }
    if (lower.contains(QStringLiteral("java version")) || lower.contains(QStringLiteral("parse")) ||
        lower.contains(QStringLiteral("could not query java"))) {
        return QStringLiteral("A Java verziója nem ismerhető fel. Próbálj másik Java-t a Beállításokban.");
    }

    // Authentication / session.
    if (lower.contains(QStringLiteral("session")) || lower.contains(QStringLiteral("authentic"))) {
        return QStringLiteral("Nincs bejelentkezve, vagy lejárt a munkamenet. Jelentkezz be újra.");
    }

    // Missing instance.
    if (lower.contains(QStringLiteral("no instance"))) {
        return QStringLiteral("A játék példánya nincs betöltve. Indítsd újra a launchert.");
    }

    // Directory creation.
    if (lower.contains(QStringLiteral("could not create")) || lower.contains(QStringLiteral("directory"))) {
        return QStringLiteral("Nem sikerült létrehozni a szükséges mappát. Ellenőrizd a jogosultságokat és a szabad helyet a lemezen.");
    }

    // Downloads.
    if (lower.contains(QStringLiteral("already in progress"))) {
        return QStringLiteral("Már fut egy letöltés. Várd meg, amíg befejeződik.");
    }
    if (lower.contains(QStringLiteral("invalid download url")) || lower.contains(QStringLiteral("unsupported url scheme"))) {
        return QStringLiteral("Érvénytelen letöltési link. Frissítsd az alkalmazást vagy a modpackot.");
    }
    if (lower.contains(QStringLiteral("http"))) {
        const QString hm = httpMessage(raw);
        if (!hm.isEmpty())
            return hm;
    }
    if (lower.contains(QStringLiteral("could not open destination")) ||
        lower.contains(QStringLiteral("write failed")) || lower.contains(QStringLiteral("commit failed"))) {
        return QStringLiteral("Nem sikerült a fájl lemezre írása. Ellenőrizd a szabad helyet és a jogosultságokat.");
    }
    if (lower.contains(QStringLiteral("sha-256 mismatch")) || lower.contains(QStringLiteral("mismatch"))) {
        return QStringLiteral("A letöltött fájl sérült (ellenőrzőösszeg-eltérés). Próbáld újra a letöltést.");
    }
    if (lower.contains(QStringLiteral("timeout")) || lower.contains(QStringLiteral("connection")) ||
        lower.contains(QStringLiteral("refused")) || lower.contains(QStringLiteral("network"))) {
        return QStringLiteral("Hálózati hiba a letöltés közben. Ellenőrizd az internetkapcsolatot, majd próbáld újra.");
    }

    // Game process exit.
    {
        const QRegularExpression re(QStringLiteral("exited \\(code (\\d+)\\)"), QRegularExpression::CaseInsensitiveOption);
        const auto m = re.match(raw);
        if (m.hasMatch()) {
            const int code = m.captured(1).toInt();
            if (code == 0)
                return QStringLiteral("A játék normálisan leállt.");
            return QStringLiteral("A játék váratlanul leállt (kód %1). Részletek a naplóban.").arg(code);
        }
    }
    // Fallback: keep a short snippet so the user still has something to search for.
    QString snippet = raw.trimmed();
    if (snippet.length() > 80)
        snippet = snippet.left(80) + QStringLiteral("…");
    return QStringLiteral("Váratlan hiba történt: %1. Részletek a naplóban.").arg(snippet);
}

} // namespace FriendlyError
