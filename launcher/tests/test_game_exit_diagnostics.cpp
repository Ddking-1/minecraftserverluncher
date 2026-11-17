// SPDX-License-Identifier: GPL-3.0-only
// A game that starts and then dies of memory pressure must be explained with
// the same numbers as the RAM panel, while ordinary crashes keep the generic
// message. Everything here is deterministic: no real OOM is provoked.
#include <QCoreApplication>
#include <QString>

#include "core/GameExitDiagnostics.h"
#include "launch/LaunchStep.h"
#include "launch/LaunchTask.h"

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

// Stands in for the real launch step: reports success, then a late failure.
class FakeLaunchStep : public LaunchStep
{
public:
    explicit FakeLaunchStep(LaunchTask *parent) : LaunchStep(parent) {}

    void executeTask() override {}

    void reportStarted() { emit readyForLaunch(); }
    void reportLateFailure(const QString &reason) { emit failed(reason); }
};

} // namespace

int main(int argc, char **argv)
{
    QCoreApplication app(argc, argv);

    // --- classification -----------------------------------------------------
    check(!GameExitDiagnostics::looksLikeMemoryExhaustion(QString(), 0),
          "clean exit is never memory exhaustion");
    check(!GameExitDiagnostics::looksLikeMemoryExhaustion(QString(), 1),
          "silent non-zero exit is a plain crash");
    check(!GameExitDiagnostics::looksLikeMemoryExhaustion(
              QStringLiteral("Exception in thread \"main\" java.lang.NullPointerException"), 1),
          "unrelated JVM exception is not a memory problem");
    check(GameExitDiagnostics::looksLikeMemoryExhaustion(
              QStringLiteral("Exception in thread \"main\" java.lang.OutOfMemoryError: Java heap space"),
              1),
          "heap space OOM detected");
    check(GameExitDiagnostics::looksLikeMemoryExhaustion(
              QStringLiteral("java.lang.OutOfMemoryError: GC overhead limit exceeded"), 1),
          "GC overhead OOM detected");
    check(GameExitDiagnostics::looksLikeMemoryExhaustion(
              QStringLiteral("Error occurred during initialization of VM\n"
                             "Could not reserve enough space for object heap"), 1),
          "native heap reservation failure detected");
    check(GameExitDiagnostics::looksLikeMemoryExhaustion(
              QStringLiteral("There is insufficient memory for the Java Runtime Environment"), 1),
          "JRE memory error detected");
    check(GameExitDiagnostics::looksLikeMemoryExhaustion(
              QStringLiteral("JAVA.LANG.OUTOFMEMORYERROR"), 1),
          "detection is case-insensitive");
    check(GameExitDiagnostics::looksLikeMemoryExhaustion(QString(), 137),
          "OOM-killer exit code 137 detected");
    check(!GameExitDiagnostics::looksLikeMemoryExhaustion(
              QStringLiteral("java.lang.OutOfMemoryError"), 0),
          "exit code 0 wins over log text");

    // --- message ------------------------------------------------------------
    const QString msg = GameExitDiagnostics::memoryHelpMessage(6144, 8192, 6144);
    check(!msg.isEmpty(), "memory message is produced");
    check(msg.contains(QStringLiteral("6144")), "message names the requested memory");
    check(msg.contains(QStringLiteral("8192")), "message names the installed memory");
    check(msg.contains(QStringLiteral("Csökkentsd")), "message tells what to do");
    check(msg.contains(QStringLiteral("zárj be")), "message suggests closing programs");
    check(msg.contains(QStringLiteral("6 GB")), "message shows GB for readability");
    check(msg.contains(QStringLiteral("8 GB")), "message shows installed GB");

    const QString unknownTotal = GameExitDiagnostics::memoryHelpMessage(4096, 0, 65536);
    check(unknownTotal.contains(QStringLiteral("nem állapítható meg")),
          "unknown installed RAM is stated honestly");
    check(unknownTotal.contains(QStringLiteral("65536")), "unknown total still shows the limit");
    check(GameExitDiagnostics::memoryHelpMessage(0, 8192, 6144).isEmpty(),
          "no message without a requested amount");

    check(GameExitDiagnostics::amountText(8192) == QStringLiteral("8 GB (8192 MB)"),
          "whole gigabytes render exactly");
    check(GameExitDiagnostics::amountText(0) == QStringLiteral("?"), "unknown amount is not faked");

    // --- diagnoseExit -------------------------------------------------------
    check(GameExitDiagnostics::diagnoseExit(0, QString(), 6144, 8192, 6144).isEmpty(),
          "clean exit gets no diagnosis");
    check(GameExitDiagnostics::diagnoseExit(
             1, QStringLiteral("Exception in thread \"main\" java.lang.NullPointerException"), 6144,
             8192, 6144)
              .isEmpty(),
          "ordinary crash keeps the generic branch");
    const QString oom = GameExitDiagnostics::diagnoseExit(
        1, QStringLiteral("java.lang.OutOfMemoryError: Java heap space"), 6144, 8192, 6144);
    check(!oom.isEmpty(), "OOM exit gets a diagnosis");
    check(oom.contains(QStringLiteral("6144")) && oom.contains(QStringLiteral("8192")),
          "diagnosis carries the numbers");

    // --- late failure reaches the pipeline ----------------------------------
    {
        LaunchTask task(nullptr);
        auto *step = new FakeLaunchStep(&task);
        task.appendStep(step);

        QString failureReason;
        bool failedSeen = false;
        QObject::connect(&task, &LaunchTask::failed, &task,
                         [&failedSeen, &failureReason](const QString &reason) {
                             failedSeen = true;
                             failureReason = reason;
                         });

        task.start();
        step->reportStarted();
        check(task.state() == LaunchTask::State::Finished, "pipeline finished after start");
        check(!failedSeen, "no failure while the game is running");

        step->reportLateFailure(oom);
        check(failedSeen, "late process failure is surfaced");
        check(failureReason.contains(QStringLiteral("memóriából")),
              "late failure keeps the friendly memory wording");
        check(task.state() == LaunchTask::State::Failed, "state reflects the late failure");
    }

    // An aborted launch must stay quiet: the player stopped it on purpose.
    {
        LaunchTask task(nullptr);
        auto *step = new FakeLaunchStep(&task);
        task.appendStep(step);

        bool failedSeen = false;
        QObject::connect(&task, &LaunchTask::failed, &task,
                         [&failedSeen](const QString &) { failedSeen = true; });

        task.start();
        task.abort();
        step->reportLateFailure(QStringLiteral("killed"));
        check(!failedSeen, "aborted launch reports no failure");
    }

    if (failures == 0)
        std::printf("game_exit_diagnostics: all checks passed\n");
    return failures == 0 ? 0 : 1;
}
