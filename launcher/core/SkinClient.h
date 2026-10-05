// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher core. Client-side skin upload/preview logic.
#pragma once

#include <QObject>
#include <QString>
#include <QByteArray>

class BackendAuthClient;

// SkinClient uploads a local skin file to the backend and reports the stored
// skin URL. It validates the image dimensions locally before uploading so the
// user gets immediate feedback. It does not render anything itself; the UI
// layer binds to its signals.
class SkinClient : public QObject
{
    Q_OBJECT
public:
    // The two dimensions accepted by the backend (Minecraft skin standard).
    static bool isSupportedSize(int width, int height);

    explicit SkinClient(BackendAuthClient *client, QObject *parent = nullptr);

    // Read, validate and upload a skin file for the given variant.
    // Emits uploadSucceeded/uploadFailed.
    void uploadSkin(const QString &accessToken, const QString &filePath,
                    const QString &variant = QStringLiteral("default"));

signals:
    void uploadSucceeded(const QString &variant);
    void uploadFailed(const QString &reason);

private:
    BackendAuthClient *m_client;
    bool m_busy = false;
};
