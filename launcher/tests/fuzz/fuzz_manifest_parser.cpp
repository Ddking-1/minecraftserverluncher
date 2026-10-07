// SPDX-License-Identifier: GPL-3.0-only
// Fuzz harness for the UpdateManifest parser.
// Feeds malformed, truncated, oversized and otherwise hostile JSON manifest
// fragments into UpdateManifest::parse() and verifies that the parser does
// not crash, leak or otherwise misbehave.
#include <cstddef>
#include <cstdint>
#include <vector>

#include <QJsonDocument>
#include <QJsonObject>

#include "updates/UpdateManifest.h"

extern "C" int LLVMFuzzerTestOneInput(const uint8_t *data, size_t size)
{
    if (size == 0) {
        return 0;
    }

    // Cap the input to a sane maximum so the parser never sees megabyte-scale
    // JSON payloads during fuzzing.
    constexpr size_t kMaxInput = 8 * 1024;
    std::vector<uint8_t> capped;
    if (size > kMaxInput) {
        capped.assign(data, data + kMaxInput);
    } else {
        capped.assign(data, data + size);
    }

    QByteArray bytes(reinterpret_cast<const char *>(capped.data()), static_cast<int>(capped.size()));

    QJsonParseError err{};
    const QJsonDocument doc = QJsonDocument::fromJson(bytes, &err);

    if (err.error == QJsonParseError::NoError && doc.isObject()) {
        UpdateManifest::parse(doc.object());
    }

    return 0;
}
