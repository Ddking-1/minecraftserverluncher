// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher core. Launch step that resolves game libraries.
#pragma once

#include "launch/LaunchStep.h"

class MinecraftInstance;

// Ensures the instance's library directory exists and is ready for the launch.
// NOTE: actual library download depends on the metadata source decision
// (docs/DECISIONS.md DECISION-005, currently pending); this step prepares the
// local layout and is the seam where download logic will be added.
class ResolveLibraries : public LaunchStep
{
    Q_OBJECT
public:
    ResolveLibraries(LaunchTask *parent, MinecraftInstance *instance);

    void executeTask() override;

private:
    MinecraftInstance *m_instance;
};
