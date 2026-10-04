// SPDX-License-Identifier: GPL-3.0-only
// Standalone verification for the update checker.
#include <iostream>

#include "updates/UpdateChecker.h"

namespace {

bool check(bool cond, const char *what)
{
    if (!cond)
        std::cerr << "FAIL: " << what << std::endl;
    return cond;
}

UpdateManifest::ManifestComponent component(const QString &version, const QString &minLauncher)
{
    UpdateManifest::ManifestComponent c;
    c.name = QStringLiteral("launcher");
    c.version = version;
    c.minLauncher = minLauncher;
    return c;
}

} // namespace

int main()
{
    using namespace UpdateChecker;

    if (!check(compareVersions(QStringLiteral("1.2.10"), QStringLiteral("1.2.9")) == 1, "1.2.10 > 1.2.9")
        || !check(compareVersions(QStringLiteral("1.2"), QStringLiteral("1.2.0")) == 0, "trailing zero equal")
        || !check(compareVersions(QStringLiteral("0.9"), QStringLiteral("1.0")) == -1, "0.9 < 1.0")
        || !check(compareVersions(QStringLiteral("2"), QStringLiteral("1.9.9")) == 1, "2 > 1.9.9"))
        return 1;

    if (!check(check(QStringLiteral("0.1.0"), component(QStringLiteral("0.2.0"), QStringLiteral("0.1.0")))
                   == Decision::UpdateAvailable,
               "newer available"))
        return 1;

    if (!check(check(QStringLiteral("0.2.0"), component(QStringLiteral("0.2.0"), QStringLiteral("0.1.0")))
                   == Decision::UpToDate,
               "same version up to date"))
        return 1;

    if (!check(check(QStringLiteral("0.1.0"), component(QStringLiteral("0.3.0"), QStringLiteral("0.2.0")))
                   == Decision::Incompatible,
               "below min launcher incompatible"))
        return 1;

    if (!check(check(QStringLiteral("0.1.0"), component(QStringLiteral("0.1.0"), QString()))
                   == Decision::UpToDate,
               "empty min launcher up to date"))
        return 1;

    std::cout << "OK" << std::endl;
    return 0;
}
