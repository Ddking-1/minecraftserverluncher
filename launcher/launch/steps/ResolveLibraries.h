// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher core. Launch step that resolves game libraries.
#pragma once

#include "launch/LaunchStep.h"

class MinecraftInstance;

// Resolves the instance's library files from the instance `version.json` metadata
// (DECISION-005: the self-hosted backend is the metadata source). Every
// download is SHA-1 verified before an atomic write and any metadata, path or
// hash mismatch fails the launch instead of proceeding.
class ResolveLibraries : public LaunchStep
{
    Q_OBJECT
public:
    ResolveLibraries(LaunchTask *parent, MinecraftInstance *instance);

    void executeTask() override;

private:
    MinecraftInstance *m_instance;
};
