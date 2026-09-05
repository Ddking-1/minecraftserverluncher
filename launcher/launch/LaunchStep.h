// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher core. Abstract base for a single launch step.
#pragma once

#include <QObject>
#include <QString>

class LaunchTask;

// One stage of a launch pipeline. A step performs its work in executeTask()
// and signals readyForLaunch() when done so LaunchTask can proceed to the next
// step. Steps are re-entrant and abortable.
class LaunchStep : public QObject
{
    Q_OBJECT
public:
    explicit LaunchStep(LaunchTask *parent);
    ~LaunchStep() override;

    // Perform the step's work. Must eventually emit readyForLaunch() (or
    // emitFailed()) exactly once.
    virtual void executeTask() = 0;

    // Called by LaunchTask when the step is aborted.
    virtual void abort() {}

    // Called by LaunchTask when the step has finished and the launch ended.
    virtual void finalize() {}

signals:
    // Append a line to the launch log.
    void logLine(const QString &line);
    // Emitted when the step has completed its work.
    void readyForLaunch();
    // Emitted when the step failed; the launch is aborted.
    void failed(const QString &reason);

protected:
    LaunchTask *m_parent;
};
