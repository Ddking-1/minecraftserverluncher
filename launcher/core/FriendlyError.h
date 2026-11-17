// SPDX-License-Identifier: GPL-3.0-only
// Turns a technical launch/download error into a short Hungarian sentence a
// basic user understands. Technical detail belongs in the log; this is what the
// UI shows.
#pragma once

#include <QString>

namespace FriendlyError {

// Map a raw (often English/technical) error reason to a plain-language
// Hungarian message. Already-friendly Hungarian messages are passed through
// unchanged, so calling this twice is safe.
QString explain(const QString &raw);

} // namespace FriendlyError
