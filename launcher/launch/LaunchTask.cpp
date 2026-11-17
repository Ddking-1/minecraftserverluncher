// SPDX-License-Identifier: GPL-3.0-only
#include "LaunchTask.h"

#include "LaunchStep.h"
#include "core/BaseInstance.h"
#include "core/FriendlyError.h"

LaunchTask::LaunchTask(BaseInstance *instance, QObject *parent)
    : QObject(parent), m_instance(instance)
{
}

LaunchTask::~LaunchTask() = default;

void LaunchTask::appendStep(LaunchStep *step)
{
    if (m_state != State::NotStarted)
        return;
    step->setParent(this);
    connect(step, &LaunchStep::readyForLaunch, this, &LaunchTask::onStepReady);
    connect(step, &LaunchStep::failed, this, &LaunchTask::onStepFailed);
    connect(step, &LaunchStep::logLine, this, &LaunchTask::logLine);
    m_steps.push_back(step);
}

void LaunchTask::start()
{
    if (m_state != State::NotStarted)
        return;
    setState(State::Running);
    runNextStep();
}

void LaunchTask::abort()
{
    if (m_state == State::Finished || m_state == State::Aborted || m_state == State::Failed)
        return;
    if (m_currentStep >= 0 && m_currentStep < static_cast<int>(m_steps.size()))
        m_steps.at(m_currentStep)->abort();
    setState(State::Aborted);
}

LaunchTask::State LaunchTask::state() const
{
    return m_state;
}

BaseInstance *LaunchTask::instance() const
{
    return m_instance;
}

void LaunchTask::runNextStep()
{
    ++m_currentStep;
    if (m_currentStep >= static_cast<int>(m_steps.size())) {
        setState(State::Finished);
        emit succeeded();
        return;
    }
    m_steps.at(m_currentStep)->executeTask();
}

void LaunchTask::onStepReady()
{
    if (m_state != State::Running)
        return;
    runNextStep();
}

void LaunchTask::onStepFailed(const QString &reason)
{
    // A step may also report a problem *after* the pipeline finished: the game
    // process started successfully and then died (for example out of memory).
    // That is a real failure and must reach the UI instead of being dropped.
    if (m_state == State::Failed || m_state == State::Aborted)
        return;
    // Keep the technical detail in the log, but show a plain-language message.
    emit logLine(reason);
    setState(State::Failed);
    emit failed(FriendlyError::explain(reason));
}

void LaunchTask::setState(State state)
{
    if (m_state == state)
        return;
    m_state = state;
    emit stateChanged(state);
}