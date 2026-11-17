// SPDX-License-Identifier: GPL-3.0-only
#include "GameExitDiagnostics.h"

#include <QLocale>
#include <QStringList>

namespace {

// The JVM's own wording. All matches are case-insensitive and are only trusted
// for a process that actually failed (exitCode != 0).
const QStringList &memoryMarkers()
{
    // NB: "signals" is a Qt keyword macro, so the list is called markers.
    static const QStringList markers = {
        QStringLiteral("outofmemoryerror"),
        QStringLiteral("java heap space"),
        QStringLiteral("gc overhead limit exceeded"),
        QStringLiteral("could not reserve enough space for object heap"),
        QStringLiteral("insufficient memory for the java runtime environment"),
        QStringLiteral("cannot allocate memory"),
        QStringLiteral("native memory allocation"),
        QStringLiteral("killed process"),        // Linux OOM killer
        QStringLiteral("out of memory"),
    };
    return markers;
}

// 128 + SIGKILL(9): what the Linux OOM killer leaves behind.
constexpr int kExitKilled = 137;

} // namespace

namespace GameExitDiagnostics {

bool looksLikeMemoryExhaustion(const QString &output, int exitCode)
{
    if (exitCode == 0)
        return false;
    if (exitCode == kExitKilled)
        return true;

    const QString lower = output.toLower();
    for (const QString &marker : memoryMarkers()) {
        if (lower.contains(marker))
            return true;
    }
    return false;
}

QString amountText(int mb)
{
    if (mb <= 0)
        return QStringLiteral("?");
    if (mb % 1024 == 0)
        return QStringLiteral("%1 GB (%2 MB)").arg(mb / 1024).arg(mb);
    return QStringLiteral("%1 GB (%2 MB)")
        .arg(QLocale().toString(mb / 1024.0, 'f', 1))
        .arg(mb);
}

QString memoryHelpMessage(int requestedMb, int totalMb, int allowedMb)
{
    if (requestedMb <= 0)
        return QString();

    const QString requested = amountText(requestedMb);
    const QString allowed = amountText(allowedMb);

    if (totalMb > 0) {
        return QStringLiteral(
                   "A játék elfogyott a memóriából és leállt. Kért memória: %1 · a gépben: %2 · "
                   "ebből engedélyezett maximum: %3. Csökkentsd a RAM csúszkát, vagy zárj be más "
                   "programokat, majd indítsd újra a játékot.")
            .arg(requested, amountText(totalMb), allowed);
    }

    return QStringLiteral(
               "A játék elfogyott a memóriából és leállt. Kért memória: %1 · a gép memóriája nem "
               "állapítható meg · engedélyezett maximum: %2. Csökkentsd a RAM csúszkát, vagy zárj "
               "be más programokat, majd indítsd újra a játékot.")
        .arg(requested, allowed);
}

QString diagnoseExit(int exitCode, const QString &output, int requestedMb, int totalMb,
                     int allowedMb)
{
    if (!looksLikeMemoryExhaustion(output, exitCode))
        return QString();
    return memoryHelpMessage(requestedMb, totalMb, allowedMb);
}

} // namespace GameExitDiagnostics
