// SPDX-License-Identifier: GPL-3.0-only
#include "ClaimAuthSession.h"

ClaimAuthSession::ClaimAuthSession(LaunchTask *parent, AuthSessionPtr session)
    : LaunchStep(parent), m_session(std::move(session))
{
}

void ClaimAuthSession::executeTask()
{
    if (!m_session) {
        emit failed(QStringLiteral("no authenticated session available"));
        return;
    }
    if (!m_session->is_online || m_session->access_token.isEmpty()) {
        emit failed(QStringLiteral("session is not authenticated"));
        return;
    }
    emit logLine(QStringLiteral("Claimed backend session for %1").arg(m_session->username));
    emit readyForLaunch();
}
