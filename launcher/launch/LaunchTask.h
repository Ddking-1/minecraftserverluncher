// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher core. Orchestrates a sequence of LaunchSteps.
#pragma once

#include <QObject>
#include <QString>
#include <vector>
#include <memory>

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

    // Append a step to the pipeline. Must be called before start().
    void appendStep(std::unique_ptr<LaunchStep> step);

    // Begin executing the pipeline.
    void start();

    // Abort the running pipeline.
    void abort();

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
    std::vector<std::unique_ptr<LaunchStep>> m_steps;
    int m_currentStep = -1;
    State m_state = State::NotStarted;
};