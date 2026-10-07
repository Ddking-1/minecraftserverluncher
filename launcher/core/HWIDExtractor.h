#pragma once: GPL-3.0-only

#include <string>
#include <string_view>
#include <windows.h>

class HWIDExtractor {
public:
    // Returns a SHA-256 hardware identifier derived from SMBIOS UUID + CPUID +
    // primary MAC address. [[nodiscard]] because ignoring the result is almost
    // always a bug.
    [[nodiscard]]
    static std::string getHardwareUUID();

private:
    [[nodiscard]]
    static std::string getSMBIOSUUID();
    [[nodiscard]]
    static std::string getCPUId();
    [[nodiscard]]
    static std::string getPrimaryMACAddress();
    [[nodiscard]]
    static std::string generateSHA256(std::string_view input);
};
