// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher core. Launch step that prepares native libraries.
#pragma once

#include "launch/LaunchStep.h"

class MinecraftInstance;

// Ensures the natives directory exists and is clean for the launch.
// NOTE: actual native extraction from downloaded libraries depends on the
// metadata source decision (docs/DECISIONS.md DECISION-005, currently pending).
class ExtractNatives : public LaunchStep
{
    Q_OBJECT
public:
    ExtractNatives(LaunchTask *parent, MinecraftInstance *instance);

    void executeTask() override;

private:
    MinecraftInstance *m_instance;
};
