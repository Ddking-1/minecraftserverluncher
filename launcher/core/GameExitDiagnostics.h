// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher core. Turns a dead game process into a diagnosis the
// user can act on, without inventing a second error vocabulary: the memory
// wording reuses the numbers shown by the RAM panel (SystemMemory).
#pragma once

#include <QString>

namespace GameExitDiagnostics {

// True when the captured process output looks like the JVM ran out of memory
// (heap exhaustion, native allocation failure, GC overhead) or the OS killed
// the process for RAM pressure (SIGKILL / exit code 137 from the OOM killer).
// A clean exit (0) is never a memory problem.
[[nodiscard]] bool looksLikeMemoryExhaustion(const QString &output, int exitCode);

// "6 GB (6144 MB)" style amount, so the message stays readable and exact.
[[nodiscard]] QString amountText(int mb);

// Friendly Hungarian explanation with the numbers the user needs: requested
// heap, installed RAM and the allowed maximum. Returns a non-empty string only
// when the numbers make sense (requested > 0).
[[nodiscard]] QString memoryHelpMessage(int requestedMb, int totalMb, int allowedMb);

// The message to report for this exit, or an empty string when the exit is not
// a memory problem (the caller then keeps its generic crash handling).
[[nodiscard]] QString diagnoseExit(int exitCode, const QString &output, int requestedMb,
                                   int totalMb, int allowedMb);

} // namespace GameExitDiagnostics
