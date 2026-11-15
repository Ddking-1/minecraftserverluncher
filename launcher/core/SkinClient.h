// SPDX-License-Identifier: GPL-3.0-only
// Project-owned launcher core. Skin upload/download client for the Go backend.
#pragma once

#include <QObject>
#include <QString>

#include <functional>

class BackendAuthClient;

// Async skin client backed by QNetworkAccessManager.
//
// Designed to be testable: callbacks are the primary integration point, and the
// implementation cleans up its QNetworkReply/QTimer objects so the unit tests
// can exercise upload/download against a local QTcpServer mock.
class SkinClient : public QObject
{
    Q_OBJECT
public:
    // The two dimensions accepted by the backend (Minecraft skin standard).
    static bool isSupportedSize(int width, int height);

    explicit SkinClient(BackendAuthClient *client, QObject *parent = nullptr);

    // Upload a local PNG skin file to /api/v1/skin-files.
    // `filePath` is the absolute path to the PNG on disk.
    // The callback receives (success, skinId, error). On success the skinId is
    // the backend-assigned identifier; on failure it is empty and error contains
    // a human-readable reason.
    void uploadSkin(const QString &filePath,
                    std::function<void(bool success, QString skinId,
                                       QString error)> callback);

    // Download a 64x64 avatar PNG for the given skinId from
    // /api/v1/skin-files/{id}/avatar and convert it to a QPixmap. The callback
    // receives (success, avatar).
    void downloadAvatar(const QString &skinId,
                        std::function<void(bool success, QPixmap avatar)> callback);

private:
    BackendAuthClient *m_client = nullptr;

    static void invokeCallback(std::function<void(bool, QString, QString)> cb,
                               bool success, QString skinId, QString error);
    static void invokeCallback(std::function<void(bool, QPixmap)> cb,
                               bool success, QPixmap avatar);
};
