// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher core. Orchestrates a sequence of LaunchSteps.
#pragma once

#include <QObject>
#include <QString>
#include <vector>

class LaunchStep;
class BaseInstance;

// Runs a launch pipeline: executes each LaunchStep in order, waiting for each
// to signal readyForLaunch() before starting the next. Emits state/log signals
// for the UI.
class LaunchTask : public QObject
{
    Q_OBJECT
public:
    enum class State
    {
        NotStarted,
        Running,
        Failed,
        Aborted,
        Finished,
    };

    explicit LaunchTask(BaseInstance *instance, QObject *parent = nullptr);
    ~LaunchTask() override;

    // Append a step to the pipeline and take ownership of it. Must be called
    // before start(). The step is reparented to this task (single ownership).
    void appendStep(LaunchStep *step);

    // Begin executing the pipeline.
    void start();

    // Abort the running pipeline.
    void abort();

    // Note on signal order: succeeded() means "the game process started". If
    // that process later dies (out of memory or a crash), failed() follows even
    // though the state was already Finished, and state() becomes Failed.

    State state() const;
    BaseInstance *instance() const;

signals:
    void logLine(const QString &line);
    void stateChanged(LaunchTask::State state);
    void succeeded();
    void failed(const QString &reason);

private:
    void runNextStep();
    void onStepReady();
    void onStepFailed(const QString &reason);
    void setState(State state);

    BaseInstance *m_instance;
    std::vector<LaunchStep *> m_steps;
    int m_currentStep = -1;
    State m_state = State::NotStarted;
};