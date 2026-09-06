// SPDX-License-Identifier: GPL-3.0-only
#include "LaunchStep.h"

#include "LaunchTask.h"

LaunchStep::LaunchStep(LaunchTask *parent) : QObject(parent), m_parent(parent)
{
}

LaunchStep::~LaunchStep() = default;
