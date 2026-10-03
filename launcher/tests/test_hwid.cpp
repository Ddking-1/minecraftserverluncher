// SPDX-License-Identifier: GPL-3.0-only
// Standalone verification for HWIDExtractor::getHardwareUUID().
#include <cctype>
#include <iostream>
#include <string>

#include "core/HWIDExtractor.h"

int main()
{
    const std::string id = HWIDExtractor::getHardwareUUID();
    std::cout << "HWID: " << id << std::endl;

    if (id.size() != 64) {
        std::cerr << "FAIL: expected 64 hex characters, got " << id.size() << std::endl;
        return 1;
    }
    for (char c : id) {
        if (!std::isxdigit(static_cast<unsigned char>(c))) {
            std::cerr << "FAIL: non-hex character in HWID" << std::endl;
            return 1;
        }
    }
    std::cout << "OK" << std::endl;
    return 0;
}
