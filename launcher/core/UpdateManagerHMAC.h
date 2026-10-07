// SPDX-License-Identifier: GPL-3.0-only
// HMAC-SHA256 verification utilities for launcher update manifest validation.
// Mirrors the canonical string form used by the Go backend:
//   version|url|sha256|mandatory(1/0)
#pragma once

#include <QCryptographicHash>
#include <QString>
#include <QByteArray>

// Compute HMAC-SHA256 over canonical manifest string.
// key: signing key, version, url, sha256: manifest fields, mandatory: 1 or 0.
[[nodiscard]]
QString hmacManifestSignature(const QString &key,
                             const QString &version,
                             const QString &url,
                             const QString &sha256,
                             bool mandatory);

// Constant-time-ish comparison of two hex signature strings.
[[nodiscard]]
bool hmacSignaturesEqual(const QString &a, const QString &b);

// Verify a manifest signature. Returns true only if signature matches the
// expected HMAC over the canonical fields using the given key.
[[nodiscard]]
bool verifyManifestSignature(const QString &signingKey,
                             const QString &version,
                             const QString &url,
                             const QString &sha256,
                             bool mandatory,
                             const QString &signature);
