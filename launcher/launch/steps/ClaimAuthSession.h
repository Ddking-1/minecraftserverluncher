// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher core. Launch step that claims the backend session.
#pragma once

#include <QString>

#include "launch/LaunchStep.h"
#include "core/InstanceTypes.h"

// Verifies a valid backend-authenticated session is available for the launch.
// Replaces Prism's ClaimAccount (no Mojang/MSA/Xbox checks).
class ClaimAuthSession : public LaunchStep
{
    Q_OBJECT
public:
    ClaimAuthSession(LaunchTask *parent, AuthSessionPtr session);

    void executeTask() override;

private:
    AuthSessionPtr m_session;
};
