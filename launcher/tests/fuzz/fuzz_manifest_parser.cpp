// SPDX-License-Identifier: GPL-3.0-only
#include <cstdint>
#include <cstddef>
#include <string>

#include <QByteArray>
#include <QJsonDocument>
#include <QJsonObject>
#include <QRegularExpression>

#include "updates/UpdateManifest.h"

// libFuzzer entry point: parse arbitrary bytes as a manifest JSON and ensure
// the parser never crashes on malformed input.
extern "C" int LLVMFuzzerTestOneInput(const uint8_t *data, size_t size)
{
    if (size == 0)
        return 0;

    QByteArray bytes(reinterpret_cast<const char *>(data), static_cast<int>(size));
    QJsonParseError err{};
    const QJsonDocument doc = QJsonDocument::fromJson(bytes, &err);
    if (!doc.isObject())
        return 0;

    UpdateManifest::ParseResult result = UpdateManifest::parse(doc.object());
    (void)result; // keep the call live; result is inspected in dedicated tests

    // Additional sanity: ensure UpdateChecker can tolerate any manifest we
    // successfully parsed (prevents regressions in version comparison paths).
    if (result.ok && !result.component.name.isEmpty() && !result.component.version.isEmpty()) {
        UpdateChecker::Decision decision = UpdateChecker::check(QStringLiteral("0.0.0"), result.component);
        (void)decision;
    }

    return 0;
}
