// SPDX-License-Identifier: GPL-3.0-only
// Standalone verification for the file integrity verifier.
#include <iostream>

#include <QFile>
#include <QTemporaryDir>

#include "core/FileIntegrity.h"

namespace {

// SHA-256 of the empty input, used to validate known-answer hashing.
const char *const kEmptyHash =
    "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855";

bool check(bool cond, const char *what)
{
    if (!cond)
        std::cerr << "FAIL: " << what << std::endl;
    return cond;
}

} // namespace

int main()
{
    using namespace FileIntegrity;

    QTemporaryDir dir;
    if (!check(dir.isValid(), "temporary dir"))
        return 1;

    const QString empty = dir.filePath(QStringLiteral("empty.bin"));
    {
        QFile f(empty);
        if (!check(f.open(QIODevice::WriteOnly), "create empty"))
            return 1;
    }

    if (!check(sha256File(empty) == QString::fromLatin1(kEmptyHash), "empty file hash"))
        return 1;

    const QString data = dir.filePath(QStringLiteral("data.bin"));
    {
        QFile f(data);
        if (!check(f.open(QIODevice::WriteOnly), "create data"))
            return 1;
        f.write("hello");
    }

    const QString hash = sha256File(data);
    if (!check(!hash.isEmpty(), "non-empty hash")
        || !check(verify(data, hash), "verify match")
        || !check(verify(data, hash.toUpper()), "verify case-insensitive")
        || !check(!verify(data, QString::fromLatin1(kEmptyHash)), "verify mismatch"))
        return 1;

    if (!check(sha256File(dir.filePath(QStringLiteral("missing.bin"))).isEmpty(), "missing file empty")
        || !check(!verify(dir.filePath(QStringLiteral("missing.bin")), hash), "missing verify false"))
        return 1;

    std::cout << "OK" << std::endl;
    return 0;
}
